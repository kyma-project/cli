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

### Nightly build

Download the latest build from the main branch from [0.0.0-dev](https://github.com/kyma-project/cli/releases/tag/0.0.0-dev) release assets.

To get Kyma CLI for MacOS or Linux, run the following script from the command line:

```sh
curl -sL "https://raw.githubusercontent.com/kyma-project/cli/refs/heads/main/hack/install_cli_nightly.sh" | sh -
kyma@v3
```

### Latest build

Download latest (stable or unstable) v3 build from the [releases](https://github.com/kyma-project/cli/releases) assets.

To get latest Kyma CLI for MacOS or Linux, run the following script from the command line:

```sh
curl -sL "https://raw.githubusercontent.com/kyma-project/cli/refs/heads/main/hack/install_cli_latest.sh" | sh -
kyma@v3
```

### Stable release

The Kyma CLI has not yet been stable released. To test it before the final stage use nightly or the latest.

## Usage

Inspect the new available alpha commands by calling the `--help` option:

```sh
kyma@v3 alpha --help
```

### Import Image Into Kyma's Internal Docker Registry

> [!NOTE]
> To use the following `image-import` command, you must [install the Docker Registry module](https://github.com/kyma-project/docker-registry?tab=readme-ov-file#install) on your Kyma runtime

```sh
docker pull kennethreitz/httpbin

./bin/kyma@v3 alpha registry image-import kennethreitz/httpbin:latest
```

Run a Pod from a locally hosted image

```sh
kubectl run my-pod --image=localhost:32137/kennethreitz/httpbin:latest --overrides='{ "spec": { "imagePullSecrets": [ { "name": "dockerregistry-config" } ] } }'
```

## Development

To build a Kyma CLI binary, run:

```sh
go build -o kyma-cli  main.go
```

You can run the command directly from the go code. For example:

```sh
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
