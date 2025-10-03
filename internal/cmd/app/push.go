package app

import (
	"fmt"
	"os"
	"time"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/dockerfile"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/resources"
	"github.com/kyma-project/cli.v3/internal/pack"
	"github.com/kyma-project/cli.v3/internal/registry"
	"github.com/spf13/cobra"
)

type appPushConfig struct {
	*cmdcommon.KymaConfig

	name                 string
	namespace            string
	image                string
	imagePullSecretName  string
	dockerfilePath       string
	dockerfileSrcContext string
	dockerfileArgs       types.Map
	packAppPath          string
	containerPort        types.NullableInt64
	istioInject          types.NullableBool
	expose               bool
	mountSecrets         []string
	mountConfigmaps      []string
}

func NewAppPushCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := appPushConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "push [flags]",
		Short: "Push the application to the Kubernetes cluster",
		Long:  "Use this command to push the application to the Kubernetes cluster.",

		PreRun: func(cmd *cobra.Command, args []string) {
			clierror.Check(config.complete())
			clierror.Check(flags.Validate(cmd.Flags(),
				flags.MarkRequired("name"),
				flags.MarkExactlyOneRequired("image", "dockerfile", "code-path"),
				flags.MarkExclusive("dockerfile-context", "image", "code-path"),
				flags.MarkExclusive("dockerfile-build-arg", "image", "code-path"),
				flags.MarkPrerequisites("expose", "container-port"),
				flags.MarkPrerequisites("image-pull-secret", "image"),
			))
			clierror.Check(config.validate())
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runAppPush(&config))
		},
	}

	// common flags
	cmd.Flags().StringVar(&config.name, "name", "", "Name of the app")

	// image flags
	cmd.Flags().StringVar(&config.image, "image", "", "Name of the image to deploy")
	cmd.Flags().StringVar(&config.imagePullSecretName, "image-pull-secret", "", "Name of the Kubernetes Secret with credentials to pull the image")

	// dockerfile flags
	cmd.Flags().StringVar(&config.dockerfilePath, "dockerfile", "", "Path to the Dockerfile")
	cmd.Flags().StringVar(&config.dockerfileSrcContext, "dockerfile-context", "", "Context path for building Dockerfile (defaults to the current working directory)")
	cmd.Flags().Var(&config.dockerfileArgs, "dockerfile-build-arg", "Variables used while building an application from Dockerfile as args")

	// pack flags
	cmd.Flags().StringVar(&config.packAppPath, "code-path", "", "Path to the application source code directory")

	// k8s flags
	cmd.Flags().StringVar(&config.namespace, "namespace", "default", "Namespace where the app is deployed")
	cmd.Flags().Var(&config.containerPort, "container-port", "Port on which the application is exposed")
	cmd.Flags().Var(&config.istioInject, "istio-inject", "Enables Istio for the app")
	cmd.Flags().BoolVar(&config.expose, "expose", false, "Creates an APIRule for the app")
	cmd.Flags().StringArrayVar(&config.mountSecrets, "mount-secret", []string{}, "Mounts Secret content to the "+resources.SecretMountPathPrefix+"<SECRET_NAME> path")
	cmd.Flags().StringArrayVar(&config.mountConfigmaps, "mount-config", []string{}, "Mounts ConfigMap content to the "+resources.ConfigmapMountPathPrefix+"<CONFIGMAP_NAME> path")

	return cmd
}

func (apc *appPushConfig) complete() clierror.Error {
	var err error
	var info os.FileInfo

	if apc.dockerfilePath != "" {
		// add /Dockerfile suffix if path is a directory
		info, err = os.Stat(apc.dockerfilePath)
		if err != nil {
			return clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to get stat info for path: %s", apc.dockerfilePath)))
		}
		if info.IsDir() {
			apc.dockerfilePath = fmt.Sprintf("%s/Dockerfile", apc.dockerfilePath)
		}

		// set dockerfile context to working directory if its empty
		if apc.dockerfileSrcContext == "" {
			apc.dockerfileSrcContext, err = os.Getwd()
			if err != nil {
				return clierror.Wrap(err, clierror.New("failed to get current working directory",
					"Please provide the path to the dockerfile context using --dockerfile-context flag"))
			}
		}
	}

	return nil
}

func (apc *appPushConfig) validate() clierror.Error {
	// TODO: enable this code when api-gateway provide its module configuration (ConfigMap)
	// detect if ApiRule resource is installed on the cluster
	// extensions := apc.GetRawExtensions()
	// if apc.expose && !extensions.ContainResource("ApiRule") {
	// 	return clierror.New(
	// 		"application can't be exposed because ApiRule extension is not detected",
	// 		"make sure api-gateway module is installed",
	// 	)
	// }

	return nil
}

func runAppPush(cfg *appPushConfig) clierror.Error {
	quietMode := flags.GetBoolFlagValue("--quiet") || flags.GetBoolFlagValue("-q")
	image := cfg.image
	imagePullSecret := cfg.imagePullSecretName

	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	if cfg.dockerfilePath != "" || cfg.packAppPath != "" {
		registryConfig, cliErr := registry.GetInternalConfig(cfg.Ctx, client)
		if cliErr != nil {
			return cliErr
		}

		pushedImage, clierr := buildAndImportImage(client, cfg, registryConfig)
		if clierr != nil {
			return clierr
		}
		image = fmt.Sprintf("%s/%s", registryConfig.SecretData.PullRegAddr, pushedImage)
		imagePullSecret = registryConfig.SecretName
	}

	if !quietMode {
		fmt.Printf("\nCreating deployment %s/%s\n", cfg.namespace, cfg.name)
	}

	err := resources.CreateDeployment(cfg.Ctx, client, resources.CreateDeploymentOpts{
		Name:            cfg.name,
		Namespace:       cfg.namespace,
		Image:           image,
		ImagePullSecret: imagePullSecret,
		InjectIstio:     cfg.istioInject,
		SecretMounts:    cfg.mountSecrets,
		ConfigmapMounts: cfg.mountConfigmaps,
	})
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create deployment"))
	}

	if cfg.containerPort.Value != nil {
		if !quietMode {
			fmt.Printf("\nCreating service %s/%s\n", cfg.namespace, cfg.name)
		}
		err = resources.CreateService(cfg.Ctx, client, cfg.name, cfg.namespace, int32(*cfg.containerPort.Value))
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to create service"))
		}
	}

	if cfg.expose {
		if !quietMode {
			fmt.Printf("\nCreating API Rule %s/%s\n", cfg.namespace, cfg.name)
		}
		url := fmt.Sprintf("%s.<CLUSTER_DOMAIN>", cfg.name)

		err = resources.CreateAPIRule(cfg.Ctx, client.RootlessDynamic(), cfg.name, cfg.namespace, cfg.name, uint32(*cfg.containerPort.Value))
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to create API Rule resource", "Make sure API Gateway module is installed", "Make sure APIRule CRD is available in v2 version"))
		}

		// try to get domain from resulting virtual service
		// Check if the user can watch virtualservices
		authRes, authErr := resources.CreateSelfSubjectAccessReview(cfg.Ctx, client, "watch", "virtualservices", cfg.namespace, "networking.istio.io")

		if authErr != nil {
			return clierror.Wrap(authErr, clierror.New("failed to check permissions to get virtualservices"))
		}
		if authRes.Status.Allowed {
			url, clierr = client.Istio().GetHostFromVirtualServiceByApiruleName(cfg.Ctx, cfg.name, cfg.namespace)
			if clierr != nil {
				return clierror.WrapE(clierr, clierror.New("failed to get host address of ApiRule's Virtual Service"))
			}
		}
		if !quietMode {
			fmt.Printf("\nThe %s app is available under the https://%s/ address\n", cfg.name, url)
		} else {
			fmt.Fprint(os.Stdout, url)
		}
	}

	return nil
}

func buildAndImportImage(client kube.Client, cfg *appPushConfig, registryConfig *registry.InternalRegistryConfig) (string, clierror.Error) {
	fmt.Print("Building image\n\n")
	imageName, err := buildImage(cfg)
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("failed to build image from dockerfile"))
	}

	pushFunc := registry.NewPushWithPortforwardFunc(
		client.RestConfig(),
		registryConfig.PodMeta.Name,
		registryConfig.PodMeta.Namespace,
		registryConfig.PodMeta.Port,
		registryConfig.SecretData.PullRegAddr,
		registry.NewBasicAuth(registryConfig.SecretData.Username, registryConfig.SecretData.Password),
	)

	fmt.Println("\nImporting", imageName)
	externalRegistryConfig, cliErr := registry.GetExternalConfig(cfg.Ctx, client)
	if cliErr == nil {
		fmt.Println("  Using registry external endpoint")
		// if external connection exists, use it
		pushFunc = registry.NewPushFunc(
			externalRegistryConfig.SecretData.PushRegAddr,
			registry.NewBasicAuth(externalRegistryConfig.SecretData.Username, externalRegistryConfig.SecretData.Password),
		)
	}

	pushedImage, cliErr := registry.ImportImage(
		cfg.Ctx,
		imageName,
		pushFunc,
	)
	if cliErr != nil {
		return "", clierror.WrapE(cliErr, clierror.New("failed to import image to in-cluster docker registry"))
	}

	return pushedImage, nil
}

func buildImage(cfg *appPushConfig) (string, error) {
	imageTag := time.Now().Format("2006-01-02_15-04-05")
	imageName := fmt.Sprintf("%s:%s", cfg.name, imageTag)

	var err error
	if cfg.packAppPath != "" {
		// build application from sources
		err = pack.Build(cfg.Ctx, imageName, cfg.packAppPath)
	} else {
		// build application from dockerfile
		err = dockerfile.Build(cfg.Ctx, &dockerfile.BuildOptions{
			ImageName:      imageName,
			BuildContext:   cfg.dockerfileSrcContext,
			DockerfilePath: cfg.dockerfilePath,
			Args:           cfg.dockerfileArgs.GetNullableMap(),
		})
	}

	return imageName, err
}
