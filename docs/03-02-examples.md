---
title: Kyma CLI command usage examples
type: Details
---

The following examples show how to provision a cluster, install Kyma, and run the tests.

## Provision a cluster locally or using cloud providers

To provision a cluster on a specific cloud provider (in this example GCP), run:

```bash
kyma provision gke -c {SERVICE_ACCOUNT_KEY_FILE_PATH} -n {CLUSTER_NAME} -p {GCP_PROJECT} 
```

To provision a Minikube cluster, run:

```bash
kyma provision minikube
```

## Install Kyma

To install Kyma using your own domain, run:

```bash
kyma install --domain {DOMAIN} --tlsCert {TLS_CERT} --tlsKey {TLS_KEY}
```

To install Kyma from the latest `master` branch, run:

```bash
kyma install -s latest
```

To install Kyma using your own Kyma installer image, run:

```bash
kyma install -s {IMAGE}
```

To build an image from your local sources and install Kyma based on this image, run:

```bash
kyma install -s local --custom-image {IMAGE}
```

To install kyma with only specific components, run:

```bash
kyma install -c {YAML_FILE_PATH}
```
- {YAML_FILE_PATH} points to a YAML file with the desired component list to be installed such as the one below:
- ```bash
  components:
  - name: "cluster-essentials"
    namespace: "kyma-system"
  - name: "testing"
    namespace: "kyma-system"
  ```
- In this example, only `cluster-essentials` and `testing` components will be installed on the cluster.

To override the standard Kyma installation, run:
```bash
kyma install -o {YAML_FILE_PATH}
```
- {YAML_FILE_PATH} points to a YAML file with parameters to override such as the one below:
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
  kyma install --override {YAML_FILE_1_PATH} --override {YAML_FILE_2_PATH}
  ```
      

### Test Kyma

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
