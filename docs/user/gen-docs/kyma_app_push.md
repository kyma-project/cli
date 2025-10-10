# kyma app push

Push the application to the Kubernetes cluster.

## Synopsis

Use this command to push the application to the Kubernetes cluster.

```bash
kyma app push [flags]
```

## Examples

```bash
## Push an application based on its source code located in the current directory:
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
	--env-from-secret MY_ENV3=my-secret:key3
```

## Flags

```text
      --code-path string                   Path to the application source code directory
      --container-port int                 Port on which the application is exposed
      --dockerfile string                  Path to the Dockerfile
      --dockerfile-build-arg stringArray   Variables used while building an application from Dockerfile as args
      --dockerfile-context string          Context path for building Dockerfile (defaults to the current working directory)
      --env stringArray                    Environment variables for the app in format NAME=VALUE
      --env-from-configmap stringArray     Environment variables for the app loaded from a ConfigMap in format ENV_NAME=RESOURCE:RESOURCE_KEY for a single key or RESOURCE[:ENVS_PREFIX] to fetch all keys
      --env-from-file stringArray          Environment variables for the app loaded from a file in format ENV_NAME=FILE_PATH:FILE_KEY for a single key or FILE_PATH[:ENVS_PREFIX] to fetch all keys
      --env-from-secret stringArray        Environment variables for the app loaded from a Secret in format ENV_NAME=RESOURCE:RESOURCE_KEY for a single key or RESOURCE[:ENVS_PREFIX] to fetch all keys
      --expose                             Creates an APIRule for the app
      --image string                       Name of the image to deploy
      --image-pull-secret string           Name of the Kubernetes Secret with credentials to pull the image
      --istio-inject                       Enables Istio for the app
      --mount-config stringArray           Mounts ConfigMap content to the /bindings/configmap-<CONFIGMAP_NAME> path (default "[]")
      --mount-secret stringArray           Mounts Secret content to the /bindings/secret-<SECRET_NAME> path (default "[]")
      --name string                        Name of the app
  -n, --namespace string                   Namespace where the app is deployed (default "default")
  -q, --quiet                              Suppresses non-essential output (prints only the URL of the pushed app, if exposed)
  -h, --help                               Help for the command
      --kubeconfig string                  Path to the Kyma kubeconfig file
      --show-extensions-error              Prints a possible error when fetching extensions fails
      --skip-extensions                    Skip fetching extensions from the target Kyma environment
```

## See also

* [kyma app](kyma_app.md) - Manages applications on the Kubernetes cluster
