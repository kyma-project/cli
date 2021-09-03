package deploy

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"strings"

	"github.com/kyma-incubator/reconciler/pkg/cluster"
	"github.com/kyma-incubator/reconciler/pkg/keb"
	"github.com/kyma-incubator/reconciler/pkg/reconciler"
	"github.com/kyma-incubator/reconciler/pkg/reconciler/service"
	"github.com/kyma-incubator/reconciler/pkg/scheduler"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/download"
	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/overrides"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	_ "github.com/kyma-incubator/reconciler/pkg/reconciler/instances"
)

type command struct {
	cli.Command
	opts *Options
}

//NewCmd creates a new deploy command
func NewCmd(o *Options) *cobra.Command {

	cmd := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:     "deploy",
		Short:   "Deploys Kyma on a running Kubernetes cluster.",
		Long:    "Use this command to deploy, upgrade, or adapt Kyma on a running Kubernetes cluster.",
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run(cmd.opts) },
		Aliases: []string{"d"},
	}
	return cobraCmd
}

func componentsFromStrings(list []string, overrides map[string]string) []keb.Components {
	var components []keb.Components
	for _, item := range list {
		s := strings.Split(item, "@")

		component := keb.Components{Component: s[0], Namespace: s[1]}

		for k, v := range overrides {
			overrideComponent := strings.Split(k, ".")[0]
			if s[0] == overrideComponent {
				component.Configuration = append(component.Configuration, keb.Configuration{Key: k, Value: v})
			}
		}
		components = append(components, component)
	}
	return components
}

func defaultComponentList(overrides map[string]string) []keb.Components {
	defaultComponents := []string{
		"cluster-essentials@kyma-system",
		"istio-configuration@istio-system",
		"certificates@istio-system",
		"logging@kyma-system",
		"tracing@kyma-system",
		"kiali@kyma-system",
		"monitoring@kyma-system",
		"eventing@kyma-system",
		"ory@kyma-system",
		"api-gateway@kyma-system",
		"service-catalog@kyma-system",
		"service-catalog-addons@kyma-system",
		"rafter@kyma-system",
		"helm-broker@kyma-system",
		"cluster-users@kyma-system",
		"serverless@kyma-system",
		"application-connector@kyma-integration"}
	return componentsFromStrings(defaultComponents, overrides)
}

func (cmd *command) Run(o *Options) error {

	var err error

	//start := time.Now()

	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}
	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	overridesBuilder := &overrides.Builder{}

	kymaHome, err := files.KymaHome()
	if err != nil {
		return errors.Wrap(err, "Could not find or create Kyma home directory")
	}

	valuesPath, err := download.GetFile("https://raw.githubusercontent.com/kyma-project/kyma/main/installation/resources/values.yaml", kymaHome)

	overridesStep := cmd.NewStep("Applying Kyma2 overrides")
	if err = overridesBuilder.AddFile(valuesPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			overridesStep.LogInfof("Kyma2 override path not found but continuing: %s", err)
		} else {
			return errors.Wrap(err, "Could not add overrides for Kyma 2.0")
		}
	}
	overridesStep.Success()

	kubecfgFile := kube.KubeconfigPath(cmd.KubeconfigPath)

	_, err = service.NewComponentReconciler("base")
	if err != nil {
		return err
	}

	overridesBuilder.AddInterceptor([]string{"global.domainName", "global.ingress.domainName"}, overrides.NewDomainNameOverrideInterceptor(cmd.K8s.Static()))
	overridesBuilder.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, overrides.NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", cmd.K8s.Static()))
	overridesBuilder.AddInterceptor([]string{"serverless.dockerRegistry.internalServerAddress", "serverless.dockerRegistry.serverAddress", "serverless.dockerRegistry.registryAddress"}, overrides.NewRegistryInterceptor(cmd.K8s.Static()))
	overridesBuilder.AddInterceptor([]string{"serverless.dockerRegistry.enableInternal"}, overrides.NewRegistryDisableInterceptor(cmd.K8s.Static()))

	isK3d, err := overrides.IsK3dCluster(cmd.K8s.Static())
	if err != nil {
		return err
	}

	if _, err := overrides.PatchCoreDNS(cmd.K8s.Static(), overridesBuilder, isK3d); err != nil {
		return err
	}

	nestedOverrides, err := overridesBuilder.Build()
	if err != nil {
		return err
	}

	flattenedOverrides := flattenOverrides(nestedOverrides.Map())

	kubecfg, _ := ioutil.ReadFile(kubecfgFile)
	kebCluster := keb.Cluster{
		Kubeconfig: string(kubecfg),
		KymaConfig: keb.KymaConfig{
			Version:    "main",
			Profile:    "evaluation",
			Components: defaultComponentList(flattenedOverrides),
		},
	}

	workerFactory, _ := scheduler.NewLocalWorkerFactory(
		&cluster.MockInventory{},
		scheduler.NewDefaultOperationsRegistry(),
		func(component string, status reconciler.Status) {
			fmt.Printf("Component %s has status %s\n", component, status)
		},
		true)

	localScheduler := scheduler.NewLocalScheduler(workerFactory,
		scheduler.WithPrerequisites("cluster-essentials", "istio-configuration", "certificates"),
		scheduler.WithCRDComponents("cluster-essentials", "istio-configuration"))
	err = localScheduler.Run(context.TODO(), &kebCluster)
	if err != nil {
		return err
	}

	// TODO see stable deploy command
	// certificates
	// printsummary

	//cmd.duration = time.Since(start)

	return nil
}

func flattenOverrides(overrides map[string]interface{}) map[string]string {

	result := make(map[string]string)

	for key, v := range overrides {
		if valueAsMap, ok := v.(map[string]interface{}); ok {
			mapWithIncompleteKeys := flattenOverrides(valueAsMap)
			for k1, v1 := range mapWithIncompleteKeys {
				result[key+"."+k1] = v1
			}
		} else {
			result[key] = fmt.Sprintf("%v", v)
		}
	}

	return result
}
