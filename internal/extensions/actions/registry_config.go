package actions

import (
	"context"
	"fmt"
	"os"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/registry"
	"github.com/spf13/cobra"
)

type registryConfigActionConfig struct {
	PushRegAddrOnly bool   `yaml:"pushRegAddrOnly"`
	PullRegAddrOnly bool   `yaml:"pullRegAddrOnly"`
	Output          string `yaml:"output"`
	UseExternal     bool   `yaml:"useExternal"`
}

func NewRegistryConfig(kymaConfig *cmdcommon.KymaConfig, actionConfig types.ActionConfig) types.CmdRun {
	return func(_ *cobra.Command, _ []string) {
		cfg := registryConfigActionConfig{}
		clierror.Check(parseActionConfig(actionConfig, &cfg))
		clierror.Check(runConfig(kymaConfig, &cfg))
	}
}

func runConfig(kymaConfig *cmdcommon.KymaConfig, cfg *registryConfigActionConfig) clierror.Error {
	client, err := kymaConfig.GetKubeClientWithClierr()
	if err != nil {
		return err
	}

	secretData, err := getRegistrySecretData(kymaConfig.Ctx, client, cfg.UseExternal)
	if err != nil {
		return err
	}

	outputString := ""
	if cfg.PushRegAddrOnly {
		outputString = secretData.PushRegAddr
	} else if cfg.PullRegAddrOnly {
		outputString = secretData.PullRegAddr
	} else {
		outputString = secretData.DockerConfigJSON
	}

	if cfg.Output != "" {
		writeErr := os.WriteFile(cfg.Output, []byte(outputString), os.ModePerm)
		if writeErr != nil {
			return clierror.New("failed to write docker config to file")
		}
		return nil
	}

	fmt.Println(outputString)
	return nil
}

func getRegistrySecretData(ctx context.Context, client kube.Client, useExternal bool) (*registry.SecretData, clierror.Error) {
	if useExternal {
		registryConfig, err := registry.GetExternalConfig(ctx, client)
		if err != nil {
			return nil, err
		}

		return registryConfig.SecretData, nil
	}

	registryConfig, err := registry.GetInternalConfig(ctx, client)
	if err != nil {
		return nil, err
	}

	return registryConfig.SecretData, nil
}
