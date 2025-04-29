# Create Extension

This article provides you through process of preparing own extension responsible for ConfigMap management. This show-case extension will provide following functionalities:

* Get ConfigMap from a cluster
* Create ConfigMap with given name, namespace and data
* Delete ConfigMap based on its name and namespace

The extension will provides main command (command group) `configmap` that will do nothing instead of printing the `help` on execution, but will has three sub-commands (`create`, `get`, `delete`) with resource-oriented actions described in the list above.

## Steps

1. Prepare ConfigMap with root command

    The very first step will be preparing [a ConfigMap](./README.md#configmap) with required labels and data. For this usecase we would like to have root command `configmap` without any action performed on execution. Let's create ConfigMap with such command and description, following [extensions standards](./README.md#extension-standards), what this command will do:

    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: my-extension
      labels:
        kyma-cli/extension: commands
        kyma-cli/extension-version: v1
    data:
      kyma-commands.yaml: |-
        metadata:
          name: configmap
          description: "Manage ConfigMap resources"
          descriptionLong: "Use this command to manage ConfigMap resources."
    ```

    After applying ConfigMap on a cluster we can use the Kyma CLI to validate that extension is visible:

    ```bash
    $ kyma alpha configmap

    Use this command to manage ConfigMap resources.

    Usage:
    kyma alpha configmap [flags]

    Global Flags:
    -h, --help                    Help for the command
        --kubeconfig string       Path to the Kyma kubeconfig file
        --show-extensions-error   Prints a possible error when fetching extensions fails
        --skip-extensions         Skip fetching extensions from the cluster
    ```

2. Support the ConfigMap `create` command

    Now, when the extension base is working we can focus on preparing first functionality of creating ConfigMap. In the very first step we will create empty ConfigMap with no data field.

    To do so we need to use [the resource_create action](./actions.md#resource_create) and define its configuration under the `with` field:

    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: my-extension
      labels:
        kyma-cli/extension: commands
        kyma-cli/extension-version: v1
    data:
      kyma-commands.yaml: |-
        metadata:
          name: configmap
          description: "Manage ConfigMap resources"
          descriptionLong: "Use this command to manage ConfigMap resources."
        subCommands:
        - metadata:
            name: create
            description: "Create ConfigMap resource"
            descriptionLong: "Use this command to create ConfigMap resource."
          uses: resource_create
          with:
            resource:
              apiVersion: v1
              kind: ConfigMap
              metadata:
                name: cm-from-extension
                namespace: default
    ```

    We can apply new extension version and check if it works:

    ```bash
    $ kyma alpha configmap create

    resource cm-from-extension applied
    ```

3. Extend the `create` command with resource-oriented features

    It would be much easier for an end-user to create ConfigMap with defined name, namespace and data. In this section we will use [flags and args](./inputs.md#arguments-and-flags) to collect these data from user and pass them to the `resource_create` action using configuration under the `with` field using the [go tempaltes](./inputs.md#go-templates) and it's available [custom functions](./actions.md#custom-functions):

    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: my-extension
      labels:
        kyma-cli/extension: commands
        kyma-cli/extension-version: v1
    data:
      kyma-commands.yaml: |-
        metadata:
          name: configmap
          description: "Manage ConfigMap resources"
          descriptionLong: "Use this command to manage ConfigMap resources."
        subCommands:
        - metadata:
            name: create <resource_name> [flags]
            description: "Create ConfigMap resource"
            descriptionLong: "Use this command to create ConfigMap resource."
          uses: resource_create
          args:
            type: string
          flags:
          - name: "namespace"
            description: "ConfigMap's namespace"
            shorthand: "n"
            type: string
            default: "default"
          - name: "from-literal"
            description: "Data element in format <KEY>=<VALUE>"
            type: map
          with:
            resource:
              apiVersion: v1
              kind: ConfigMap
              metadata:
                name: ${{ .args.value }}
                namespace: ${{ .flags.namespace.value }}
              data: ${{ .flags.fromliteral.value | toYaml }}
    ```

    Important thing is that in this case we are building the `--from-literal` flag that has `map` type. This allows to set this flag many times to collect more than one data, but it needs additional conversion to array using [the toYaml function](./actions.md#custom-functions). Also the `.metadata.name` is updated because command got new flags and args (following [quality standards](./README.md#extension-standards)). Let's apply new version and then test it:

    ```bash
    $ kyma alpha configmap create cm-from-extension --namespace default --from-literal data1=value1 --from-literal data2=value2

    resource cm-from-extension applied
    ```

    Using the `kubectl` we can check if ConfigMap has all expected fields:

    ```bash
    $ kubectl get configmap cm-from-extension -oyaml

    apiVersion: v1
    data:
      data1: value1
      data2: value2
    kind: ConfigMap
    metadata:
      creationTimestamp: "2025-04-29T11:18:25Z"
      name: cm-from-extension
      namespace: default
      resourceVersion: "406306"
      uid: 0ace84cc-a057-4141-b0da-bc6d3f1249a7
    ```

4. Add the kubectl-like `get` command

    [The resource_get action](./actions.md#resource_get) allows to display requested resources in a kubectl-like table view with one custom column that counts the data length (using the JQ expression). Our command will work in a few modes depending on the given argument/flags:

    * `kyma alpha configmap get` - get all ConfigMaps from the default namepsace (default value for the `namespace` flag)
    * `kyma alpha configmap get <resource_name>` - get only the ConfigMap with the given name
    * `kyma alpha configMap get --all-namespaces` - get all ConfigMaps from all namespaces

    Let's add another sub-command with such functionality:

    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: my-extension
      labels:
        kyma-cli/extension: commands
        kyma-cli/extension-version: v1
    data:
      kyma-commands.yaml: |-
        metadata:
          name: configmap
          description: "Manage ConfigMap resources"
          descriptionLong: "Use this command to manage ConfigMap resources."
        subCommands:
        - metadata:
            name: create <resource_name> [flags]
            description: "Create ConfigMap resource"
            descriptionLong: "Use this command to create ConfigMap resource."
          uses: resource_create
          args:
            type: string
          flags:
          - name: "namespace"
            description: "ConfigMap's namespace"
            shorthand: "n"
            type: string
            default: "default"
          - name: "from-literal"
            description: "Data element in format <KEY>=<VALUE>"
            type: map
          with:
            resource:
              apiVersion: v1
              kind: ConfigMap
              metadata:
                name: ${{ .args.value }}
                namespace: ${{ .flags.namespace.value }}
              data: ${{ .flags.fromliteral.value | toYaml }}
        - metadata:
            name: get [<resource_name>] [flags]
            description: "Get ConfigMap resource"
            descriptionLong: "Use this command to get ConfigMap resource."
          uses: resource_get
          args:
            type: string
            optional: true
          flags:
          - name: "namespace"
            description: "ConfigMap's namespace"
            shorthand: "n"
            type: string
            default: "default"
          - name: "all-namespaces"
            description: "Get resources from all namespaces"
            type: bool
            shorthand: "A"
          with:
            fromAllNamespaces: ${{.flags.allnamespaces.value}}
            resource:
              apiVersion: v1
              kind: ConfigMap
              metadata:
                name: ${{.args.value}}
                namespace: ${{.flags.namespace.value}}
            outputParameters:
            - resourcePath: '.data | length'
              name: "data length"
    ```

    After applying the Extension ConfigMap we can test it using:

    ```bash
    $ kyma alpha configmap get
    
    NAME                    DATA LENGTH
    cm-from-extension       2
    my-extension            1
    ```

5. Allow ConfigMap deletion

    To cover all basic operations on a ConfigMap resource we would like to implement [the resource_delete action](./actions.md#resource_delete) allowing end-users to delete ConfigMap resources. Such command will receive one required argument (resource name) and one optional flag (`--namespace`):

    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: my-extension
      labels:
        kyma-cli/extension: commands
        kyma-cli/extension-version: v1
    data:
      kyma-commands.yaml: |-
        metadata:
          name: configmap
          description: "Manage ConfigMap resources"
          descriptionLong: "Use this command to manage ConfigMap resources."
        subCommands:
        - metadata:
            name: create <resource_name> [flags]
            description: "Create ConfigMap resource"
            descriptionLong: "Use this command to create ConfigMap resource."
          uses: resource_create
          args:
            type: string
          flags:
          - name: "namespace"
            description: "ConfigMap's namespace"
            shorthand: "n"
            type: string
            default: "default"
          - name: "from-literal"
            description: "Data element in format <KEY>=<VALUE>"
            type: map
          with:
            resource:
              apiVersion: v1
              kind: ConfigMap
              metadata:
                name: ${{ .args.value }}
                namespace: ${{ .flags.namespace.value }}
              data: ${{ .flags.fromliteral.value | toYaml }}
        - metadata:
            name: get [<resource_name>] [flags]
            description: "Get ConfigMap resource"
            descriptionLong: "Use this command to get ConfigMap resource."
          uses: resource_get
          args:
            type: string
            optional: true
          flags:
          - name: "namespace"
            description: "ConfigMap's namespace"
            shorthand: "n"
            type: string
            default: "default"
          - name: "all-namespaces"
            description: "Get resources from all namespaces"
            type: bool
            shorthand: "A"
          with:
            fromAllNamespaces: ${{.flags.allnamespaces.value}}
            resource:
              apiVersion: v1
              kind: ConfigMap
              metadata:
                name: ${{.args.value}}
                namespace: ${{.flags.namespace.value}}
            outputParameters:
            - resourcePath: '.data | length'
              name: "data length"
        - metadata:
            name: delete <resource_name> [flags]
            description: "Delete ConfigMap resource"
            descriptionLong: "Use this command to delete ConfigMap resource."
          uses: resource_delete
          args:
            type: string
          flags:
          - name: "namespace"
            description: "ConfigMap's namespace"
            shorthand: "n"
            type: string
            default: "default"
          with:
            resource:
              apiVersion: v1
              kind: ConfigMap
              metadata:
                name: ${{ .args.value }}
                namespace: ${{ .flags.namespace.value }}
    ```

    The new applyed extension version allows us to delete previously created ConfigMap:

    ```bash
    $ kyma alpha configmap delete cm-from-extension

    resource cm-from-extension deleted
    ```

    To verify that ConfigMap is removed we can use the `kyma alpha configmap get` command:

    ```bash
    $ kyma alpha configmap get
    
    NAME                    DATA LENGTH
    my-extension            1
    ```
