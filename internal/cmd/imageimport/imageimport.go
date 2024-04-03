package imageimport

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/registry"
	"github.com/spf13/cobra"
)

type provisionConfig struct {
	ctx        context.Context
	kubeClient kube.Client

	image      string
	kubeconfig string
}

func NewImportCMD() *cobra.Command {
	config := provisionConfig{}

	cmd := &cobra.Command{
		Use:   "image-import",
		Short: "Import image to in-cluster registry.",
		Long:  `Import image from daemon to in-cluster registry.`,
		Args:  cobra.ExactArgs(1),

		PreRunE: func(_ *cobra.Command, args []string) error {
			if err := config.complete(args); err != nil {
				return err
			}
			return config.validate()
		},
		RunE: func(_ *cobra.Command, args []string) error {
			return runImageImport(&config)
		},
	}

	cmd.Flags().StringVar(&config.kubeconfig, "kubeconfig", "", "Path to the kubeconfig file.")

	return cmd
}

func (pc *provisionConfig) validate() error {
	imageElems := strings.Split(pc.image, ":")
	if len(imageElems) != 2 {
		return fmt.Errorf("image '%s' not in expected format 'image:tag'", pc.image)
	}

	return nil
}

func (pc *provisionConfig) complete(args []string) error {
	pc.image = args[0]

	// TODO: think about timeout and moving context to persistent `kyma` command configuration
	pc.ctx = context.Background()

	var err error
	pc.kubeClient, err = kube.NewClient(pc.kubeconfig)

	return err
}

func runImageImport(config *provisionConfig) error {
	// TODO: Add "serverless is not installed" error message
	registryConfig, err := registry.GetConfig(config.ctx, config.kubeClient.Static())
	if err != nil {
		return fmt.Errorf("failed to load in-cluster registry configuration: %s", err.Error())
	}

	if !registryConfig.IsInternal {
		return errors.New("this command does not work for external docker registry")
	}

	workloadMeta, err := registry.GetWorkloadMeta(config.ctx, config.kubeClient.Static(), registryConfig)
	if err != nil {
		return fmt.Errorf("failed to load in-cluster registry pod content: %s", err.Error())
	}

	fmt.Println("Importing", config.image)

	pushedImage, err := registry.ImportImage(
		config.ctx,
		config.image,
		registry.ImportOptions{
			ClusterAPIRestConfig: config.kubeClient.RestConfig(),
			RegistryAuth:         registry.NewBasicAuth(registryConfig.Username, registryConfig.Password),
			RegistryPullHost:     registryConfig.PullRegAddr,
			RegistryPodName:      workloadMeta.Name,
			RegistryPodNamespace: workloadMeta.Namespace,
			RegistryPodPort:      workloadMeta.Port,
		},
	)
	if err != nil {
		return fmt.Errorf("faile to import image to in-cluster registry: %s", err.Error())
	}

	fmt.Println("\nSuccessfully imported image")
	fmt.Printf("Use it as '%s' and use the serverless-registry-config-default secret.\n", pushedImage)
	fmt.Printf("\nExample usage:\nkubectl run my-pod --image=%s --overrides='{ \"spec\": { \"imagePullSecrets\": [ { \"name\": \"serverless-registry-config-default\" } ] } }'\n", pushedImage)

	return nil
}
