package app

import (
	"fmt"
	"os"
	"time"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/envs"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/docker"
	"github.com/kyma-project/cli.v3/internal/dockerfile"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/resources"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/kyma-project/cli.v3/internal/pack"
	"github.com/kyma-project/cli.v3/internal/registry"
	"github.com/spf13/cobra"
)

type appPushConfig struct {
	*cmdcommon.KymaConfig

	name                       string
	namespace                  string
	image                      string
	imagePullSecretName        string
	dockerfilePath             string
	dockerfileSrcContext       string
	dockerfileArgs             types.Map
	packAppPath                string
	containerPort              types.NullableInt64
	istioInject                types.NullableBool
	envs                       types.EnvMap
	fileEnvs                   types.SourcedEnvArray
	configmapEnvs              types.SourcedEnvArray
	secretEnvs                 types.SourcedEnvArray
	expose                     bool
	mountSecrets               types.MountArray
	mountConfigmaps            types.MountArray
	mountServiceBindingSecrets types.ServiceBindingSecretArray
	quiet                      bool
	insecure                   bool
}

func NewAppPushCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := appPushConfig{
		KymaConfig: kymaConfig,
		envs:       types.EnvMap{Map: &types.Map{Values: map[string]interface{}{}}},
	}

	cmd := &cobra.Command{
		Use:   "push [flags]",
		Short: "Push the application to the Kubernetes cluster",
		Long:  "Use this command to push the application to the Kubernetes cluster.",
		Example: `  ## Push an application based on its source code located in the current directory:
  # The application will be built using Cloud Native Buildpacks:
  kyma app push --name my-app --code-path .

  # Push an application based on a Dockerfile located in the current directory:
  kyma app push --name my-app --dockerfile ./Dockerfile --dockerfile-context .

  # Push an application based on a pre-built image:
  kyma app push --name my-app --image eu.gcr.io/my-project/my-app:latest

  # Push an application and expose it using an APIRule:
  kyma app push --name my-app --code-path . --container-port 8080 --expose --istio-inject=true

  ## Push an application and set environment variables:
  #  This flag overrides existing environment variables with the same name from other sources (file, ConfigMap, Secret).
  #  To set an environment variable, use the format 'NAME=VALUE' or 'name=<NAME>,value=<VALUE>'.
  kyma app push --name my-app --code-path . --env NAME1=VALUE --env NAME2=VALUE2

  ## Push an application and set environment variables from different sources:
  #  You can set environment variables using --env-from-file, --env-from-configmap, and --env-from-secret flags
  #  depending on your needs. You can use these flags multiple times to set more than one environment variable
  #  or use the '--env' flag to override existing environment variables with the same name.
  #  To get a single key from source or load all keys, use one of the following formats:
  #  - To get a single key, use: 'ENV_NAME=RESOURCE:RESOURCE_KEY' or 'name=ENV_NAME,resource=RESOURCE,key=RESOURCE_KEY'
  #  - To fetch all keys, use: 'RESOURCE[:ENVS_PREFIX]' or 'resource=RESOURCE,prefix=ENVS_PREFIX'
  kyma app push --name my-app --code-path . \
    --env-from-file ./my-env-file \ 
    --env-from-file MY_ENV=./my-env-file:key1 \ 
    --env-from-configmap my-configmap:CONFIG_ \
    --env-from-configmap MY_ENV2=my-configmap:key2 \
    --env-from-secret my-secret:SECRET_ \
    --env-from-secret MY_ENV3=my-secret:key3

  ## Push an application and mount a Secret, ConfigMap or Service Binding Secret:
  #  Depending on your needs, you can mount specific keys or the whole resource.
  #  You can use these flags multiple times to mount more than one resource.
  #  Flags --mount-secret and --mount-config support below formats:
  #  - Normal format: 'name=NAME,path=MOUNT_PATH,key=KEY,ro=READ_ONLY' (key and ro are optional)
  #  - Shorthand format: 'NAME:KEY=MOUNT_PATH:ro' (ro is optional)
  #  - Legacy format: 'NAME' (mounts the whole secret at /bindings/secret-<NAME> or configmap at /bindings/configmap-<NAME>) 
  kyma app push --name my-app --code-path . \
    --mount-secret name=my-secret,path=/app/secret,key=secret-key,ro=true \
    --mount-secret my-secret:secret-key=/app/secret:ro \
    --mount-secret my-secret 
    --mount-config name=my-configmap,path=/app/config,key=config-key \
    --mount-config my-configmap:config-key=/app/config:ro \
    --mount-service-binding-secret my-service-binding-secret`,

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
	cmd.Flags().BoolVar(&config.insecure, "insecure", false, "Disables SecurityContext configuration for the app deployment")
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
	istioInjectFlag := cmd.Flags().VarPF(&config.istioInject, "istio-inject", "", "Enables Istio for the app")
	istioInjectFlag.NoOptDefVal = "true" // default value when flag is provided without value
	cmd.Flags().BoolVar(&config.expose, "expose", false, "Creates an APIRule for the app")
	cmd.Flags().Var(&config.mountSecrets, "mount-secret", "Mounts Secret content. Format: 'name=secret,path=/app/config,key=key,ro=true' or shorthand 'secret:key=/app/config:ro'. Path traversal (..) is prohibited.")
	cmd.Flags().Var(&config.mountConfigmaps, "mount-config", "Mounts ConfigMap content. Format: 'name=configmap,path=/app/config,key=key,ro=false' or shorthand 'configmap:key=/app/config'. Path traversal (..) is prohibited.")
	cmd.Flags().Var(&config.mountServiceBindingSecrets, "mount-service-binding-secret", "Mounts Secret as service binding at /bindings/secret-<NAME> (readOnly)")

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
	if cfg.quiet {
		out.DisableMsg()
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

	out.Msgfln("\nCreating deployment %s/%s", cfg.namespace, cfg.name)

	clierr = createDeployment(cfg, client, image, imagePullSecret)
	if clierr != nil {
		return clierr
	}

	if cfg.containerPort.Value != nil {
		out.Msgfln("\nCreating service %s/%s", cfg.namespace, cfg.name)
		err := resources.CreateService(cfg.Ctx, client, cfg.name, cfg.namespace, int32(*cfg.containerPort.Value))
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to create service"))
		}
	}

	if cfg.expose {
		out.Msgfln("\nCreating API Rule %s/%s", cfg.namespace, cfg.name)
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

		out.Msgfln("\nThe %s app is available under the", cfg.name)
		// print the URL regardless if in quiet mode
		out.Prio(url)

	}

	return nil
}

func createDeployment(cfg *appPushConfig, client kube.Client, image, imagePullSecret string) clierror.Error {
	configmapEnvs, err := envs.BuildFromConfigmap(cfg.Ctx, client, cfg.namespace, cfg.configmapEnvs)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to build envs from configmap"))
	}

	secretEnvs, err := envs.BuildFromSecret(cfg.Ctx, client, cfg.namespace, cfg.secretEnvs)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to build envs from secret"))
	}

	fileEnvs, err := envs.BuildFromFile(cfg.fileEnvs)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to build envs from file"))
	}

	plainEnvs := envs.Build(cfg.envs)

	// append envs in order to have correct precedence
	// plain envs > file envs > secret envs > configmap envs
	envs := configmapEnvs
	envs = append(envs, secretEnvs...)
	envs = append(envs, fileEnvs...)
	envs = append(envs, plainEnvs...)

	err = resources.CreateDeployment(cfg.Ctx, client, resources.CreateDeploymentOpts{
		Name:                       cfg.name,
		Namespace:                  cfg.namespace,
		Image:                      image,
		ImagePullSecret:            imagePullSecret,
		InjectIstio:                cfg.istioInject,
		SecretMounts:               cfg.mountSecrets,
		ConfigmapMounts:            cfg.mountConfigmaps,
		ServiceBindingSecretMounts: cfg.mountServiceBindingSecrets,
		Envs:                       envs,
		Insecure:                   cfg.insecure,
	})
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create deployment"))
	}

	return nil
}

func buildAndImportImage(client kube.Client, cfg *appPushConfig, registryConfig *registry.InternalRegistryConfig) (string, clierror.Error) {
	out.Msgln("Building image\n")
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

	out.Msgfln("\nImporting %s", imageName)
	externalRegistryConfig, cliErr := registry.GetExternalConfig(cfg.Ctx, client)
	if cliErr == nil {
		out.Msgln("  Using registry external endpoint")
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
		err = dockerfile.Build(cfg.Ctx, docker.BuildOptions{
			ImageName:      imageName,
			BuildContext:   cfg.dockerfileSrcContext,
			DockerfilePath: cfg.dockerfilePath,
			Args:           cfg.dockerfileArgs.GetNullableMap(),
		})
	}

	return imageName, err
}
