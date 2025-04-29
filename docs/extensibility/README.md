# Extensibility in Kyma CLI

## Overview

Like any other CLI, Kyma CLI compiles every functionality into its binary. With the Kyma CLI extensibility feature, you can create module-oriented functionality that is versioned and kept together with the module on a cluster. With this feature, you can extend the CLI with resource-oriented commands to manage your module resources or with module-oriented commands that allow for interaction with a module.

Extensions can be added by creating a ConfigMap with the expected label and data in the expected format (see [ConfigMap](./README.md#configmap)). The CLI binary with access to listing ConfigMaps on a cluster fetches all extensions from it when run and builds additional commands based on them.

All commands built from extensions can be accessed under the `kyma alpha` commands group.

## Concept

The main goal of the Kyma CLI extensibility is to support the basic Kyma resources and modules so that you can intuitively use Kyma and the CLI. Basically, Kyma CLI is designed to keep module-oriented commands as extensions. This means that the CLI that is not connected to the cluster or connected to the cluster without any module installed contains only an essential, minimalistic list of commands.

Such a solution provides control over the interaction with the module on the side of the team responsible for the module, but maintains a uniform CLI standard.

In addition, you don't need to migrate extensions on the CLI code side. If the team wants to introduce a change in the definition of a command or group of commands, they must only release a new version of the module containing the updated version of the extension. For example, the team responsible for the APIRule resource in version `v2alpha1` created an extension that allows adding and removing APIRule resources and wants to release a new version of the resource. In that case, the only thing they must do, along with adding a new version to the module, is update the extension and release a new version of the module. Also, it is possible to have different extension versions for every component and release channel.

### Concept diagram

![cli-extensibility.svg](./assets/cli-extensibility.svg)

Steps:

1. Run the CLI binary (for example `kyma alpha function create new-function`)
2. Load all extensions ConfigMaps from the cluster
3. Build new commands based on ConfigMaps
4. Execute desired command

## ConfigMap

The extension is defined and enabled with the proper ConfigMap deployed on a cluster that CLI has access to (for example, by exporting the `KUBECONFIG` env or passing the correct argument to the `--kubeconfig` flag). The ConfigMap can have any name and be located in any namespace, but must contain the `kyma-cli/extension: commands` and `kyma-cli/extension-version: v1` labels, and the `kyma-commands.yaml` data key with the correct extension configuration. For example:

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

For the example of the Serverless module extension ConfigMap, see [cli-extension.yaml](https://github.com/kyma-project/serverless/blob/main/config/serverless/templates/cli-extension.yaml).

## kyma-commands.yaml

The extension definition is represented by the YAML file inside the `kyma-commands.yaml` key in the ConfigMap. The given file must be in the proper format describing the command tree:

```yaml
metadata: {...}
uses: "..."
with: {...}
args: {...}
flags: [...]
subCommands: [...]
```

**fields:**

| Name | Required | Type | Description |
| --- | --- | --- | --- |
| **metadata** | yes | object | Basic information about the command, for example, name or description |
| **uses** | no | string | Action that is run on the command execution |
| **with** | no | object | Configuration passed to the run action |
| **args** | no | object | Command arguments definition used to overwrite values in the configuration under the `with` field |
| **flags** | no | array | Command flags definition used to overwrite values in the configuration under the `with` field |
| **subCommands** | no | array | List of sub-commands. Every sub-command has the same schema as its parent |

For the example of the Serverless module extension, see [kyma-commands.yaml](https://github.com/kyma-project/serverless/blob/main/config/serverless/files/kyma-commands.yaml).
For the in-code definition of types, see [types.go](https://github.com/kyma-project/cli/blob/main/internal/extensions/types/types.go).

### metadata

This is the only required field and contains basic information about the built command:

```yaml
metadata:
  name: function
  description: A set of commands for managing Functions
  descriptionLong: Use this command to manage Functions.
```

**fields:**

| Name | Required | Type | Description |
| --- | --- | --- | --- |
| **name** | yes | string | The name of the command |
| **description** | no | string | Short description displayed in the parent's command help |
| **descriptionLong** | no | string | Description displayed in the command's help |

### uses

The `uses` field is based on GitHub Actions nomenclature and represents the ID of the action that is run on every command execution. If it's empty, then the command works as a command grouping all sub-commands (non-executable parent of all sub-commands):

```yaml
uses: resource_get
```

For all available action descriptions, see [Actions](actions.md).

### with

This field contains action-specific configuration. It supports [Go templates](https://pkg.go.dev/text/template), with the `$` prefix, that can be used to dynamically pass the right values from args or flags. For example:

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

Configuration under the `with` field is action-specific, and its scheme depends on the used action.

### flags & args

Arguments and flags are the only way to get inputs from the end user and pass them to the config under the `with` field.

For more information about flags and arguments, see [Inputs](inputs.md).

### subCommands

Kyma extension has a tree structure, which means that every command can work as a node or a leaf. The `subCommands` array contains a full list of child commands. Every sub-command is in the same format as its parent and may have its own `subCommands`:

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

This yaml example builds the extension with the following command tree:

```text
─ cmd1
  ├ cmd2
  | └ cmd3
  └ cmd4
    ├ cmd5
    └ cmd6
```

## Extension Standards

Kyma CLI provides basic field validation only. The extension owner is responsible for its quality. The following list provides standards every well-prepared extension must meet:

| Field | Rule |
| --- | --- |
| **metadata** | It must have the `.metadata.name`, `.metadata.description`, and `.metadata.descriptionLong` fields |
| **metadata.name** | It describes possible arguments and flags. For example, `name: "get [<resource_name>] [flags]"`, `name: "delete <resource_name> [flags]"` or `name: "explain [flags]"` |
| **metadata.description** | It must start with a capital letter |
| **metadata.descriptionLong** | It must start with a capital letter and end with a dot |
| **flag[].name** | It must be one word or multiple words split by the `-` sign |
| **flag[].description** | It must not be empty and start with capital letter |
| **flag[].shorthand** | It is optional and must be used only for the essential flags. It must be intuitive, like shorthand `r` for `replicas` or `f` for `file`, etc. |
