package config

import (
	"fmt"
	"os"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/registry"
	"github.com/spf13/cobra"
)

type cfgConfig struct {
	*cmdcommon.KymaConfig

	externalurl bool
	output      string
}

func NewConfigCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := cfgConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Saves Kyma registry dockerconfig to a file",
		Long:  "Use this command to save Kyma registry dockerconfig to a file",
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runConfig(&cfg))
		},
	}

	cmd.Flags().BoolVar(&cfg.externalurl, "externalurl", false, "External URL for the Kyma registry.")
	cmd.Flags().StringVar(&cfg.output, "output", "", "Path where the output file should be saved to. NOTE: docker expects the file to be named `config.json`.")

	return cmd
}

func runConfig(cfg *cfgConfig) clierror.Error {
	registryConfig, err := registry.GetExternalConfig(cfg.Ctx, cfg.KubeClient)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to load in-cluster registry configuration"))
	}

	if cfg.externalurl && cfg.output == "" {
		fmt.Print(registryConfig.SecretData.PushRegAddr)
		return nil
	}

	if cfg.externalurl && cfg.output != "" {
		writeErr := os.WriteFile(cfg.output, []byte(registryConfig.SecretData.PushRegAddr), os.ModePerm)
		if writeErr != nil {
			return clierror.New("failed to write docker config to file")
		}
		return nil
	}

	if cfg.output == "" {
		fmt.Print(registryConfig.SecretData.DockerConfigJSON)
	} else {
		writeErr := os.WriteFile(cfg.output, []byte(registryConfig.SecretData.DockerConfigJSON), os.ModePerm)
		if writeErr != nil {
			return clierror.New("failed to write docker config to file")
		}
		fmt.Print("Docker config saved to ", cfg.output)
	}

	return nil
}
