<!-- markdown-link-check-disable-next-line -->

[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/cli)](https://api.reuse.software/info/github.com/kyma-project/cli)

# Kyma CLI

> [!WARNING]
> The Kyma CLI version `v2`, with all commands available within this version, is deprecated. Use the commands available as part of `v3` and contribute with use case ideas and general feedback.

Kyma CLI is a command-line interface tool designed to simplify the use of [Kyma](https://github.com/kyma-project/kyma) for application developers. It helps to manage Kyma modules and deploy simple applications, automating complex tasks with simple commands and accelerating development cycles.

## Install

For the installation instructions, see [How to Install](./docs/user/README.md#how-to-install).

## Usage

Inspect the new available commands by calling the `--help` option:

```sh
kyma --help
```

Run a simple app from the image on Kyma by calling:

```sh
kyma app push --name my-first-kyma-app --image kennethreitz/httpbin --expose --container-port 80


Creating deployment default/my-first-kyma-app

Creating service default/my-first-kyma-app

Creating API Rule default/my-first-kyma-app

The `my-first-kyma-app` application is available under the https://my-first-kyma-app.{CLUSTER_DOMAIN}/ address
```

For more usage scenarios, see [user documentation](./docs/user/README.md).

## Development

To build a Kyma CLI binary, run:

```sh
go build -o kyma main.go
```

You can run the command directly from the go code. For example:

```sh
go run main.go module list
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
