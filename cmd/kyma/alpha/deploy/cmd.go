package deploy

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/download"
	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/internal/overrides"
	"github.com/pkg/errors"
	"io/fs"
	"io/ioutil"
	"strings"

	"github.com/kyma-incubator/reconciler/pkg/cluster"
	"github.com/kyma-incubator/reconciler/pkg/keb"
	"github.com/kyma-incubator/reconciler/pkg/reconciler"
	"github.com/kyma-incubator/reconciler/pkg/reconciler/service"
	"github.com/kyma-incubator/reconciler/pkg/scheduler"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
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

func componentsFromStrings(list []string) []keb.Components {
	var components []keb.Components
	for _, item := range list {
		s := strings.Split(item, "@")
		components = append(components, keb.Components{Component: s[0], Namespace: s[1]})
	}
	return components
}

func defaultComponentList() []keb.Components {
	defaultComponents := []string{
		"cluster-essentials",
		"istio-configuration@istio-system",
		"certificates@istio-system",
		"loggin@kyma-system",
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
	return componentsFromStrings(defaultComponents)
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
	//
	//var callback func(deployment.ProcessUpdate)
	//if !cmd.Verbose {
	//	ui := asyncui.AsyncUI{StepFactory: &cmd.Factory}
	//	callback = ui.Callback()
	//	if err != nil {
	//		return err
	//	}
	//}
	//
	//restConfig, err := config.RestConfig(cfg.KubeconfigSource)
	//if err != nil {
	//	return nil, err
	//}
	//
	//kubeClient, err := kubernetes.NewForConfig(restConfig)
	//if err != nil {
	//	return nil, err
	//}

	//registerOverridesInterceptors(ob, kubeClient, cfg.Log)

	kubecfgFile := kube.KubeconfigPath(cmd.KubeconfigPath)

	_, err = service.NewComponentReconciler("base")
	if err != nil {
		return err
	}


	overrides, err := overridesBuilder.Build()

	for k, v := range overrides.Map() {
		fmt.Printf("Key: %s Value: %v \n", k, v)
	}

	kubecfg, _ := ioutil.ReadFile(kubecfgFile)
	kebCluster := keb.Cluster{
		Kubeconfig: string(kubecfg),
		KymaConfig: keb.KymaConfig{
			Version: "main",
			Profile: "evaluation",
			Components:defaultComponentList(),
		},
	}

	workerFactory, _ := scheduler.NewLocalWorkerFactory(
		&cluster.MockInventory{},
		scheduler.NewDefaultOperationsRegistry(),
		func(component string, status reconciler.Status) {
			fmt.Printf("Component %s has status %s\n", component, status)
		},
		true)

	localScheduler, _ := scheduler.NewLocalScheduler(kebCluster, workerFactory, true)
	err = localScheduler.Run(context.TODO())
	if err != nil {
		return err
	}

	//cmd.duration = time.Since(start)

	return nil
}
