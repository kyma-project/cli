# Actions

## Overview

Actions are the base functionality of every executable command. They determine the procedure that the command runs on execution. Every action expects a specific configuration under the `.with` field. They are designed to cover most common and generic cases, such as [CRUD operations](#available-resource-oriented-actions) (create, read, update, delete), [module-oriented operations](#available-module-oriented-actions) (like importing images to the in-cluster registry for the `docker-registry` module), and [cluster call operations](#available-cluster-call-actions) allowing to run script remotely on a cluster.

## Go Templates

Any field in the `with` object supports [go-templates](https://pkg.go.dev/text/template) with all built-in features like [functions](https://pkg.go.dev/text/template#hdr-Functions) or [actions](https://pkg.go.dev/text/template#hdr-Actions). This allows access to values from args and flags. For more information, see  [Inputs](./inputs.md#go-templates).

All templates start with the `${{` prefix and end with the `}}` suffix. For example:

```yaml
name: ${{ .flags.name.value }}
```

### Custom Functions

| Function | For Type | Description | Example |
| --- | --- | --- | --- |
| **newLineIndent** | string | Adds given indent to string if it's multiline | `source: ${{ .flags.source.value \| newLineIndent 20 }}` |
| **toEnvs** | map | Converts the input data map to an array of Kubernetes-like envs | `envs: ${{ .flags.env.value \| toEnvs }}` |
| **toArray** | map | Converts the input data map to an array in a given format. Use `{{.key}}` and `{{.value}}` to access map data | `secretMounts: ${{ .flags.secretmount.value \| toArray "{'secretName':'{{.key}}','mountPath':'{{.value}}'}" }}` |
| **toYaml** | map | Converts the input data map to an YAML object | `data: ${{ .flags.configmapdata.value \| toYaml }}` |

## Available Resource-Oriented Actions

| Name | Description |
| --- | --- |
| **resource_create** | Creates a resource in a cluster |
| **resource_get** | Gets a resource from a cluster |
| **resource_delete** | Deletes a resource from a cluster |
| **resource_explain** | Explains a resource by displaying info about it |

### resource_create

Use this action to create any cluster resource.

**Action configuration:**

```yaml
dryRun: false
output: "..."
resource: {...}
```

**Fields:**

| Name | Type | Description |
| --- | --- | --- |
| **dryRun** | bool | Simulates resource deletion if set to `true` |
| **output** | enum | Changes the output format if not empty. It can be `yaml` or `json` |
| **resource** | object | Raw object applied to a cluster |

> [!NOTE]
> For the action usage example, see [kyma-commands.yaml](https://github.com/kyma-project/serverless/blob/98b03d4d5f721564ade3e22a446c737aed17d0bf/config/serverless/files/kyma-commands.yaml#L81-L149).

### resource_get

Use this action to get or list resources of one kind from the cluster. Output resources are displayed as a kubectl-like table.

**Action configuration:**

```yaml
fromAllNamespaces: false
output: "..."
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
| **output** | enum | Changes the output format if not empty. It can be `yaml` or `json` |
| **fromAllNamespaces** | bool | Determines if resources must be taken from all namespaces |
| **resource.apiVersion** | string | Output resources ApiVersion |
| **resource.kind** | string | Output resources Kind |
| **resource.metadata.name** | string | Name of the resource to get. If empty, it gets all resources in the namespace |
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
dryRun: false
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
| **dryRun** | bool | Simulates resource deletion if set to `true` |
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

Action designed for the `docker-registry` module to import image from a local registry to the in-cluster docker registry.

**Action configuration:**

```yaml
image: "..."
```

**Fields:**

| Name | Type | Description |
| --- | --- | --- |
| **image** | string | Image from local registry to import in format \<image\>:\<tag\> |

### registry_config

Action designed for the `docker-registry` module to get in-cluster Docker Config JSON.

**Action configuration:**

```yaml
pushRegAddrOnly: false
pullRegAddrOnly: false
output: "..."
useExternal: false
```

**Fields:**

| Name | Type | Description |
| --- | --- | --- |
| **pushRegAddrOnly** | bool | Return the push registry address only |
| **pullRegAddrOnly** | bool | Return the pull registry address only |
| **output** | string | Path to the file to write output to instead of printing it in the terminal |
| **useExternal** | bool | Use external configuration instead of internal one |

### function_init

Action designed for the `serverless` module to init local workspace with hello world Functions files with handler and dependencies.

**Action configuration:**

```yaml
useRuntime: "..."
outputDir: "..."
runtimes:
  <runtime>:
    depsFilename: "..."
    depsData: "..."
    handlerFilename: "..."
    handlerData: "..."
```

**Fields:**

| Name | Type | Description |
| --- | --- | --- |
| **useRuntime** | string | Desired runtime to generate workspace for |
| **outputDir** | string | Path to the output directory where the workspace is generated |
| **runtimes** | map | Map of available runtimes from which workspace is generated by using the `useRuntime` field |
| **runtimes[\<runtime\>].depsFilename** | string | The filename of the depdencies file |
| **runtimes[\<runtime\>].depsData** | string | The output dependencies file content  |
| **runtimes[\<runtime\>].handlerFilename** | string | The filename of the handler file |
| **runtimes[\<runtime\>].handlerData** | string | The output handler file content |

## Available Cluster-Call Actions

Some functionality implemented by the resource and module oriented actions may be not enought for some more complex cases. To cover such cases it is possible to define own procedures/scripts on the module controller level, open endpoint allowing to run script and call it using the Kyma CLI.

| Name | Description |
| --- | --- |
| **cluster_call_files_to_save** | Call the container for a list of files and save them on a machine |

**Server error handling:**

Every server can return two types of responses: error or data. Data is something expected and specific to the action. An error is a JSON response with the `error` field, which is allowed for every cluster-call action and is expected when the response status is greater than 400. Error response structure:

```json
{
  "error": "..."
}
```

### cluster_call_files_to_save

This action sends request to the server on a cluster for a list of files and next saves them locally on a machine.

**Action configuration:**

```yaml
outputDir: "..."
request:
  parameters: {...}
targetPod:
  path: "..."
  port: "..."
  namespace: "..."
  selector: {...}
```

**Fields:**

| Name | Type | Description |
| --- | --- | --- |
| **outputDir** | string | Path to the output directory where the workspace is generated |
| **request.parameters** | string | Additional parameters passed to the request |
| **targetPod.path** | string | Target server path |
| **targetPod.port** | string | Target server port |
| **targetPod.namespace** | string | Target pod namespace |
| **targetPod.selector** | string | Target pod label selector (same as k8s [selector concept](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors)) |

### Server Response

Action expects that data response will contain the status code `204` and JSON data type in format:

```json
{
  "files": [
    {
      "name": "...",
      "data": "..."
    },
    ...
  ]
}
```

| Name | Description |
| --- | --- |
| **files** | List of the output files to write on a machine |
| **files[].name** | Name of the file (may contain directories like `bin/readme.md`) |
| **files[].data** | Encoded by base64 file content |
