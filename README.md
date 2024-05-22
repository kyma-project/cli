<!-- markdown-link-check-disable-next-line -->
[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/cli)](https://api.reuse.software/info/github.com/kyma-project/cli)

# Kyma CLI

> [!WARNING]
> The Kyma CLI version `v2`, with all commands available within this version, is deprecated. We've started designing the `v3` commands that will be first released within the `alpha` command group.
> Read more about the decision [here](https://github.com/kyma-project/community/issues/872).

Kyma CLI is a command line tool which supports [Kyma](https://github.com/kyma-project/kyma) users.

## Usage

### Import image into kyma's internal docker registry

> [!NOTE]
> The following `image-import` command requires `docker-registry` module to be [installed](https://github.com/kyma-project/docker-registry?tab=readme-ov-file#install) on you kyma runtime

```
docker pull kennethreitz/httpbin

go run main.go image-import kennethreitz/httpbin:latest
```
Run a pod from locally hosted image
```
kubectl run my-pod --image=localhost:32137/kennethreitz/httpbin:latest --overrides='{ "spec": { "imagePullSecrets": [ { "name": "dockerregistry-config" } ] } }'

```
## Development

To build a kyma cli binary run:
```
go build -o kyma-cli  main.go
```

You can run a command directly from the go code. For example:
```
go run main.go provision --help
```
## Contributing
<!--- mandatory section - do not change this! --->

See the [Contributing Rules](CONTRIBUTING.md).

## Code of Conduct
<!--- mandatory section - do not change this! --->

See the [Code of Conduct](CODE_OF_CONDUCT.md) document.

## Licensing
<!--- mandatory section - do not change this! --->

See the [license](LICENSE) file.
