package modules

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/model"
	"io"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net/http"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/spf13/cobra"
)

type modulesConfig struct {
	*cmdcommon.KymaConfig
	cmdcommon.KubeClientConfig

	catalog   bool
	managed   bool
	installed bool
}

const URL = "https://raw.githubusercontent.com/kyma-project/community-modules/main/model.json"

func NewModulesCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := modulesConfig{
		KymaConfig:       kymaConfig,
		KubeClientConfig: cmdcommon.KubeClientConfig{},
	}

	cmd := &cobra.Command{
		Use:   "modules",
		Short: "List modules.",
		Long:  `List either installed, managed or available Kyma modules.`,
		PreRun: func(_ *cobra.Command, args []string) {
			clierror.Check(cfg.KubeClientConfig.Complete())
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(listModules(&cfg))
		},
	}
	cfg.KubeClientConfig.AddFlag(cmd)

	cmd.Flags().BoolVar(&cfg.catalog, "catalog", false, "List of al available Kyma modules.")
	cmd.Flags().BoolVar(&cfg.managed, "managed", false, "List of all Kyma modules managed by central control-plane.")
	cmd.Flags().BoolVar(&cfg.installed, "installed", false, "List of all currently installed Kyma modules.")

	cmd.MarkFlagsOneRequired("catalog", "managed", "installed")
	cmd.MarkFlagsMutuallyExclusive("catalog", "managed")
	cmd.MarkFlagsMutuallyExclusive("catalog", "installed")
	cmd.MarkFlagsMutuallyExclusive("managed", "installed")

	return cmd
}

func listModules(cfg *modulesConfig) clierror.Error {
	var err clierror.Error

	if cfg.catalog {
		err = listAllModules()
		if err != nil {
			return clierror.WrapE(err, clierror.New("failed to list all Kyma modules"))
		}
		return nil
	}

	if cfg.managed {
		err = listManagedModules(cfg)
		if err != nil {
			return clierror.WrapE(err, clierror.New("failed to list managed Kyma modules"))
		}
		return nil
	}

	if cfg.installed {
		err = listInstalledModules(cfg)
		if err != nil {
			return clierror.WrapE(err, clierror.New("failed to list installed Kyma modules"))
		}
		return nil
	}

	return clierror.WrapE(err, clierror.New("failed to get modules", "please use one of: catalog, managed or installed flags"))
}

func listInstalledModules(cfg *modulesConfig) clierror.Error {
	installed, err := getInstalledModules(cfg)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get installed Kyma modules"))
	}
	fmt.Println("Installed modules:\n")
	for _, rec := range installed {
		fmt.Println(rec)
	}
	return nil
}

func listManagedModules(cfg *modulesConfig) clierror.Error {
	managed, err := getManagedModules(cfg)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get managed Kyma modules"))
	}
	fmt.Println("Managed modules:\n")
	for _, rec := range managed {
		fmt.Println(rec)
	}
	return nil
}

func listAllModules() clierror.Error {
	modules, err := getAllModules()
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get all Kyma modules"))
	}
	fmt.Println("Available modules:\n")
	for _, rec := range modules {
		fmt.Println(rec)
	}
	return nil
}

func getAllModules() ([]string, clierror.Error) {
	resp, err := http.Get(URL)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while getting modules list from github"))
	}
	defer resp.Body.Close()

	var template model.Module

	template, respErr := handleResponse(err, resp, template)
	if respErr != nil {
		return nil, clierror.WrapE(respErr, clierror.New("while handling response"))
	}

	var out []string
	for _, rec := range template {
		out = append(out, rec.Name)
	}
	return out, nil
}

func handleResponse(err error, resp *http.Response, template model.Module) (model.Module, clierror.Error) {
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while reading http response"))
	}
	err = json.Unmarshal(bodyText, &template)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while unmarshalling"))
	}
	return template, nil
}

func getManagedModules(cfg *modulesConfig) ([]string, clierror.Error) {
	GVRKyma := schema.GroupVersionResource{
		Group:    "operator.kyma-project.io",
		Version:  "v1beta2",
		Resource: "kymas",
	}

	unstruct, err := cfg.KubeClient.Dynamic().Resource(GVRKyma).Namespace("kyma-system").Get(cfg.Ctx, "default", metav1.GetOptions{})
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while getting Kyma CR"))
	}

	moduleNames, err := getModuleNames(unstruct)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while getting module names from CR"))
	}

	return moduleNames, nil
}

func getInstalledModules(cfg *modulesConfig) ([]string, clierror.Error) {
	resp, err := http.Get(URL)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while getting modules list from github"))
	}
	defer resp.Body.Close()

	var template model.Module

	template, respErr := handleResponse(err, resp, template)
	if respErr != nil {
		return nil, clierror.WrapE(respErr, clierror.New("while handling response"))
	}

	var out []string
	for _, rec := range template {
		managerPath := strings.Split(rec.Versions[0].ManagerPath, "/")
		managerName := managerPath[len(managerPath)-1]
		version := rec.Versions[0].Version
		deployment, err := cfg.KubeClient.Static().AppsV1().Deployments("kyma-system").Get(cfg.Ctx, managerName, metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			msg := "while getting the " + managerName + " deployment"
			return nil, clierror.Wrap(err, clierror.New(msg))
		}
		if !errors.IsNotFound(err) {
			deploymentImage := strings.Split(deployment.Spec.Template.Spec.Containers[0].Image, "/")
			installedVersion := strings.Split(deploymentImage[len(deploymentImage)-1], ":")

			if version == installedVersion[len(installedVersion)-1] {
				out = append(out, rec.Name+" - "+installedVersion[len(installedVersion)-1])
			} else {
				out = append(out, rec.Name+" - "+"outdated version, latest version is "+version)
			}
		}
	}
	return out, nil
}

func getModuleNames(unstruct *unstructured.Unstructured) ([]string, error) {
	var moduleNames []string
	managedFields := unstruct.GetManagedFields()
	for _, field := range managedFields {
		var data map[string]interface{}
		err := json.Unmarshal(field.FieldsV1.Raw, &data)
		if err != nil {
			return nil, err
		}

		spec, ok := data["f:spec"].(map[string]interface{})
		if !ok {
			continue
		}

		modules, ok := spec["f:modules"].(map[string]interface{})
		if !ok {
			continue
		}

		for key := range modules {
			if strings.Contains(key, "name") {
				name := strings.TrimPrefix(key, "k:{\"name\":\"")
				name = strings.Trim(name, "\"}")
				moduleNames = append(moduleNames, name)
			}
		}
	}
	return moduleNames, nil
}
