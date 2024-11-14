package app

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube/resources"
	"github.com/spf13/cobra"
)

type appPushConfig struct {
	*cmdcommon.KymaConfig

	name      string
	namespace string
	image     string
	// containerPort int
}

func NewAppPushCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := appPushConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push the application to the Kyma cluster.",
		Long:  "Use this command to push the application to the Kyma cluster.",
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runAppPush(&config))
		},
	}

	cmd.Flags().StringVar(&config.name, "name", "", "Name of the app")
	cmd.Flags().StringVar(&config.namespace, "namespace", "default", "Namespace where app should be deployed")
	cmd.Flags().StringVar(&config.image, "image", "", "Name of the image to deploy")
	// cmd.Flags().IntVar(&config.containerPort, "containerPort", 80, "")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("image")

	return cmd
}

func runAppPush(cfg *appPushConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}
	err := resources.CreateDeployment(cfg.Ctx, client, cfg.name, cfg.namespace, cfg.image)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create deployment"))
	}

	return nil
}
