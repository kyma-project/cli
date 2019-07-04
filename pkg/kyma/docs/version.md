# version

## Description

Prints the version of the Kyma CLI and the version of the Kyma cluster the current KUBECONFIG points to.

## Usage 

```bash
kyma version [OPTIONS]
```

## Options

| Name     | Short Name | Default value| Description|
| ----------|---------|-----|------|
| --client | -c |false|Print out the client only. No server is required to get the client version.|


## Examples

The following examples include the most common cases of using the install command. 
1. Print out the version of Kyma CLI and the Kyma cluster:
   ```bash
   kyma version 
   ```
2. Print out the version of Kyma CLI only
   ```bash
   kyma version --client
   ```
