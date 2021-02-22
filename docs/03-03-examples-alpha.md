---
title: Kyma CLI alpha command usage examples
type: Details
---

The following examples show how to provision a cluster, and to deploy and update Kyma with alpha commands.

## Provision a cluster
To provision a Kubernetes cluster, run:

```bash
kyma alpha provision k3s 
```


## Install Kyma

There are several ways to install Kyma:

- To install Kyma on a local cluster, you can simply use the `deploy` command.
Kyma provides a default domain under the URL `https://console.local.kyma.dev`.

    ```
    kyma alpha deploy 
    ```
   
- You can also install Kyma using your own domain name, or on a cluster that does not run locally.<br>
    If you use a custom domain, you must provide the certificate and key as files. If you don't have a certificate yet, you can create a self-signed certificate and key:

    ```bash
    openssl req -x509 -newkey rsa:4096 -keyout key.pem -out crt.pem -days 365
    ```
    When prompted, provide your credentials, such as your name and your domain.

    Then, pass the certificate files to the deploy command:

        ```bash
        kyma alpha deploy --domain {DOMAIN} --tls-cert crt.pem --tls-key key.pem
        ```

- Optionally, you can specify from which source you want to deploy Kyma, such as the `master` branch, a specific PR, or a release version. For more details, see the documentation for the `alpha deploy` command.<br>
For example, to install Kyma from a specific version, such as `1.19.1`, run:

    ```bash
    kyma alpha deploy --source=1.19.1
    ```

- Alternatively, to build Kyma from your local sources and deploy it on a remote cluster, run:

    ```bash
    kyma alpha deploy --source=local
    ```
    > **NOTE:** By default, Kyma expects to find the sources in the `./workspace folder`. To adjust the path, set the `-w ${PATH_TO_KYMA_SOURCES}` parameter.

- To deploy Kyma with only specific components, run:

    ```bash
    kyma alpha deploy --components {COMPONENTS_FILE_PATH}
    ```
    `{COMPONENTS_FILE_PATH}` is the path to a YAML file containing the desired component list to be installed. In the following example, only these 6 components are deployed on the cluster:

    ```bash
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

- To override the standard Kyma installation, run:

    ```bash
    kyma alpha deploy --values-file {OVERRIDE_FILE_PATH}
    ```

    `{COMPONENTS_FILE_PATH}` is the path to a YAML file containing the desired component list to be installed. In the following example, only these 6 components are deployed on the cluster:
 
    - For `ory`, the values of `hydra.deployment.resources.limits.cpu` and `hydra.deployment.resources.requests.cpu` will be overriden to `153m` and `53m` respectively.
    
    - For `monitoring`, the values of `alertmanager.alertmanagerSpec.resources.limits.memory` and `alertmanager.alertmanagerSpec.resources.requests.memory` will be overriden to `304Mi` and `204Mi` respectively.
    
    ```bash
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

- You can also provide multiple override files at the same time:

    ```bash
    kyma deploy --values-file {OVERRIDEL_FILE_1_PATH} --values-file {OVERRIDE_FILE_2_PATH}
    ```

- Alternatively, you can specify single values instead of a file:

    ```bash
    kyma deploy --value ory.hydra.deployment.resources.limits.cpu=153m \
    --value ory.hydra.deployment.resources.requests.cpu=53m \
    --value monitoring.alertmanager.alertmanagerSpec.resources.limits.memory=304Mi \
    --value monitoring.alertmanager.alertmanagerSpec.resources.requests.memory=204Mi
    ```

## Upgrade Kyma

To upgrade the Kyma version on the cluster, you also use the `alpha deploy` command as described in section [Install Kyma](#install-kyma).

If you upgrade from one Kyma release to a newer one, an automatic compatibility check compares your current version and the new release.<br>
Note that the compatibility check doesn't work if you installed Kyma from a PR, branch, revision, or local version.


## Set a global password

To set the administrator password, use:

```bash
--value=global.xxx=myPWD
```
