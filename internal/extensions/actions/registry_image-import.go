package actions

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/extensions/actions/common"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/kyma-project/cli.v3/internal/registry"
	"github.com/spf13/cobra"
)

type registryImageImportActionConfig struct {
	Image string `yaml:"image"`
}

func (c *registryImageImportActionConfig) validate() clierror.Error {
	imageElems := strings.Split(c.Image, ":")
	if len(imageElems) != 2 {
		return clierror.New(fmt.Sprintf("image '%s' not in the expected format 'image:tag'", c.Image))
	}

	return nil
}

type registryImageImportAction struct {
	common.TemplateConfigurator[registryImageImportActionConfig]

	kymaConfig *cmdcommon.KymaConfig
}

func NewRegistryImageImport(kymaConfig *cmdcommon.KymaConfig) types.Action {
	return &registryImageImportAction{
		kymaConfig: kymaConfig,
	}
}

func (a *registryImageImportAction) Run(cmd *cobra.Command, _ []string) clierror.Error {
	err := a.Cfg.validate()
	if err != nil {
		return err
	}

	client, err := a.kymaConfig.GetKubeClientWithClierr()
	if err != nil {
		return err
	}

	registryConfig, err := registry.GetInternalConfig(a.kymaConfig.Ctx, client)
	if err != nil {
		return err
	}

	pushFunc := registry.NewPushWithPortforwardFunc(
		client.RestConfig(),
		registryConfig.PodMeta.Name,
		registryConfig.PodMeta.Namespace,
		registryConfig.PodMeta.Port,
		registryConfig.SecretData.PullRegAddr,
		registry.NewBasicAuth(registryConfig.SecretData.Username, registryConfig.SecretData.Password),
	)

	out.Msgfln("Importing %s", a.Cfg.Image)

	externalRegistryConfig, err := registry.GetExternalConfig(a.kymaConfig.Ctx, client)
	if err == nil {
		out.Msgln("  Using registry external endpoint")
		// if external connection exists, use it
		pushFunc = registry.NewPushFunc(
			externalRegistryConfig.SecretData.PushRegAddr,
			registry.NewBasicAuth(externalRegistryConfig.SecretData.Username, externalRegistryConfig.SecretData.Password),
		)
	}

	pushedImage, err := registry.ImportImage(
		a.kymaConfig.Ctx,
		a.Cfg.Image,
		pushFunc,
	)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to import the image to the in-cluster Docker registry"))
	}

	pullImageName := fmt.Sprintf("%s/%s", registryConfig.SecretData.PullRegAddr, pushedImage)
	out.Msgln("\nSuccessfully imported image")
	out.Msgfln("Use it as '%s' and use the %s secret.", pullImageName, registryConfig.SecretName)
	out.Msgfln("\nExample usage:\nkubectl run my-pod --image=%s --overrides='{ \"spec\": { \"imagePullSecrets\": [ { \"name\": \"%s\" } ] } }'", pullImageName, registryConfig.SecretName)

	return nil
}
