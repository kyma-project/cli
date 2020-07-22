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