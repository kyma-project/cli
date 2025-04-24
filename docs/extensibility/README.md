# Extensibility

## Overview

The Kyma CLI acts like any other CLI, it means that it compiles every functionality into its binary. With the extensibility, you can create module-oriented functionality versioned and kept together with the module on a cluster. This feature may be used to extend the CLI with resource-oriented commands, allowing module resource management, or module-oriented commands allowing interaction with a module.

Extensions can be added by creating a ConfigMap with the expected label and data in the expected format. The CLI binary with access to listing ConfigMaps on a cluster will fetch all extensions from it when run and build additional commands based on them.

All commands build from extensions can be accessed under the `kyma alpha` commands group.

## Concept

The main goal of CLI extensibility is to support the basic resources and modules of Kyma so that the end user can intuitively use Kyma and CLI. After creating a cluster, the CLI should contain only the necessary and basic functionality, but with the installation of additional modules, the CLI should be extended with additional commands, built based on extensions provided with the modules.

Such a solution provides control over the interaction with the module on the side of the team responsible for the team, but maintaining a uniform standard defined on the CLI side.

An additional advantage is that there is no need to migrate extensions on the CLI code side. If the team wants to introduce a change in the definition of a command/group of commands, the only thing that needs to be done is to release a new version of the module containing the updated version of the extension. For example, if the team responsible for the APIRule resource in version v2alpha1 created an extension that allows adding and removing APIRule resources and wants to release a new version of the resource, all they need to do is, along with adding a new version to the module, update the extension and release a new version of the module. Also possible is having different versions of the extension for every component version and/or on every release channel.

## ConfigMap

The extension is defined and enabled by right prepared ConfigMap deployed on a cluster that CLI has access to (for example by exporting the `KUBECONFIG` env or passing right argument to the `--kubeconfig` flag). The ConfigMap can has any name and be located in any namespace but should contains the `kyma-cli/extension: commands` and `kyma-cli/extension-version: v1` labels, and the `kyma-commands.yaml` data key with right the extension configuration. An example ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-configmap
  labels:
    kyma-cli/extension: commands
    kyma-cli/extension-version: v1
data:
  kyma-commands.yaml: |-
    metadata:
      name: my-command
```

[Here](https://github.com/kyma-project/serverless/blob/main/config/serverless/templates/cli-extension.yaml) is the example of the Serverless module extension ConfigMap.

## kyma-commands.yaml

The extension definition is represented by the YAML file put iside the `kyma-commands.yaml` key in ConfigMap. The given file needs to be in proper format describing commands tree:

```yaml
metadata: {}
uses: "..."
with: {}
args: {}
flags: []
subCommands: []
```

**fields:**

| Name | Type | Description |
| --- | --- | --- |
| **metadata** | object | Basic information about command like name or description |
| **uses** | string | Action that will be run on command execution |
| **with** | object | Configuration pass to the run action |
| **args** | object | Command arguments definition used to overwrite values in the config |
| **flags** | array | Command flags definition used to overwrite values in the config |
| **subCommands** | array | List of sub-commands. Every sub-command has the same schema as its parent |

[Here](https://github.com/kyma-project/serverless/blob/main/config/serverless/files/kyma-commands.yaml) is the example of the Serverless module extension `kyma-commands.yaml`.
[Here](https://github.com/kyma-project/cli/blob/main/internal/extensions/types/types.go) is in-code definition of types.

### metadata

This is the only required field and contains basic informations about built command:

```yaml
metadata:
  name: function
  description: A set of commands for managing Functions
  descriptionLong: Use this command to manage Functions.
```

**fields:**

| Name | Type | Description |
| --- | --- | --- |
| **name** | string | The name of the command |
| **description** | string | Short description displayd in the parent's command help |
| **description** | string | Description displayed in the command's help |

### uses

The `uses` field is based on GitHub Actions nomenclature and represents ID of the action that will be run on every command execution. This field is not required and if it's empty, then command works as command grouping all sub-commands (non executable parent of all sub-commands):

```yaml
uses: resource_get
```

All available actions descriptions can be found [here](actions.md).

### with

This field contains action specific configuration. This field supports [go templates](https://pkg.go.dev/text/template), with the `$` prefix, that can be used to dynamicly pass right values from args or flags. The example:

```yaml
uses: resource_delete
with:
  resource:
    apiVersion: serverless.kyma-project.io/v1alpha2
    kind: Function
    metadata:
      name: ${{ .args.value }}
      namespace: ${{ .flags.namespace.value }}
```

Configuration under the `with` field is action specific and its scheme depends on the used action.

### flags & args

Arguments and flags are the only way to get inputs from the end user and pass them to the config under the `with` field.

More information about flags and arguments can be found [here](inputs.md).

### subCommands

Kyma extensions has tree structure and it means that every command can work as node or leaf. The `subCommands` array contains full list of child commands. Every sub-command is in the same format as its parent and has own `subCommands`:

```yaml
metadata:
  name: cmd1
subCommands:
- name: cmd2
  subCommands:
  - name: cmd3
- name: cmd4
  subCommands:
  - name: cmd5
  - name: cmd6
```

The yaml abouve will build extension with the following commands tree:

```text
─ cmd1
  ├ cmd2
  | └ cmd3
  └ cmd4
    ├ cmd5
    └ cmd6
```

## Extension standards

Kyma CLI provides basic fields validation only but extension owner is responsible for its quality. Here is list of standards every well preprared extension should meet:

- the `.metadata` should have field all `.metadata.name`, `.metadata.description` and `.metadata.descriptionLong` fields
- the `.metadata.name` should describes possible argument and flags. For example:
  - `name: "get [<resource_name>] [flags]"` - optional resource name argument and possible flags
  - `name: "get <resource_name> [flags]"` - required resource name argument and possible flags
  - `name: "get [flags]"` - possible flags and no args
- the `.metadata.description` should start with a capital letter
- the `.metadata.descriptionLong` should start with a capital letter and end with a dot
- every `.flag[].name` should be one word or multiple words split by the `-` sign
- every `.flag[].description` should not be empty and starts with capital letter
- the `.flag[].shorthand` is optional and should be used only for the essential flags and should follow be intuitive, like shorthand `r` for `replicas` or `f` for `file`
