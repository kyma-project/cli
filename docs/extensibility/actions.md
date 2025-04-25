# Actions

## Overview

Actions are the base functionality of every executable command. They determine the procedure that the command runs on execution. Every action expects a specific configuration under the `.with` field. They are designed to cover most common and generic cases, such as CRUD operations (create, read, update, delete) and module-oriented operations (like importing images to the in-cluster registry for the `docker-registry` module).

> [!NOTE] 
> For the in-code definition, see [alpha.go](https://github.com/kyma-project/cli/blob/main/internal/cmd/alpha/alpha.go#L35).

## Go Templates

Any field in the `with` object supports [go-templates](https://pkg.go.dev/text/template) with all built-in features like [functions](https://pkg.go.dev/text/template#hdr-Functions) or [actions](https://pkg.go.dev/text/template#hdr-Actions). This allows access to values from args and flags. For more information, see  [Inputs](./inputs.md#go-templates).

All templates start with the `${{` prefix and end with the `}}` suffix. For example:

```yaml
name: ${{ .flags.name.value }}
```

### Custom Functions


| Function | Description | Example |
| --- | --- | --- |
| **newLineIndent** | Adds given indent to string if it's multiline | `source: ${{ .flags.source.value \| newLineIndent 20 }}` |
| **toEnvs** | Converts the input data map to array of Kubernetes-like envs | `envs: ${{ .flags.env.value \| toEnvs }}` |
| **toArray** | Converts the input data map to an array in a given format. Use `{{.key}}` and `{{.value}}` to access map data | `secretMounts: ${{ .flags.secretmount.value \| toArray "{'secretName':'{{.key}}','mountPath':'{{.value}}'}" }}` |

## Available Resource-Oriented Actions

| Name | Description |
| --- | --- |
| **resource_create** | Creates a resource on a cluster |
| **resource_get** | Gets a resource from a cluster |
| **resource_delete** | Deletes a resource from a cluster |
| **resource_explain** | Explains a resource by displaying info about it |

### resource_create

Use this action to create any resource on the cluster.

**Action configuration:**

```yaml
resource: {}
```

**Fields:**

| Name | Type | Description |
| --- | --- | --- |
| **resource** | object | Raw object that must be applied to the cluster |

> [!NOTE]
> For the action usage example, see [kyma-commands.yaml](https://github.com/kyma-project/serverless/blob/98b03d4d5f721564ade3e22a446c737aed17d0bf/config/serverless/files/kyma-commands.yaml#L81-L149).

### resource_get

Use this action to get or list resources of one kind from the cluster. Output resources are displayed as a kubectl-like table.

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
| **fromAllNamespaces** | bool | Determines if resources must be taken from all namespaces |
| **resource.apiVersion** | string | Output resources ApiVersion |
| **resource.kind** | string | Output resources Kind |
| **resource.metadata.name** | string | Name of the resource to get. If empty, it gets all resources in the namespace. |
| **resource.metadata.namespace** | string | Namespace from which resources are obtained |
| **outputParameters[]** | array | List of additional parameters displayed in the table view |
| **outputParameters[].name** | string | Additional column name |
| **outputParameters[].resourcePath** | string | Path in the resource from which the value is obtained. Supports the [JQ](https://jqlang.org/) language |

> [!NOTE]
> For the action usage example, see [kyma-commands.yaml](https://github.com/kyma-project/serverless/blob/98b03d4d5f721564ade3e22a446c737aed17d0bf/config/serverless/files/kyma-commands.yaml#L7-L43).

### resource_delete

Use this action to delete a resource from the cluster.

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

> [!NOTE]
> For the action usage example, see [kyma-commands.yaml](https://github.com/kyma-project/serverless/blob/98b03d4d5f721564ade3e22a446c737aed17d0bf/config/serverless/files/kyma-commands.yaml#L61-L79).

### resource_explain

Use this action to display an explanatory note about the resource.

**Action configuration:**

```yaml
output: "..."
```

**Fields:**

| Name | Type | Description |
| --- | --- | --- |
| **output** | string | Note to display |

> [!NOTE]
> For the action usage example, see [kyma-commands.yaml](https://github.com/kyma-project/serverless/blob/98b03d4d5f721564ade3e22a446c737aed17d0bf/config/serverless/files/kyma-commands.yaml#L45-L58).

## Available Module-Oriented Actions

| Name | Module | Description |
| --- | --- | --- |
| **registry_image_import** | `docker-registry` | Import image from local registry to in-cluster Docker registry |
| **registry_config** | `docker-registry` | Get in-cluster registry configuration |
| **function_init** | `serverless` | Generate Function's source and dependencies locally |

### registry_image_import

TODO

### registry_config

TODO

### function_init

TODO
