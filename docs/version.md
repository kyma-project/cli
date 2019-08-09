# version

## Description

Use this command to get the Kyma CLI version and the version of the Kyma cluster the current KUBECONFIG points to.

## Usage 

```bash
kyma version [OPTIONS]
```

## Options

| Name     | Short name | Default value| Description|
| ----------|---------|-----|------|
| --client | -c |`false`|Returns only the Kyma CLI client version. You don't need a valid KUBECONFIG to get the client version.|


## Examples

The following examples include the most common uses of the `version` command. 
* Return the version of the Kyma CLI and the Kyma cluster:
   ```bash
   kyma version 
   ```
* Return only the version of the Kyma CLI:
   ```bash
   kyma version --client
   ```
