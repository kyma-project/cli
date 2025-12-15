# Command structure standards

Creation date: 2025.11.21

## Description

I would collect and propose all requirements and good practices that all new commands should meet before moving them out of the `alpha` group.

This document is focused on what users see, and it's not about command functionalities.

## Name

The command's name field is defined under the `cobra.Command{}.Use` field. This field must contain the command name and the symbolic visualisation of allowed arguments and flags.

### Names in practice

Command groups:

```yaml
function <command> [flags] # function command group that is not runnable and can receive optional flags
```

Runnable commands:

```yaml
explain [flags] # runnable command without argument
create <name> [flags] # runnable command that receive one required arg and optional flags
scale <name> <replicas> [flags] # runnable command that can receive two arguments and optional flags
delete <name>... # runnable command that receive at least one argument and has no flags
get [<name>...] [flags] # runnable command that receive optional one argument and optional flags
module-deploy (<name>|<namespacedName>|<filepath>)... [flags] # runnable command that deploy module based on at least one argument in one of three types
```

Important elements:

- `<>` - means this is positional argument
- `[]` - means elemts are optional
- `...` - means at least one element is required (this applies only for positional arguments)
- `(a|b)` - means element `a` or `b` is possible

### Name details

The command name must be a verb or noun. In most cases, it is recommended to use nouns to define domain-related command groups (like `kyma function`, `kyma app`, or `kyma apirule`) and verbs to define runnable operations around the domain (like `kyma app push`, `kyma function create`, `kyma apirule expose`). This rule is only a suggestion, and it depends on the use case. For example, the `kyma diagnose` command works as a runnable command and command group at the same (`kyma diagnose logs`) time, and because of this, it's not a noun. On the other hand, the `kyma registry config` in another example, where after the noun is another noun (runnable noun), but in this case, the `config` word is shorter than verbs like `get-config` or `configure`.

After the first word, there must be a description of possible arguments/commands. If the command is a non-runnable command group, then it should contain `<command>`, which means that this command accepts only one argument, and this argument is the command name. If command is runnable then is must describes possible inputs (if allowed) in the following format: `<arg_name>` for single, required argument, `<arg_name>...` for at least one required argument, `[<arg_name>...]` for one optional argument, `<arg_name>...` for optional arguments list of the same type. If the command receives more than one argument type, then it is possible to describe many arguments separated by a space. For example: `scale <name> <replicas>`

The last element must be optional flags represented by the `[flags]` element. In our case, it must be a part of every command because we add persistent flags for the parent `kyma` command, and these flags are valid for every sub-command.

## Descriptions

There are two types of description under the `cobra.Command{}.Short` and the `cobra.Command{}.Long` fields. The first one represents a shorter description that is displayed when running parent commands help, and the second one is displayed when running the current command.

### Descriptions in practice

For the long and short description:

```yaml
Lists modules catalog # short description of the catalog command
Use this command to list all available Kyma modules. # long description of the catalog command
```

The parents help:

```console
$ kyma module --help

Use this command to manage modules in the Kyma cluster.

Usage:
  kyma module [command]

Available Commands:
  catalog     Lists modules catalog # short description of the catalog command
```

The commands help:

```console
$ kyma module catalog --help

Use this command to list all available Kyma modules. # long description of the catalog command

Usage:
  kyma module catalog [flags]
```

### Description details

The short desc helps users choose the right sub-command in the context of the domain they are in. This description must start with a capital letter and end without a period.

The longer one describes exactly what the command is doing. It can be multiline, describing all use-cases. It must start with a capital letter and always end with a period.

## Flags

Flags are elements of the command that can be added using the `cobra.Command{}.Flags()` and `cobra.Command{}.PersistentFlags()` functions.

### Flags in practice

For flags configuration:

```golang
cmd.Flags().StringVarP(&cfg.channel, "channel", "c", "fast", "Name of the Kyma channel to use for the module")
cmd.Flags().StringVar(&cfg.crPath, "cr-path", "", "Path to the custom resource file")
cmd.Flags().BoolVar(&cfg.defaultCR, "default-cr", false, "Deploys the module with the default CR")
cmd.Flags().BoolVar(&cfg.autoApprove, "auto-approve", false, "Automatically approve community module installation")
cmd.Flags().StringVar(&cfg.version, "version", "", "Specifies version of the community module to install")
cmd.Flags().StringVar(&cfg.modulePath, "origin", "", "Specifies the source of the module (kyma or custom name)")
_ = cmd.Flags().MarkHidden("origin")
cmd.Flags().BoolVar(&cfg.community, "community", false, "Installs a community module (no official support, no binding SLA)")
_ = cmd.Flags().MarkHidden("community")
```

The commands help looks:

```console
$ kyma module add --help

...

Flags:
      --auto-approve     Automatically approve community module installation
  -c, --channel string   Name of the Kyma channel to use for the module (default "fast")
      --cr-path string   Path to the custom resource file
      --default-cr       Deploys the module with the default CR
      --version string   Specifies version of the community module to install
```

### Flags details

The most important thing from the user perspective is flag description, defaulting and validation.

The description should be as minimalistic as possible, but describe for which flag it can be used. If the flag is related to only one command's use case, then it's a good opportunity to add example usage of such a command (read more in the [Examples](#examples) section). Every description should start with a capital letter and end with no period.

If possible, every flag should have its own default value passed to the flag building function (to display it in commands' help).

Flag shorthand should be added only to flags that are often used in most cases to speed up typing. The shorthand should be the first letter of the full flag name (for example, `-c` for `--channel`).

Flags can be marked as hidden. This functionality may be helpful while some flags are deprecated and their functionality is removed or moved. Hiding the flag allows us to validate later if the user uses it and display a well-crafted error with detailed information on what is happening and where such functionality was moved to.

To validate flags, the [flags](../../../internal/flags/validate.go) package must be used to keep all validations of all commands in the same shape. Functionality of this package can be run in the `cobra.Command{}.PreRun`.

Persistent flags need to meet all requirements above and additionally should be implemented only in common use-cases for all referred commands. These flag types can introduce confusion when implemented for commands and don't provide any functionality.

## Examples

Examples is an optional field that is highly recommended to use. It's under the `cobra.Command{}.Examples` field and can be used to propose the most common use cases or propositions of flag usage.

### Examples in practice

For the following examples:

```yaml
  # Add a Kyma module with the default CR
  kyma module add kyma-module --default-cr

  # Add a Kyma module with a custom CR from a file
  kyma module add kyma-module --cr-path ./kyma-module-cr.yaml

  ## Add a community module with a default CR and auto-approve the SLA
  #  passed argument must be in the format <namespace>/<module-template-name>
  kyma module add my-namespace/my-module-template-name --default-cr --auto-approve
```

The help will displays:

```console
$ kyma module add --help

Use this command to add a module.

Usage:
  kyma module add <module> [flags]

Examples:
  # Add a Kyma module with the default CR
  kyma module add kyma-module --default-cr

  # Add a Kyma module with a custom CR from a file
  kyma module add kyma-module --cr-path ./kyma-module-cr.yaml

  ## Add a community module with a default CR and auto-approve the SLA
  #  passed argument must be in the format <namespace>/<module-template-name>
  kyma module add my-namespace/my-module-template-name --default-cr --auto-approve
```

### Examples details

Every line of the examples must start with double spaces to display them correctly in the terminal.

All examples for a single command must be separated by an empty line and include their own description, starting with the `#` symbol. If the example description is longer than one line, then the first line should start with the `##` prefix, and then every next line should start with `#` and two spaces after.

Examples should reflect real use cases, so, if possible, they should use real data as arguments and flags. If not, then use fiction one representing real data, like `my-module`, `my-resource`, `my-something`.

## Aliases

The alias is an array table in the `cobra.Command{}.Aliases`, and there are no specific requirements about command aliases. The good idea is always to use this functionality to provide a shorthand of the command (`del` for `delete` command), or a word form that can help with avoiding small typos (`modules` for `module` command).

### Aliases in practice

```go
cmd := &cobra.Command{
  Use:   "delete <module> [flags]",
  Aliases: []string{"del"},
}
```

## Errors and Hints

Errors were proposed in the [ADR 001](001-error-output-format.md) proposal and implemented quite a while after that. With this functionality, users can understand what is going on at three levels of abstraction. The general message called `Error` should contain the last, user-understandable operation that fails. The second thing is `Error Details`, which contains an internal error message generated by the library or the server. The `Hints` section is designed to help users identify how to fix the problem. Every hint should be in one of two formats:

- `to <what>, <do>` - format used to describe possible optional configurations that may be used or misconfigured
- `make sure <what>` - format used to describe the required configuration user may have misconfigured

The CLI is designed to always return `clierror.Error` instead of pure `error`. Both errors are not compatible with each other, and to avoid user confusion, we should not use the `cobra.Command{}.RunE` and instead of that use the `cobra.Command{}.Run` and check the error manually inside of it:

```go
cmd := &cobra.Command{
  Run: func(cmd *cobra.Command, _ []string) {
    clierror.Check(runAdd(&cfg)) // check manually using the `clierror.Check` function
  },
}
```

### Example hints

```yaml
"make sure you provide a valid module name and channel (or version)",
"to list available modules, call the `kyma module catalog` command",
"to pull available modules, call the `kyma module pull` command",
"to add a community module, use the `--origin` flag",
```

## Command messaging

It's not allowed to use the `os.Stdout`/`os.Stderr` and `fmt.Print` functionalities. Instead of that we must use the `internal/out` package. The main reason of using it is to keep control over `stdout` and `stderr` channels in one pleace. This allows us to:

- split messages by calling specific functions like `out.Err`, to send a message to `stderr`, or one of `out.Msg`, `out.Prio`, `out.Verbose`, or `out.Debug` to send it to the `stdout`
- mute less crucial output by calling the `out.DisableMsg` function that disables the `out.Msg`
- enable cli verbosity by calling the `out.EnableVerbose` function that enables the `out.Verbose`
- enable command's debug info by running it with the `--debug` flag. It allows printing messages passed to the `out.Debug` function. This solution is designed for developers who are working on the CLI. This flag is not mentioned in the user documentation. The `out.Debug` functionality is the only one that should not be re-configured in commands business logic - the `--debug` flag is defined as a persistent flag for all sub-commands, and no command should re-define this flag on its own or mask by defining a local flag with the same name

## Example command in code

The command configuration below applies all rules described above:

```go
func newCMD() *cobra.Command {
  cfg := addConfig{
    KymaConfig: kymaConfig,
  }

  cmd := &cobra.Command{
    Use:   "add <module> [flags]",
    Short: "Add a module",
    Long:  "Use this command to add a module.",
    Example: `  # Add a Kyma module with the default CR
  kyma module add kyma-module --default-cr

  # Add a Kyma module with a custom CR from a file
  kyma module add kyma-module --cr-path ./kyma-module-cr.yaml

  ## Add a community module with a default CR and auto-approve the SLA
  #  passed argument must be in the format <namespace>/<module-template-name>
  kyma module add my-namespace/my-module-template-name --default-cr --auto-approve`,

    Args: cobra.ExactArgs(1),
    PreRun: func(cmd *cobra.Command, _ []string) {
      clierror.Check(flags.Validate(cmd.Flags(),
        flags.MarkMutuallyExclusive("cr-path", "default-cr"),
        flags.MarkUnsupported("community", "the --community flag is no longer supported - community modules need to be pulled first using 'kyma module pull' command, then installed"),
        flags.MarkUnsupported("origin", "the --origin flag is no longer supported - use commands argument instead"),
      ))
    },
    Run: func(cmd *cobra.Command, args []string) {
      cfg.complete(args)
      clierror.Check(runAdd(&cfg))
    },
  }

  cmd.Flags().StringVarP(&cfg.channel, "channel", "c", "", "Name of the Kyma channel to use for the module")
  cmd.Flags().StringVar(&cfg.crPath, "cr-path", "", "Path to the custom resource file")
  cmd.Flags().BoolVar(&cfg.defaultCR, "default-cr", false, "Deploys the module with the default CR")
  cmd.Flags().BoolVar(&cfg.autoApprove, "auto-approve", false, "Automatically approve community module installation")
  cmd.Flags().StringVar(&cfg.version, "version", "", "Specifies version of the community module to install")
  cmd.Flags().StringVar(&cfg.modulePath, "origin", "", "Specifies the source of the module (kyma or custom name)")
  _ = cmd.Flags().MarkHidden("origin")
  cmd.Flags().BoolVar(&cfg.community, "community", false, "Install a community module (no official support, no binding SLA)")
  _ = cmd.Flags().MarkHidden("community")

  return cmd
}
```
