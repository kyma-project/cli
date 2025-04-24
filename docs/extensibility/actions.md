# Actions

## Overview

Actions are base functionality of every executable command. It determines procedure command run on execution. Every Action expects specific configuration under the `.with` field. All of them are designed to cover most common and generic cases like CRUD operations (create, read, update, delete) and module oriented operations (like importing images to the in-cluster registry for the `docker-registry` module).

>NOTE: in-code definition can be found [here](https://github.com/kyma-project/cli/blob/main/internal/cmd/alpha/alpha.go#L35).

## Go-Templates

Any field in the `with` object supports [go-templates](https://pkg.go.dev/text/template) with all build-in features like [functions](https://pkg.go.dev/text/template#hdr-Functions) or to perform [actions](https://pkg.go.dev/text/template#hdr-Actions). This allows to access values from args and flags (read more [here](./inputs.md#go-templates)).

All templates starts with the `${{` prefix and ends with the `}}` suffix. For example:

```yaml
name: ${{ .flags.name.value }}
```

### Custom functions

It's possible to call custom CLI function allowing to make some things easier:

| Function | Description | Example |
| --- | --- | --- |
| **newLineIndent** | Adds given indent to string if it's multiline | `source: ${{ .flags.source.value \| newLineIndent 20 }}` |
| **toEnvs** | Converts map of input data to array of kubernetes-like envs | `envs: ${{ .flags.env.value \| toEnvs }}` |
| **toArray** | Converts map of input data to array in given format. Use `{{.key}}` and `{{.value}}` to access map data | `secretMounts: ${{ .flags.secretmount.value \| toArray "{'secretName':'{{.key}}','mountPath':'{{.value}}'}" }}` |

## Available resource oriented actions

| Name | Description |
| --- | --- |
| **resource_create** | Create resource on a cluster |
| **resource_get** | Get resource/s on a cluster |
| **resource_delete** | Delete resource on a cluster |
| **resource_explain** | Explain resource by displaying info aboit it |

### resource_create

This action can be used to create any resource on the cluster.

**Action configuration:**

```yaml
resource: {}
```

**Fields:**

| Name | Type | Description |
| --- | --- | --- |
| **resource** | object | Raw object that should be applied to the cluster |

>NOTE: [example](https://github.com/kyma-project/serverless/blob/98b03d4d5f721564ade3e22a446c737aed17d0bf/config/serverless/files/kyma-commands.yaml#L81-L149) usage of the action

### resource_get

This action can be used to get/list resources in one kind from the clsuter. Output resources are displayed as kubectl-like table.

**Action configuration:**

```yaml
fromAllNamespaces: false
resource:
  apiVersion: "..."
  kind: "..."
  metadata:
    name: "..."
    namespace: "..."
outputParameters:
- resourcePath: '...'
  name: "..."
```

**Fields:**

| Name | Type | Description |
| --- | --- | --- |
| **fromAllNamespaces** | bool | Determine if resources should be get from all namespaces |
| **resource.apiVersion** | string | Output resources ApiVersion |
| **resource.kind** | string | Output resources Kind |
| **resource.metadata.name** | string | Name of the resource to get. Gets all resources in the namespace if empty |
| **resource.metadata.namespace** | string | Namespace from which resources will be obtained |
| **outputParameters[]** | array | List of additional parameters displayed in the table view |
| **outputParameters[].name** | string | Additional column name |
| **outputParameters[].resourcePath** | string | Path in the resource from which the value will be obtained. Supports [JQ](https://jqlang.org/) language |

>NOTE: [example](https://github.com/kyma-project/serverless/blob/98b03d4d5f721564ade3e22a446c737aed17d0bf/config/serverless/files/kyma-commands.yaml#L7-L43) usage of the action

### resource_delete

This action can be used to delete resource from the clsuter.

**Action configuration:**

```yaml
resource:
  apiVersion: "..."
  kind: "..."
  metadata:
    name: "..."
    namespace: "..."
```

**Fields:**

| Name | Type | Description |
| --- | --- | --- |
| **resource.apiVersion** | string | Resources ApiVersion |
| **resource.kind** | string | Resources Kind |
| **resource.metadata.name** | string | Name of the resource to delete |
| **resource.metadata.namespace** | string | Namespace of the resource to delete |

>NOTE: [example](https://github.com/kyma-project/serverless/blob/98b03d4d5f721564ade3e22a446c737aed17d0bf/config/serverless/files/kyma-commands.yaml#L61-L79) usage of the action

### resource_explain

Display explanation note about the resource.

**Action configuration:**

```yaml
output: "..."
```

**Fields:**

| Name | Type | Description |
| --- | --- | --- |
| **output** | string | Note to display |

>NOTE: [example](https://github.com/kyma-project/serverless/blob/98b03d4d5f721564ade3e22a446c737aed17d0bf/config/serverless/files/kyma-commands.yaml#L45-L58) usage of the action

## Available module oriented actions

| Name | Module | Description |
| --- | --- | --- |
| **registry_image_import** | `docker-registry` | Import image from local registry to in-cluster docker registry |
| **registry_config** | `docker-registry` | Get in-cluster registry configuration |
| **function_init** | `serverless` | Generate Function's source and dependencies locally |

### registry_image_import

TODO

### registry_config

TODO

### function_init

TODO
