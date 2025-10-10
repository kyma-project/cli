package app

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/env"
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
	envs                 types.Map
	fileEnvs             types.SourcedEnvArray
	configmapEnvs        types.SourcedEnvArray
	secretEnvs           types.SourcedEnvArray
	expose               bool
	mountSecrets         []string
	mountConfigmaps      []string
	quiet                bool
}

func NewAppPushCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := appPushConfig{
		KymaConfig: kymaConfig,
		envs:       types.Map{Values: map[string]interface{}{}},
	}

	cmd := &cobra.Command{
		Use:   "push [flags]",
		Short: "Push the application to the Kubernetes cluster",
		Long:  "Use this command to push the application to the Kubernetes cluster.",
		Example: `## Push an application based on its source code located in the current directory:
# The application will be built using Cloud Native Buildpacks:
  kyma app push --name my-app --code-path .

## Push an application based on a Dockerfile located in the current directory:
  kyma app push --name my-app --dockerfile ./Dockerfile --dockerfile-context .

## Push an application based on a pre-built image:
  kyma app push --name my-app --image eu.gcr.io/my-project/my-app:latest

## Push an application and expose it using an APIRule:
  kyma app push --name my-app --code-path . --container-port 8080 --expose --istio-inject=true

## Push an application and set environment variables:
# This flag overrides existing environment variables with the same name from other sources (file, ConfigMap, Secret).
  kyma app push --name my-app --code-path . --env NAME1=VALUE --env NAME2=VALUE2

## Push an application and set environment variables from different sources:
# You can set environment variables using --env-from-file, --env-from-configmap, and --env-from-secret flags
# depending on your needs. You can use these flags multiple times to set more than one environment variable
# or use the '--env' flag to override existing environment variables with the same name.
# To get single key from source or load all keys, use one of the following formats:
# - To get single key use: 'ENV_NAME=RESOURCE:RESOURCE_KEY' or 'name=ENV_NAME,resource=RESOURCE,key=RESOURCE_KEY'
# - To fetch all keys use: 'RESOURCE[:ENVS_PREFIX]' or 'resource=RESOURCE,prefix=ENVS_PREFIX'
  kyma app push --name my-app --code-path . \
	--env-from-file ./my-env-file \ 
	--env-from-file MY_ENV=./my-env-file:key1 \ 
	--env-from-configmap my-configmap:CONFIG_ \
	--env-from-configmap MY_ENV2=my-configmap:key2 \
	--env-from-secret my-secret:SECRET_ \
	--env-from-secret MY_ENV3=my-secret:key3`,

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
	cmd.Flags().BoolVarP(&config.quiet, "quiet", "q", false, "Suppresses non-essential output (prints only the URL of the pushed app, if exposed)")
	cmd.Flags().Var(&config.envs, "env", "Environment variables for the app in format NAME=VALUE")
	cmd.Flags().Var(&config.fileEnvs, "env-from-file", "Environment variables for the app loaded from a file in format ENV_NAME=FILE_PATH:FILE_KEY for a single key or FILE_PATH[:ENVS_PREFIX] to fetch all keys")
	cmd.Flags().Var(&config.configmapEnvs, "env-from-configmap", "Environment variables for the app loaded from a ConfigMap in format ENV_NAME=RESOURCE:RESOURCE_KEY for a single key or RESOURCE[:ENVS_PREFIX] to fetch all keys")
	cmd.Flags().Var(&config.secretEnvs, "env-from-secret", "Environment variables for the app loaded from a Secret in format ENV_NAME=RESOURCE:RESOURCE_KEY for a single key or RESOURCE[:ENVS_PREFIX] to fetch all keys")

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
	cmd.Flags().StringVarP(&config.namespace, "namespace", "n", "default", "Namespace where the app is deployed")
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

	writer := io.Writer(os.Stdout)
	if cfg.quiet {
		writer = io.Discard
	}

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

	fmt.Fprintf(writer, "\nCreating deployment %s/%s\n", cfg.namespace, cfg.name)

	clierr = createDeployment(cfg, client, image, imagePullSecret)
	if clierr != nil {
		return clierr
	}

	if cfg.containerPort.Value != nil {
		fmt.Fprintf(writer, "\nCreating service %s/%s\n", cfg.namespace, cfg.name)
		err := resources.CreateService(cfg.Ctx, client, cfg.name, cfg.namespace, int32(*cfg.containerPort.Value))
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to create service"))
		}
	}

	if cfg.expose {
		fmt.Fprintf(writer, "\nCreating API Rule %s/%s\n", cfg.namespace, cfg.name)
		url := fmt.Sprintf("%s.<CLUSTER_DOMAIN>", cfg.name)

		err := resources.CreateAPIRule(cfg.Ctx, client.RootlessDynamic(), cfg.name, cfg.namespace, cfg.name, uint32(*cfg.containerPort.Value))
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

		fmt.Fprintf(writer, "\nThe %s app is available under the \n", cfg.name)
		// print the URL regardless if in quiet mode
		fmt.Print(url)

	}

	return nil
}

func createDeployment(cfg *appPushConfig, client kube.Client, image, imagePullSecret string) clierror.Error {
	fileEnvs, err := env.BuildEnvsFromFile(cfg.fileEnvs)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to build envs from file"))
	}

	secretEnvs, err := env.BuildEnvsFromSecret(cfg.Ctx, client, cfg.namespace, cfg.secretEnvs)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to build envs from secret"))
	}

	configmapEnvs, err := env.BuildEnvsFromConfigmap(cfg.Ctx, client, cfg.namespace, cfg.configmapEnvs)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to build envs from configmap"))
	}

	envs := append(fileEnvs, secretEnvs...)
	envs = append(envs, configmapEnvs...)

	err = resources.CreateDeployment(cfg.Ctx, client, resources.CreateDeploymentOpts{
		Name:            cfg.name,
		Namespace:       cfg.namespace,
		Image:           image,
		ImagePullSecret: imagePullSecret,
		InjectIstio:     cfg.istioInject,
		SecretMounts:    cfg.mountSecrets,
		ConfigmapMounts: cfg.mountConfigmaps,
		Envs:            envs,
	})
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create deployment"))
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
