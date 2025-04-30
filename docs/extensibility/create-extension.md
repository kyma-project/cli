# Create Extension

Learn how to prepare your own extension for ConfigMap management. This showcase extension provides the following functionalities:

* Getting ConfigMap from a cluster
* Creating ConfigMap with the given name, namespace, and data
* Deleting ConfigMap based on its name and namespace

The extension provides the main command (command group) `configmap`, which prints `help` on execution. It has three subcommands (`create`, `get`, `delete`) with resource-oriented actions described in the list above.

## Steps

1. Prepare ConfigMap with the root command

    With this step, you create [ConfigMap](./README.md#configmap) with required labels and data. For this use case, you need the root command `configmap` without any action performed on execution. Create ConfigMap with such a command and description, following [extensions standards](./README.md#extension-standards):

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
          name: configmap [flags]
          description: "Manage ConfigMap resources"
          descriptionLong: "Use this command to manage ConfigMap resources."
    ```

2. Use Kyma CLI to validate that the extension is applied:

    ```bash
    kyma alpha configmap
    ```

    You should see the following result:

    ```bash
    Use this command to manage ConfigMap resources.

    Usage:
    kyma alpha configmap [flags]

    Global Flags:
    -h, --help                    Help for the command
        --kubeconfig string       Path to the Kyma kubeconfig file
        --show-extensions-error   Prints a possible error when fetching extensions fails
        --skip-extensions         Skip fetching extensions from the cluster
    ```

3. Support the ConfigMap `create` command

    Create an empty ConfigMap with no data field using [the resource_create action](./actions.md#resource_create) and define its configuration under the `with` field:

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
          name: configmap [flags]
          description: "Manage ConfigMap resources"
          descriptionLong: "Use this command to manage ConfigMap resources."
        subCommands:
        - metadata:
            name: create [flags]
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

4. Apply the new extension version and check if it works:

    ```bash
    kyma alpha configmap create
    ```

    You should see the following result:

    ```bash
    resource cm-from-extension applied
    ```

5. Extend the `create` command with resource-oriented features

    Use [flags and args](./inputs.md#arguments-and-flags) to collect the defined name, namespace, and data from the user and pass them to the `resource_create` action using configuration under the `with` field using [Go templates](./inputs.md#go-templates) and the available [custom functions](./actions.md#custom-functions):

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
          name: configmap [flags]
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
            description: "ConfigMap namespace"
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

    > [!NOTE]
    > In this case, we are building the `--from-literal` flag with the `map` type. With this, you can set this flag many times to collect more than one piece of data, but it requires additional conversion to an array using the [toYaml function](./actions.md#custom-functions). Also, the `.metadata.name` is updated because the command got new flags and args (following [quality standards](./README.md#extension-standards)).

6. Apply the new version and test it:

    ```bash
    kyma alpha configmap create cm-from-extension --namespace default --from-literal data1=value1 --from-literal data2=value2
    ```

    You should see the following result:

    ```bash
    resource cm-from-extension applied
    ```

7. Use kubectl to check if the ConfigMap has all expected fields:

    ```bash
    kubectl get configmap cm-from-extension -oyaml
    ```

    You should see the following result:

    ```bash
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

8. Add the kubectl-like `get` command

    With the [resource_get action](./actions.md#resource_get), you can display requested resources in a kubectl-like table view with one custom column that counts the data length (using the JQ expression). Our command works in a few modes depending on the given argument or flags:

    * `kyma alpha configmap get` - Gets all ConfigMaps from the default namespace (default value for the `namespace` flag)
    * `kyma alpha configmap get <resource_name>` - Gets only the ConfigMap with the given name
    * `kyma alpha configMap get --all-namespaces` - Gets all ConfigMaps from all namespaces

9. Add another subcommand with such functionality:

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
          name: configmap [flags]
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
            description: "ConfigMap namespace"
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
            description: "ConfigMap namespace"
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

10. Use Kyma CLI to test the extension:

    ```bash
    kyma alpha configmap get
    ```

    You should see the following result:

    ```bash
    NAME                    DATA LENGTH
    cm-from-extension       2
    my-extension            1
    ```

11. Provide the deletion functionality to the ConfigMap:

    Implement [the resource_delete action](./actions.md#resource_delete) to cover all basic operations on the ConfigMap resource, allowing end-users to delete ConfigMap resources. Such a command receives one required argument (resource name) and one optional flag (`--namespace`):

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
          name: configmap [flags]
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
            description: "ConfigMap namespace"
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
            description: "ConfigMap namespace"
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
            description: "ConfigMap namespace"
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

12. Now you can delete the previously created ConfigMap:

    ```bash
    kyma alpha configmap delete cm-from-extension
    ```

    You should see the following result:

    ```bash
    resource cm-from-extension deleted
    ```

13. To verify that the ConfigMap is deleted, use the `kyma alpha configmap get` command:

    ```bash
    kyma alpha configmap get
    ```

    You should see the following result:

    ```bash
    NAME                    DATA LENGTH
    my-extension            1
    ```
