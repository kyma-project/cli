package add

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/remove/managed"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/communitymodules"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yaml "sigs.k8s.io/yaml/goyaml.v3"
)

type addConfig struct {
	*cmdcommon.KymaConfig
	cmdcommon.KubeClientConfig

	wantedModules []string
	//custom        string
}

func NewAddCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := addConfig{
		KymaConfig:       kymaConfig,
		KubeClientConfig: cmdcommon.KubeClientConfig{},
	}

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Adds Kyma modules.",
		Long:  `Use this command to add Kyma modules`,
		PreRun: func(_ *cobra.Command, _ []string) {
			clierror.Check(cfg.KubeClientConfig.Complete())
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runAdd(&cfg))
		},
	}

	cmd.AddCommand(managed.NewManagedCMD(kymaConfig))

	cfg.KubeClientConfig.AddFlag(cmd)
	cmd.Flags().StringSliceVar(&cfg.wantedModules, "module", []string{}, "Name and version of the modules to add. Example: --module serverless,keda:1.1.1,etc...")
	//cmd.Flags().StringVar(&cfg.custom, "custom", "", "Path to the custom file")

	return cmd
}

func runAdd(cfg *addConfig) clierror.Error {
	err := assureNamespace("kyma-system", cfg)
	if err != nil {
		return err
	}

	return applySpecifiedModules(cfg)
}

func applySpecifiedModules(cfg *addConfig) clierror.Error {
	modules, err := communitymodules.GetAvailableModules()
	if err != nil {
		return err
	}
	for _, rec := range modules {
		if !containsModule(rec.Name, cfg.wantedModules) {
			continue
		}

		fmt.Printf("Found matching module for %s\n", rec.Name)
		latestVersion := communitymodules.GetLatestVersion(rec.Versions)

		err = applyGivenObjects(cfg, latestVersion.DeploymentYaml)
		if err != nil {
			return err
		}
		err = applyGivenObjects(cfg, latestVersion.CrYaml)
		if err != nil {
			return err
		}
	}

	return nil
}

func applyGivenObjects(cfg *addConfig, url string) clierror.Error {
	givenYaml, err := http.Get(url)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to get YAML from URL"))
	}
	defer givenYaml.Body.Close()

	yamlContent, err := io.ReadAll(givenYaml.Body)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to read YAML"))
	}

	objects, err := decodeYaml(bytes.NewReader(yamlContent))
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to decode YAML"))
	}

	err = cfg.KubeClient.RootlessDynamic().ApplyMany(cfg.Ctx, objects)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to apply module resources"))

	}
	return nil
}

//func applyCustomConfiguration(cfg *addConfig) clierror.Error {
//	fmt.Println("Applying custom configuration from " + cfg.custom)
//
//	customYaml, err := os.ReadFile(cfg.custom)
//	if err != nil {
//		return clierror.Wrap(err, clierror.New("failed to read custom file"))
//	}
//
//	objects, err := decodeYaml(bytes.NewReader(customYaml))
//	if err != nil {
//		return clierror.Wrap(err, clierror.New("failed to decode YAML"))
//	}
//
//	err = cfg.KubeClient.RootlessDynamic().ApplyMany(cfg.Ctx, objects)
//	if err != nil {
//		return clierror.Wrap(err, clierror.New("failed to apply module resources"))
//	}
//
//	return nil
//}

func assureNamespace(namespace string, cfg *addConfig) clierror.Error {
	_, err := cfg.KubeClientConfig.KubeClient.Static().CoreV1().Namespaces().Get(cfg.Ctx, namespace, metav1.GetOptions{})
	if !errors.IsNotFound(err) {
		return nil
	}
	_, err = cfg.KubeClientConfig.KubeClient.Static().CoreV1().Namespaces().Create(cfg.Ctx, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return clierror.New("failed to create namespace")
	}
	return nil
}

func containsModule(have string, want []string) bool {
	for _, rec := range want {
		if rec == have {
			return true
		}
	}
	return false
}

func decodeYaml(r io.Reader) ([]unstructured.Unstructured, error) {
	results := make([]unstructured.Unstructured, 0)
	decoder := yaml.NewDecoder(r)

	for {
		var obj map[string]interface{}
		err := decoder.Decode(&obj)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		u := unstructured.Unstructured{Object: obj}
		if u.GetObjectKind().GroupVersionKind().Kind == "CustomResourceDefinition" {
			results = append([]unstructured.Unstructured{u}, results...)
			continue
		}
		results = append(results, u)
	}

	return results, nil
}
