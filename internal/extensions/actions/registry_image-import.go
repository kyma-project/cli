package actions

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/registry"
	"github.com/spf13/cobra"
)

type registryImageImportActionConfig struct {
	Image string `yaml:"image"`
}

func (c *registryImageImportActionConfig) validate() clierror.Error {
	imageElems := strings.Split(c.Image, ":")
	if len(imageElems) != 2 {
		return clierror.New(fmt.Sprintf("image '%s' not in expected format 'image:tag'", c.Image))
	}

	return nil
}

type registryImageImportAction struct {
	configurator[registryImageImportActionConfig]

	kymaConfig *cmdcommon.KymaConfig
}

func NewRegistryImageImport(kymaConfig *cmdcommon.KymaConfig) Action {
	return &registryImageImportAction{
		kymaConfig: kymaConfig,
	}
}

func (a *registryImageImportAction) Run(cmd *cobra.Command, _ []string) clierror.Error {
	err := a.cfg.validate()
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

	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Importing", a.cfg.Image)

	pushedImage, err := registry.ImportImage(
		a.kymaConfig.Ctx,
		a.cfg.Image,
		registry.ImportOptions{
			ClusterAPIRestConfig: client.RestConfig(),
			RegistryAuth:         registry.NewBasicAuth(registryConfig.SecretData.Username, registryConfig.SecretData.Password),
			RegistryPullHost:     registryConfig.SecretData.PullRegAddr,
			RegistryPodName:      registryConfig.PodMeta.Name,
			RegistryPodNamespace: registryConfig.PodMeta.Namespace,
			RegistryPodPort:      registryConfig.PodMeta.Port,
		},
	)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to import image to in-cluster registry"))
	}

	fmt.Fprintln(out, "\nSuccessfully imported image")
	fmt.Fprintf(out, "Use it as '%s' and use the %s secret.\n", pushedImage, registryConfig.SecretName)
	fmt.Fprintf(out, "\nExample usage:\nkubectl run my-pod --image=%s --overrides='{ \"spec\": { \"imagePullSecrets\": [ { \"name\": \"%s\" } ] } }'\n", pushedImage, registryConfig.SecretName)

	return nil
}
