---
title: Kyma CLI command usage examples
type: Details
---

The following examples show how to provision a cluster, install Kyma, and run the tests.

## Provision a cluster locally or using cloud providers

To provision a cluster on a specific cloud provider (in this example GCP), run:

```bash
kyma provision gke --credentials {SERVICE_ACCOUNT_KEY_FILE_PATH} --name {CLUSTER_NAME} --project {GCP_PROJECT} 
```

To provision a Minikube cluster, run:

```bash
kyma provision minikube
```

## Install Kyma

To install Kyma using your own domain, run:

```bash
kyma install --domain {DOMAIN} --tls-cert {TLS_CERT} --tls-key {TLS_KEY}
```
- `{TLS_CERT}` and `{TLS_KEY}` are the TLS certificate and TLS key for the domain used. Note that both of them must be encoded in base64. You can use [OpenSSL](https://www.openssl.org/) (which is installed by default with macOS) to convert your TLS certificate and TLS key to base64-encoded values.
- To simplify the process:
  - You can first encode the TLS certificate and TLS key and export them as environment variables:
  - ```bash
    export TLS_CERT="$(openssl base64 -in {TLS_CERT_FILE_PATH})"
    export TLS_KEY="$(openssl base64 -in {TLS_KEY_FILE_PATH})"
    ```
  - `{TLS_CERT_FILE_PATH}` is the path to the file containing the TLS certificate and `{TLS_KEY_FILE_PATH}` is the path to the file containing the TLS key.
  - Now, you can use these environment variables to install Kyma with your own domain `{DOMAIN}` as shown below:
  - ```bash
    kyma install --domain {DOMAIN} --tls-cert $TLS_CERT --tls-key $TLS_KEY
    ```

To install Kyma from the `master` branch, run:

```bash
kyma install --source master
```

To install Kyma using your own Kyma installer image, run:

```bash
kyma install --source {IMAGE}
```

To build an image from your local sources and install Kyma on a remote cluster based on this image, run:

```bash
kyma install --source local --custom-image {IMAGE}
```
- For example, if your `{IMAGE}` is `user/my-kyma-installer:v1.4.0`, then this command will build an image from your local sources and push it your repository `user/my-kyma-installer` with the tag `v1.4.0`. Then, this image `user/my-kyma-installer:v1.4.0` will be used in the `Deployment` for `kyma-installer`.

To install kyma with only specific components, run:

```bash
kyma install --components {COMPONENTS_FILE_PATH}
```
- `{COMPONENTS_FILE_PATH}` is the path to a YAML file containing the desired component list to be installed such as the one below:
- ```bash
  components:
  - name: "cluster-essentials"
    namespace: "kyma-system"
  - name: "testing"
    namespace: "kyma-system"
  - name: "istio"
    namespace: "istio-system"
  - name: "xip-patch"
    namespace: "kyma-installer"
  - name: "istio-kyma-patch"
    namespace: "istio-system"
  - name: "dex"
    namespace: "kyma-system"
  ```
- In this example, only these 6 components will be installed on the cluster.

To override the standard Kyma installation, run:
```bash
kyma install --override {OVERRIDE_FILE_PATH}
```
- `{OVERRIDE_FILE_PATH}` is the path to a YAML file containing the parameters to override such as the one below:
- ```bash
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: ory-overrides
    namespace: kyma-installer
    labels:
      installer: overrides
      component: ory
      kyma-project.io/installation: ""
  data:
    hydra.deployment.resources.limits.cpu: "153m"
    hydra.deployment.resources.requests.cpu: "53m"
  ---
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: monitoring-overrides
    namespace: kyma-installer
    labels:
      installer: overrides
      component: monitoring
      kyma-project.io/installation: ""
  data:
    alertmanager.alertmanagerSpec.resources.limits.memory: "304Mi"
    alertmanager.alertmanagerSpec.resources.requests.memory: "204Mi"
  ```
- In this example, overrides are provided for 2 different components: `ory` and `monitoring`. For `ory`, the values of `hydra.deployment.resources.limits.cpu` and `hydra.deployment.resources.requests.cpu` will be overriden to `153m` and `53m` respectively. For `monitoring`, the values of `alertmanager.alertmanagerSpec.resources.limits.memory` and `alertmanager.alertmanagerSpec.resources.requests.memory` will be overriden to `304Mi` and `204Mi` respectively.

- It is also possible to provide multiple override files at the same time
  ```bash
  kyma install --override {OVERRIDEL_FILE_1_PATH} --override {OVERRIDE_FILE_2_PATH}
  ```

## Upgrade Kyma      

To upgrade the Kyma version on the cluster, you can run the `upgrade` command which has the same structure and same flags as the `install` command.

## Test Kyma

To check which test definitions are deployed on the cluster, run:

```bash
kyma test definitions
```

To run all the tests, run:

```bash
kyma test run
```

To check the test results, run:

```bash
kyma test status
```
