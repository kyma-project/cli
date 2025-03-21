package imageimport

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/registry"
	"github.com/spf13/cobra"
)

type provisionConfig struct {
	*cmdcommon.KymaConfig

	image string
}

func NewImportCMD(kymaConfig *cmdcommon.KymaConfig, _ interface{}) (*cobra.Command, error) {
	config := provisionConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "image-import <image> [flags]",
		Short: "Import image to in-cluster registry",
		Long:  `Import image from daemon to in-cluster registry.`,
		Args:  cobra.ExactArgs(1),

		PreRun: func(_ *cobra.Command, args []string) {
			config.complete(args)
			clierror.Check(config.validate())
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runImageImport(&config))
		},
	}

	return cmd, nil
}

func (pc *provisionConfig) validate() clierror.Error {
	imageElems := strings.Split(pc.image, ":")
	if len(imageElems) != 2 {
		return clierror.New(fmt.Sprintf("image '%s' not in expected format 'image:tag'", pc.image))
	}

	return nil
}

func (pc *provisionConfig) complete(args []string) {
	pc.image = args[0]
}

func runImageImport(config *provisionConfig) clierror.Error {
	client, err := config.GetKubeClientWithClierr()
	if err != nil {
		return err
	}

	// TODO: Add "serverless is not installed" error message
	registryConfig, err := registry.GetInternalConfig(config.Ctx, client)
	if err != nil {
		return err
	}

	fmt.Println("Importing", config.image)

	pushedImage, err := registry.ImportImage(
		config.Ctx,
		config.image,
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
