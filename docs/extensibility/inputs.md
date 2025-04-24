# Arguments and flags

Arguments and flags are the only way to get inputs from the end user and pass them to the config under the `with` field.

Example flags and args:

```yaml
...
args:
  type: string
  optional: true
flags:
- type: int
  name: "replicas"
  shorthand: "r"
  description: "Function's replicas"
  default: "1"
- type: path
  name: "source-file"
  default: |
    example default
    file
    body
- type: bool
  name: is-ok
- type: map
  name: env
```

**args fields:**

| Name | Type | Description |
| --- | --- | --- |
| **type** | string | Flag input type |
| **optional** | bool | Can be set to true if argument is not required |

The `type` field is the only required one to configure arguments.

**flags fields:**

| Name | Type | Description |
| --- | --- | --- |
| **type** | string | Flag input type |
| **name** | string | Name of the flag |
| **shorthand** | string | One letter shorthand of the flag |
| **description** | string | Description of the flags |
| **default** | string | Default value of the flag |
| **required** | bool | Can be set to true if flag is required |

The `type` and the `name` fields are the only ones required.

## type

The `.type` field defines variable type of argument or flags. Using type resulted in input validation, so CLI will validate that user pass int value for the `int` type.

**Possible types:**

| Name | Description |
| --- | --- |
| string | Flag in string type |
| int | Flag in int64 type |
| bool | Flag in bool type. Using flag without value results in changing its value to `true` (for example `--enable` instead of `--enable=true`) |
| path | Flag in string type which value is taken from the file pointed using the flag. The `.default` field defines default value for the flag, not default path to file |
| map | Flags in map type allowing user to pass many flags in format KEY=VALUE. This type can be used for example to collect envs from user by passing following input `command --env MY_ENV=MY_VALUE --env MY_ENV_2=MY_VALUE_2` |

## go templates

Flags and args values can be in the `with` field using go-templates. Afer command execution CLI will collect all inputs ald build the following data structure:

```yaml
args:
  type: "..."
  optional: false
  value: "..."
flags:
  <flagname>:
    type: "..."
    name: "..."
    shorthand: "..."
    description: "..."
    default: "..."
    value: "..."
```

**fields:**

| Name | Type | Description |
| --- | --- | --- |
| **args** | object | Arguments data |
| **args.type** | string | Type of the arguments taken from the extension definition |
| **args.optional** | bool | Determines if argument can be omitted. It's taken from the extension definition |
| **args.value** | string | Value of the argument |
| **flags** | map | Map of the commands flags. Map keys are build based on flag's name but without `-` signs (for example flag `--all-namespaces` will be represented in map as `.flags.allnamespaces` field) |
| **flags[\<flagname\>].type** | string | Type of the flag taken from the extension definition |
| **flags[\<flagname\>].name** | string | Name of the flag taken from the extension definition |
| **flags[\<flagname\>].shorthand** | string | Shorthand of the flag taken from the extension definition |
| **flags[\<flagname\>].description** | string | Description of the flag taken from the extension definition |
| **flags[\<flagname\>].default** | string | Default value of the flag taken from the extension definition |
| **flags[\<flagname\>].value** | string | Value of the flag. If flag was not set it contains default value |

### example

Flags and arguments can be used by calling right value from the structure in format described above. For example for the `resource_create` action we can overwrite configuration:

```yaml
metadata:
  name: create <resource_name> [flags]
  description: "Create resources"
  descriptionLong: "Use this command to create resources on a cluster."
uses: resource_create
args:
  type: string
flags:
- type: string
  name: namespace
  shorthand: "n"
  description: "Resource namespace"
  default: "default"
- type: int
  name: "replicas"
  description: "Function replicas"
  default: "1"
- type: path
  name: "source"
  description: "Function source file path"
  shorthand: "s"
  default: |
    module.exports = {
      main: function(event, context) {
        return 'Hello World!'
      }
    }
- type: map
  name: "env"
  description: "Function environment variables in format key=value"
with:
  resource:
    apiVersion: serverless.kyma-project.io/v1alpha2
    kind: Function
    metadata:
      name: ${{ .args.value }}
      namespace: ${{ .flags.namespace.value }}
    spec:
      runtime: nodejs22
      replicas: ${{ .flags.replicas.value }}
      env: ${{ .flags.env.value | toEnvs }}
      source:
        inline:
          source: |
            ${{ .flags.source.value | newLineIndent 20 }}
```
