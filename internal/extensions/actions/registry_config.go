package actions

import (
	"context"
	"fmt"
	"os"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/extensions/actions/common"
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

type registryConfigAction struct {
	common.TemplateConfigurator[registryConfigActionConfig]

	kymaConfig *cmdcommon.KymaConfig
}

func NewRegistryConfig(kymaConfig *cmdcommon.KymaConfig) types.Action {
	return &registryConfigAction{
		kymaConfig: kymaConfig,
	}
}

func (a *registryConfigAction) Run(cmd *cobra.Command, _ []string) clierror.Error {
	client, err := a.kymaConfig.GetKubeClientWithClierr()
	if err != nil {
		return err
	}

	secretData, err := getRegistrySecretData(a.kymaConfig.Ctx, client, a.Cfg.UseExternal)
	if err != nil {
		return err
	}

	outputString := ""
	if a.Cfg.PushRegAddrOnly {
		outputString = secretData.PushRegAddr
	} else if a.Cfg.PullRegAddrOnly {
		outputString = secretData.PullRegAddr
	} else {
		outputString = secretData.DockerConfigJSON
	}

	if a.Cfg.Output != "" {
		writeErr := os.WriteFile(a.Cfg.Output, []byte(outputString), 0600)
		if writeErr != nil {
			return clierror.New("failed to write docker config to file")
		}
		return nil
	}
	fmt.Fprintln(cmd.OutOrStdout(), outputString)
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
