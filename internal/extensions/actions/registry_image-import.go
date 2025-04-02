package actions

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/kyma-project/cli.v3/internal/registry"
	"github.com/spf13/cobra"
)

type registryImageImportActionConfig struct {
	Image string `yaml:"image"`
}

func NewRegistryImageImport(kymaConfig *cmdcommon.KymaConfig, actionConfig types.ActionConfig) types.CmdRun {
	return func(_ *cobra.Command, args []string) {
		cfg := registryImageImportActionConfig{}
		clierror.Check(parseActionConfig(actionConfig, &cfg))
		clierror.Check(validateImage(cfg.Image))
		clierror.Check(runImageImport(kymaConfig, cfg.Image))
	}
}

func validateImage(image string) clierror.Error {
	imageElems := strings.Split(image, ":")
	if len(imageElems) != 2 {
		return clierror.New(fmt.Sprintf("image '%s' not in expected format 'image:tag'", image))
	}

	return nil
}

func runImageImport(kymaConfig *cmdcommon.KymaConfig, image string) clierror.Error {
	client, err := kymaConfig.GetKubeClientWithClierr()
	if err != nil {
		return err
	}

	registryConfig, err := registry.GetInternalConfig(kymaConfig.Ctx, client)
	if err != nil {
		return err
	}

	fmt.Println("Importing", image)

	pushedImage, err := registry.ImportImage(
		kymaConfig.Ctx,
		image,
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

	fmt.Println("\nSuccessfully imported image")
	fmt.Printf("Use it as '%s' and use the %s secret.\n", pushedImage, registryConfig.SecretName)
	fmt.Printf("\nExample usage:\nkubectl run my-pod --image=%s --overrides='{ \"spec\": { \"imagePullSecrets\": [ { \"name\": \"%s\" } ] } }'\n", pushedImage, registryConfig.SecretName)

	return nil
}
