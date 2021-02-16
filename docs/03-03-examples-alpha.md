---
title: Kyma CLI alpha command usage examples
type: Details
---

The following examples show how to provision a cluster, deploy Kyma, and <!--- update is not among the alpha commands ---> with alpha commands.

## Provision a cluster locally or using cloud providers
<!---provision gke is not among the alpha commands--->
To provision a cluster on a specific cloud provider (in this example GCP), run:

```bash
kyma alpha provision gke --credentials {SERVICE_ACCOUNT_KEY_FILE_PATH} --name {CLUSTER_NAME} --project {GCP_PROJECT} 
```

To provision a Minikube cluster, run:
<!---provision minikube is not among the alpha commands--->
```bash
kyma alpha provision minikube
```

## Deploy Kyma

To deploy Kyma using your own domain, run:

```bash
kyma alpha deploy --domain {DOMAIN} --tls-cert {TLS_CERT} --tls-key {TLS_KEY}
```

`{TLS_CERT}` and `{TLS_KEY}` are the TLS certificate and TLS key for the domain used. Note that both of them must be encoded in base64. You can use [OpenSSL](https://www.openssl.org/) (which is installed by default with macOS) to convert your TLS certificate and TLS key to base64-encoded values.

You can simplify the process to deploy Kyma with your own domainin  `{DOMAIN}` the following way:

1. Encode the TLS certificate and TLS key, and export them as environment variables:
  
    ```bash
    export TLS_CERT="$(openssl base64 -in {TLS_CERT_FILE_PATH})"
    export TLS_KEY="$(openssl base64 -in {TLS_KEY_FILE_PATH})"
    ```

    `{TLS_CERT_FILE_PATH}` is the path to the file containing the TLS certificate, and `{TLS_KEY_FILE_PATH}` is the path to the file containing the TLS key.

2. Use these environment variables in the command:

    ```bash
    kyma alpha deploy --domain {DOMAIN} --tls-cert $TLS_CERT --tls-key $TLS_KEY
    ```

To deploy Kyma from the `master` branch, run:

```bash
kyma alpha deploy --source master
```

To deploy Kyma using your own Kyma installer image, run:

```bash
kyma alpha deploy --source {IMAGE}
```

To build an image from your local sources and deploy Kyma on a remote cluster based on this image, run:

```bash
kyma alpha deploy --source local --custom-image {IMAGE}
```
>    **EXAMPLE:** If your `{IMAGE}` is `user/my-kyma-installer:v1.4.0`, this command builds an image from your local sources and push it your repository `user/my-kyma-installer` with the tag `v1.4.0`.<br>
Then, this image `user/my-kyma-installer:v1.4.0` is used in the `Deployment` for `kyma-installer`.

To deploy kyma with only specific components, run:

```bash
kyma alpha deploy --components {COMPONENTS_FILE_PATH}
```
`{COMPONENTS_FILE_PATH}` is the path to a YAML file containing the desired component list to be installed, for example:
```bash
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
In this example, only these 6 components are deployed on the cluster.

To override the standard Kyma installation, run:
<!---didn't find override among the deploy options--->
```bash
kyma alpha deploy --override {OVERRIDE_FILE_PATH}
```
`{OVERRIDE_FILE_PATH}` is the path to a YAML file containing the parameters to override, for example:
```bash
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
In this example, overrides are provided for 2 different components: `ory` and `monitoring`.
 
- For `ory`, the values of `hydra.deployment.resources.limits.cpu` and `hydra.deployment.resources.requests.cpu` will be overriden to `153m` and `53m` respectively.
- For `monitoring`, the values of `alertmanager.alertmanagerSpec.resources.limits.memory` and `alertmanager.alertmanagerSpec.resources.requests.memory` will be overriden to `304Mi` and `204Mi` respectively.

It is also possible to provide multiple override files at the same time

  ```bash
  kyma install --override {OVERRIDEL_FILE_1_PATH} --override {OVERRIDE_FILE_2_PATH}
  ```

## Upgrade Kyma
<!---upgrade is not among the alpha commands--->

To upgrade the Kyma version on the cluster, you can run the `upgrade` command which has the same structure and same flags as the `install` command.

## Test Kyma
<!---test is not among the alpha commands--->

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

## Use the *--values* flag

The *--values* flag can be helpful for the following cases: 

- To set the administrator passwort, use:

```bash
--value=global.xxx=myPWD
```

- To do something similarly exciting, run: 
<!---URGENTLY NEED INPUT!!!--->

```bash
--value=???
```
