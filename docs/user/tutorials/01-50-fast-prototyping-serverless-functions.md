# Fast Prototyping on SAP BTP Kyma: Serverless Functions

This tutorial shows how to deploy a serverless function to SAP BTP Kyma runtime with BTP service bindings using the CLI, then evolve it into a fully self-managed application by ejecting to plain Kubernetes manifests. No Dockerfile, no image registry, no Deployment manifests — just write a function and deploy.

It is a good fit when you're building a lightweight API, webhook handler, or event processor and want the fastest path from code to a running workload with BTP service bindings (e.g., Object Store, XSUAA, HANA Cloud) — while keeping an escape hatch to "eject" into standard Kubernetes resources later.

> For multi-language apps or when you want buildpack-based deployments with a GitHub Actions CD path, see [Fast Prototyping With App Push](01-40-fast-prototyping-app-push.md).

## Prerequisites

Install the Kyma CLI and enable the serverless module:

```bash
# Install Kyma CLI (macOS/Linux)
curl -fsSL https://get.kyma.io | bash

# Enable the serverless module
kyma module add serverless --default-config-cr
```

## Step 1: Scaffold a Function

```bash
mkdir my-function && cd my-function

kyma function init --runtime nodejs22
```

This creates a local `handler.js` and `package.json`. Edit the handler to use the BTP Object Store binding:

```javascript
// handler.js
const fs = require('fs');

module.exports = {
  main: async function (event, context) {
    // Access BTP Object Store via mounted service binding
    const bindingPath = '/bindings/object-store';
    try {
      const creds = JSON.parse(fs.readFileSync(`${bindingPath}/credentials`, 'utf8'));
      return { statusCode: 200, body: JSON.stringify({ endpoint: creds.endpoint }) };
    } catch (err) {
      return { statusCode: 500, body: JSON.stringify({ error: err.message }) };
    }
  }
};
```

## Step 2: Deploy

Deploy the function with the BTP Object Store binding mounted:

```bash
kyma function create my-function \
  --secret-mount object-store-binding=/bindings/object-store
```

The serverless module builds and runs the function. No container image, no registry, no Deployment manifest.

## Step 3: Verify

Check the function status and invoke it:

```bash
kyma function get my-function

# Port-forward to test locally
kubectl port-forward svc/my-function 8080:80
curl http://localhost:8080
# {"endpoint":"https://..."}
```

## BTP Service Bindings

The `--secret-mount` flag mounts a BTP service instance secret into your function following the [Service Binding Specification](https://servicebinding.io/):

1. Create a BTP service instance (e.g., Object Store) via the BTP Operator module or SAP BTP cockpit
2. Create a service binding — this produces a Kubernetes Secret with credentials
3. Pass the secret name and mount path to `--secret-mount` in the format `SECRET_NAME=MOUNT_PATH`

Your function reads credentials from the mounted path at runtime.

## Evolution: Eject to Kubernetes Manifests

When your function outgrows the serverless model — you need custom resource limits, sidecar containers, or want to manage deployment with Helm/Kustomize — eject it:

```bash
kyma function eject my-function --output-dir ./k8s-manifests
```

This generates plain Kubernetes YAML (Deployment, Service, ConfigMap with your source) in `./k8s-manifests/`. From here you fully own the deployment lifecycle:

- Add it to a Helm chart
- Deploy with `kubectl apply`
- Wire it into any CI/CD pipeline (GitHub Actions, Tekton, ArgoCD)

The function runtime is preserved — your code stays exactly the same. You just moved from "managed by Serverless module" to "managed by you."

## Summary

With `kyma function create` you go from a handler function to a running workload with BTP service bindings — zero container knowledge required. When your needs evolve, `kyma function eject` gives you full ownership of standard Kubernetes manifests without rewriting your application code.

This is a stepping stone, not a dead end. Start fast, evolve at your own pace.
