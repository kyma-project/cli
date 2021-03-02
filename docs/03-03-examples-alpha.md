---
title: Kyma CLI alpha command usage examples
type: Details
---

The following examples show how to use the alpha commands to provision a cluster, install and upgrade Kyma, and handle errors.

## Provision a cluster
To provision a k3s cluster, run:

```
kyma alpha provision k3s 
```
If you want to define the name of your k3s cluster and pass arguments to the Kubernetes API server (for example, to log to stderr), run:

```
kyma alpha provision k3s --name='custom_name' --server-args='--alsologtostderr'
```


## Install Kyma

There are several ways to install Kyma:

- To install Kyma, simply use the `deploy` command without any flags, and Kyma provides a default domain. 
For example, if you install Kyma on a local cluster, the default URL is `https://console.local.kyma.dev`.

  ```
  kyma alpha deploy 
  ```

- To install Kyma using your own domain name, you must provide the certificate and key as files. 
If you don't have a certificate yet, you can create a self-signed certificate and key:

  ```
  openssl req -x509 -newkey rsa:4096 -keyout key.pem -out crt.pem -days 365
  ```

  When prompted, provide your credentials, such as your name and your domain (as wildcard: `*.$DOMAIN`).

  Then, pass the certificate files to the deploy command:

  ```
  kyma alpha deploy --domain {DOMAIN} --tls-cert crt.pem --tls-key key.pem
  ```

- Optionally, you can specify from which source you want to deploy Kyma, such as the `master` branch, a specific PR, or a release version. For more details, see the documentation for the `alpha deploy` command.<br>
For example, to install Kyma from a specific version, such as `1.19.1`, run:

  ```
  kyma alpha deploy --source=1.19.1
  ```

- Alternatively, to build Kyma from your local sources and deploy it on a remote cluster, run:

  ```
  kyma alpha deploy --source=local
  ```
  > **NOTE:** By default, Kyma expects to find the sources in the `./workspace folder`. To adjust the path, set the `-w ${PATH_TO_KYMA_SOURCES}` parameter.

- To deploy Kyma with only specific components, run:

  ```
  kyma alpha deploy --components {COMPONENTS_FILE_PATH}
  ```

  `{COMPONENTS_FILE_PATH}` is the path to a YAML file containing the desired component list to be installed. In the following example, only these six components are deployed on the cluster:

  ```
  defaultNamespace: kyma-system
  prerequisites:
    - name: "cluster-essentials"
    - name: "istio"
          namespace: "istio-system"
  components:
    - name: "testing"
    - name: "xip-patch"
    - name: "istio-kyma-patch"
    - name: "dex"
  ```

- You can also install Kyma with different configuration values than the default settings. For details, see [Change Kyma settings](#change-kyma-settings).

## Upgrade Kyma

The `alpha deploy` command not only installs Kyma, you also use it to upgrade the Kyma version on the cluster. You have the same options as described under [Install Kyma](#install-kyma).

> **NOTE:** If you upgrade from one Kyma release to a newer one, an automatic compatibility check compares your current version and the new release.<br>
The compatibility check only works with release versions. If you installed Kyma from a PR, branch, revision, or local version, the compatibility cannot be checked.


## Change Kyma settings

To change your Kyma configuration, use the `alpha deploy` command and deploy the same Kyma version that you're currently using, just with different settings.

You can use the `--values-file` and the `--value` flag.

- To override the standard Kyma configuration, run:

  ```
  kyma alpha deploy --values-file {VALUES_FILE_PATH}
  ```

  In the following example, `{VALUES_FILE_PATH}` is the path to a YAML file containing the desired configuration:

  - For `ory`, the values of `hydra.deployment.resources.limits.cpu` and `hydra.deployment.resources.requests.cpu` will be overriden to `153m` and `53m` respectively.
    
  - For `monitoring`, the values of `alertmanager.alertmanagerSpec.resources.limits.memory` and `alertmanager.alertmanagerSpec.resources.requests.memory` will be overriden to `304Mi` and `204Mi` respectively.
  
  ```
  ory:
    hydra:
      deployment:
        resources:
          limits:
            cpu: 153m
          requests:
            cpu: 53m
  monitoring:
    alertmanager:
      alertmanagerSpec:
        resources:
          limits:
            memory: 304Mi
          requests:
            memory: 204Mi
  ```

- You can also provide multiple values files at the same time:

  ```
  kyma deploy --values-file {VALUES_FILE_1_PATH} --values-file {VALUES_FILE_2_PATH}
  ```

- Alternatively, you can specify single values instead of a file:

  ```
  kyma deploy --value ory.hydra.deployment.resources.limits.cpu=153m \
  --value ory.hydra.deployment.resources.requests.cpu=53m \
  --value monitoring.alertmanager.alertmanagerSpec.resources.limits.memory=304Mi \
  --value monitoring.alertmanager.alertmanagerSpec.resources.requests.memory=204Mi
  ```

## Debugging

The alpha commands support error handling in several ways, for example:

- To tweak the values on a component level, use `alpha deploy --components` to find out the settings that work for your installation.
- To understand which component failed during deployment, deactivate the `--atomic` flag for `alpha deploy`. <br>With active atomic deployment, any component that hasn't been installed successfully is rolled back, which may make it hard to find out what went wrong.

<!-- ANY OTHER DEBUGGING USE CASES? -->
