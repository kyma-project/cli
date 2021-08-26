package deploy

import (
	"context"
	"io/ioutil"

	"github.com/kyma-incubator/reconciler/pkg/cluster"
	"github.com/kyma-incubator/reconciler/pkg/keb"
	"github.com/kyma-incubator/reconciler/pkg/reconciler"
	"github.com/kyma-incubator/reconciler/pkg/reconciler/service"
	"github.com/kyma-incubator/reconciler/pkg/scheduler"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/spf13/cobra"
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
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"d"},
	}
	return cobraCmd
}

func (cmd *command) Run() error {

	service.NewComponentReconciler("base")

	kubecfgFile := kube.KubeconfigPath(cmd.KubeconfigPath)
	kubecfg, _ := ioutil.ReadFile(kubecfgFile)
	kebCluster := keb.Cluster{
		Kubeconfig: string(kubecfg),
		KymaConfig: keb.KymaConfig{
			Version: "main",
			Profile: "evaluation",
			Components: []keb.Components{
				{Component: "cluster-essentials", Namespace: "kyma-system"},
				{Component: "istio", Namespace: "istio-system"},
			},
		},
	}

	workerFactory, _ := scheduler.NewLocalWorkerFactory(
		&cluster.MockInventory{},
		scheduler.NewDefaultOperationsRegistry(),
		func(component string, status reconciler.Status) {},
		true)

	localScheduler, _ := scheduler.NewLocalScheduler(kebCluster, workerFactory, true)
	localScheduler.Run(context.TODO())

	return nil
}
