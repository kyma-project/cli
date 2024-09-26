package config

import (
	"fmt"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/registry"
	"github.com/spf13/cobra"
	"os"
)

type cfgConfig struct {
	*cmdcommon.KymaConfig
	cmdcommon.KubeClientConfig

	dockerconfig bool
	externalurl  bool
	output       string
}

func NewConfigCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := cfgConfig{
		KymaConfig:       kymaConfig,
		KubeClientConfig: cmdcommon.KubeClientConfig{},
	}

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Saves Kyma registry dockerconfig to a file",
		Long:  "Use this command to save Kyma registry dockerconfig to a file",
		PreRun: func(_ *cobra.Command, _ []string) {
			clierror.Check(cfg.KubeClientConfig.Complete())
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runConfig(&cfg))
		},
	}

	cfg.KubeClientConfig.AddFlag(cmd)
	cmd.Flags().BoolVar(&cfg.dockerconfig, "dockerconfig", false, "Generate a docker config.json file for the Kyma registry")
	cmd.Flags().BoolVar(&cfg.externalurl, "externalurl", false, "External URL for the Kyma registry")
	cmd.Flags().StringVar(&cfg.output, "output", "config.json", "Path where the output file should be saved to")

	return cmd
}

func runConfig(cfg *cfgConfig) clierror.Error {
	registryConfig, err := registry.GetConfig(cfg.Ctx, cfg.KubeClient)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to load in-cluster registry configuration"))
	}

	if cfg.dockerconfig {
		writeErr := os.WriteFile(cfg.output, []byte(registryConfig.SecretData.DockerConfigJSON), os.ModePerm)
		if writeErr != nil {
			return clierror.New("failed to write docker config to file")
		}
		fmt.Print("Docker config saved to ", cfg.output)
	}

	if cfg.externalurl {
		fmt.Print(registryConfig.SecretData.PushRegAddr)
	}
	return nil
}
