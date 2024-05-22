<!-- markdown-link-check-disable-next-line -->
[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/cli)](https://api.reuse.software/info/github.com/kyma-project/cli)

> [!IMPORTANT]  
> After introducing modular architecture in Kyma, we had to revisit the purpose of its CLI. 
> 
> We have released the last patch for the v2 version ([2.20.4](https://github.com/kyma-project/cli/releases/tag/2.20.4)) and deprecated all Kyma CLI v2 commands.
>
> A new version (v3) with a whole new set of commands (targeting users of both open source and managed Kyma) will be developed and first introduced within an alpha command group.

# CLI (v3)

## Usage

> The following usage examples will change after the first v3 release. Until then, you can execute commands from the code developed in the main branch.


To build a kyma cli binary run:
```
go build -o kyma-cli  main.go
```

You can run a command directly from the go code. For example:
```
go run main.go provision --help
```

## Development

> Add instructions on how to develop the project or example. It must be clear what to do and, for example, how to trigger the tests so that other contributors know how to make their pull requests acceptable. Include the instructions or provide links to related documentation.

## Contributing
<!--- mandatory section - do not change this! --->

See the [Contributing Rules](CONTRIBUTING.md).

## Code of Conduct
<!--- mandatory section - do not change this! --->

See the [Code of Conduct](CODE_OF_CONDUCT.md) document.

## Licensing
<!--- mandatory section - do not change this! --->

See the [license](LICENSE) file.
