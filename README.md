<!-- markdown-link-check-disable-next-line -->
[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/cli)](https://api.reuse.software/info/github.com/kyma-project/cli)

# Kyma CLI

> [!WARNING]
> The Kyma CLI version `v2`, with all commands available within this version, is deprecated. We've started designing the `v3` commands that will be first released within the `alpha` command group.
> Read more about the decision [here](https://github.com/kyma-project/community/issues/872).

Kyma CLI is a command line tool which supports [Kyma](https://github.com/kyma-project/kyma) users.

## Install 

> [!WARNING]
> `v3` is still in the prototyping stage. All commands are still considered alpha - use it at your own risk.

Download the latest build from the main branch from [v0.0.0-dev](https://github.com/kyma-project/cli/releases/tag/v0.0.0-dev) release assets.

To get Kyma CLI for MacOS or Linux, run the following script from the project's root folder:
```sh
curl -L "https://github.com/kyma-project/cli/releases/download/v0.0.0-dev/kyma_$(uname -s)_$(uname -m).tar.gz" | tar -zxvf - -C bin kyma && mv bin/kyma bin/kyma@v3
```

This script downloads a Linux/MacOS variant of Kyma CLI v3 binary into `bin/kyma@v3`.

Inspect the new available alpha commands by calling the `--help` option: 

```sh
 ./bin/kyma@v3 alpha  --help
```


## Usage

### Import Image Into Kyma's Internal Docker Registry

> [!NOTE]
> To use the following `image-import` command, you must [install the Docker Registry module](https://github.com/kyma-project/docker-registry?tab=readme-ov-file#install) on your Kyma runtime

```
docker pull kennethreitz/httpbin

./bin/kyma@v3 alpha image-import kennethreitz/httpbin:latest
```
Run a Pod from a locally hosted image
```
kubectl run my-pod --image=localhost:32137/kennethreitz/httpbin:latest --overrides='{ "spec": { "imagePullSecrets": [ { "name": "dockerregistry-config" } ] } }'

```
## Development

To build a Kyma CLI binary, run:
```
go build -o kyma-cli  main.go
```

You can run the command directly from the go code. For example:
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
