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
	cmdcommon.KubeClientConfig

	image string
}

func NewImportCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := provisionConfig{
		KymaConfig:       kymaConfig,
		KubeClientConfig: cmdcommon.KubeClientConfig{},
	}

	cmd := &cobra.Command{
		Use:   "image-import",
		Short: "Import image to in-cluster registry.",
		Long:  `Import image from daemon to in-cluster registry.`,
		Args:  cobra.ExactArgs(1),

		PreRun: func(_ *cobra.Command, args []string) {
			clierror.Check(config.complete(args))
			clierror.Check(config.validate())
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runImageImport(&config))
		},
	}

	config.KubeClientConfig.AddFlag(cmd)

	return cmd
}

func (pc *provisionConfig) validate() clierror.Error {
	imageElems := strings.Split(pc.image, ":")
	if len(imageElems) != 2 {
		return clierror.New(fmt.Sprintf("image '%s' not in expected format 'image:tag'", pc.image))
	}

	return nil
}

func (pc *provisionConfig) complete(args []string) clierror.Error {
	pc.image = args[0]

	return pc.KubeClientConfig.Complete()
}

func runImageImport(config *provisionConfig) clierror.Error {
	// TODO: Add "serverless is not installed" error message
	registryConfig, err := registry.GetConfig(config.Ctx, config.KubeClient)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to load in-cluster registry configuration"))
	}

	fmt.Println("Importing", config.image)

	pushedImage, err := registry.ImportImage(
		config.Ctx,
		config.image,
		registry.ImportOptions{
			ClusterAPIRestConfig: config.KubeClient.RestConfig(),
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
