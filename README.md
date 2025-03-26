<!-- markdown-link-check-disable-next-line -->

[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/cli)](https://api.reuse.software/info/github.com/kyma-project/cli)

# Kyma CLI

> [!WARNING]
> The Kyma CLI version `v2`, with all commands available within this version, is deprecated. Please start using alpha commands available as part of `v3` and please contribute with usecase ideas and general feedback.

Kyma CLI is a command-line interface tool designed to simplify the use of [Kyma](https://github.com/kyma-project/kyma) for application developers. It helps to manage Kyma modules and deploy simple applications, automating complex tasks with simple commands and accelerating development cycles.

## Install

For the installation instructions, see [How to Install](./docs/user/README.md#how-to-install).

## Usage

Inspect the new available alpha commands by calling the `--help` option:

```sh
kyma alpha --help
```

### Import Image Into Kyma's Internal Docker Registry

> [!NOTE]
> To use the following `image-import` command, you must [install the Docker Registry module](https://github.com/kyma-project/docker-registry?tab=readme-ov-file#install) on your Kyma runtime

```sh
docker pull kennethreitz/httpbin

kyma alpha registry image-import kennethreitz/httpbin:latest
```

Run a Pod from a locally hosted image

```sh
kubectl run my-pod --image=localhost:32137/kennethreitz/httpbin:latest --overrides='{ "spec": { "imagePullSecrets": [ { "name": "dockerregistry-config" } ] } }'
```

## Development

To build a Kyma CLI binary, run:

```sh
go build -o kyma main.go
```

You can run the command directly from the go code. For example:

```sh
go run main.go alpha module list
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
