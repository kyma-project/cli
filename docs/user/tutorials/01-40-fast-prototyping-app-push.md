# Fast Prototyping on SAP BTP Kyma: App Push

This tutorial shows how to deploy a single-container application to SAP BTP Kyma runtime in one CLI command using `kyma app push`, then evolve it into an automated GitHub Actions CD pipeline. No YAML hand-crafting, no container registry setup, no CI/CD pipeline to configure upfront — just code → deploy → iterate.

It is a good fit when you have an app in any language supported by [Cloud Native Buildpacks](https://buildpacks.io/) (Java, Node.js, Go, Python, .NET) and want a clear path from local development to a working prototype in the SAP BTP context — all without writing a Dockerfile.

> **Note:** For lightweight event-driven workloads where you want zero container knowledge, see [Fast Prototyping With Serverless Functions](01-50-fast-prototyping-serverless-functions.md).

## Prerequisites

- [Kyma CLI](https://help.sap.com/docs/btp/sap-business-technology-platform/kyma-cli?locale=en-US#install-kyma-cli) installed.
- An [SAP BTP Kyma runtime](https://help.sap.com/docs/btp/sap-business-technology-platform/kyma-environment) provisioned in your subaccount.
- The following [Kyma modules enabled](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module#adding-a-kyma-module) on your runtime:
  - **Istio** (enabled by default) — service mesh and networking
  - **API Gateway** (enabled by default) — external exposure via APIRule
  - **BTP Operator** (enabled by default) — manages BTP service instances and bindings
  - **Docker Registry** ([community module](https://kyma-project.io/external-content/community-modules/docs/user/README.html#quick-install)) — in-cluster container registry for building and storing images (no external registry needed)
- Your SAP BTP subaccount must be entitled to use the SAP Object Store service (service plan `standard`). Add the entitlement in the BTP cockpit under **Entitlements** if not already assigned.

## Step 1: Create Your App

For this example, use a Spring Boot application that exposes a REST API for managing movies, storing data in BTP Object Store. A ready-to-use version is available in the [`examples/movies-api`](../../../examples/movies-api) folder.

Create the project directory and mirror the same file structure:

```bash
mkdir "movies-rest" && cd "movies-rest"
```

Create the following files, using [`examples/movies-api`](../../../examples/movies-api) as reference:

```
movies-rest/
├── .env
├── pom.xml
└── src/
    └── main/
        ├── java/
        │   └── com/
        │       └── example/
        │           └── movies/
        │               ├── Application.java
        │               ├── Movie.java
        │               ├── MovieController.java
        │               └── ObjectStoreConfig.java
        └── resources/
            └── application.properties
```

## Step 2: Create the BTP Object Store Service Instance and Binding

The application needs an Object Store service instance secret mounted into the workload.
Create the service instance and binding using the ServiceInstance and ServiceBinding custom resources:

```bash
kubectl apply -f - <<EOF
apiVersion: services.cloud.sap.com/v1
kind: ServiceInstance
metadata:
  name: object-store-instance
spec:
  serviceOfferingName: objectstore
  servicePlanName: standard
---
apiVersion: services.cloud.sap.com/v1
kind: ServiceBinding
metadata:
  name: object-store-binding
spec:
  serviceInstanceName: object-store-instance
EOF
```

Wait for the binding to become ready:

```bash
kubectl get servicebindings object-store-binding -w
```

Once the status shows `Ready: True`, a Kubernetes Secret named `object-store-binding` is created in the namespace with the Object Store credentials. This secret is mounted to the application workload as part of `kyma app push` command execution.

## Step 3: Deploy

One command builds, pushes, and deploys your app — and exposes it externally:

```bash
kyma app push \
  --name movies-rest \
  --code-path . \
  --container-port 8080 \
  --expose \
  --istio-inject=true \
  --mount-service-binding-secret object-store-binding \
  --env-from-file .env
```

The `.env` file contains JVM memory tuning required to fit within the default 512Mi container limit:

```properties
BPL_JVM_THREAD_COUNT=20
JAVA_TOOL_OPTIONS=-XX:ReservedCodeCacheSize=40M -XX:MaxMetaspaceSize=80M -Xss512k
```

What happens under the hood:

1. Source code is built into a container image using [Cloud Native Buildpacks](https://buildpacks.io/) (Paketo).
2. Image is pushed to the in-cluster docker-registry.
3. A Deployment, Service, and APIRule are created.
4. The BTP Object Store binding secret is mounted at `/bindings/secret-object-store-binding`.
5. `SERVICE_BINDING_ROOT=/bindings` environment variable is set automatically.

> **Note:** No Dockerfile required. Buildpacks detect `pom.xml` and automatically build a Java application with the correct JDK. The same approach works for Node.js, Go, Python, .NET, and more.
>
> If you need more control over the image build, `kyma app push` also supports:
> - `--dockerfile <path>` — build the image from your own Dockerfile instead of using Buildpacks.
> - `--image <image>` — skip the build entirely and deploy a pre-built image already pushed to a registry.

## Step 4: Verify

The OpenAPI specification is available at `https://<YOUR-APP-URL>/v3/api-docs` and the interactive Swagger UI at `https://<YOUR-APP-URL>/swagger-ui.html`.
Use the Swagger UI to test the CRUD operations.


## Evolution: Move to GitHub Actions CD

Once your prototype stabilizes, automate deployments.
Push the code to a GitHub repository, for example `https://github.com/acme/movies-rest`, and authorize the repository's GitHub Actions workflows to apply changes in the target cluster.

To authorize the repository, run:

```bash
kyma alpha authorize repository \
  --client-id my-client-id-for-gh-action \
  --cluster-wide \
  --clusterrole edit \
  --repository acme/movies-rest
```

> **Note:** Use the most restrictive ClusterRole that satisfies your workflow's needs — `edit` is used here for simplicity, but consider a narrower role. You can also limit authorization to a specific workflow, branch, or environment by passing additional required OIDC claims with `--require-claim`. Run `kyma alpha authorize repository --help` for details.

Now you can automate deployments on every push to the main branch using the [`kyma-project/setup-kyma-cli/app-push`](https://github.com/kyma-project/setup-kyma-cli/tree/main/app-push) GitHub Action.

The action wraps the same `kyma app push` command you ran locally — same flags, same behavior. The workflow obtains cluster access using a GitHub OIDC token — no long-lived credentials are involved. The only values stored as secrets are the API server URL and CA certificate, which are not sensitive credentials but rather connection details needed to reach the cluster.

```yaml
name: Deploy

permissions:
  id-token: write
  contents: read

on:
  push:
    branches:
      - main

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Kyma CLI
        uses: kyma-project/setup-kyma-cli@v1.1.0

      - name: "get kubeconfig"
        id: oidc
        uses: kyma-project/setup-kyma-cli/kubeconfig@v1.1.0
        with:
          audience: "my-client-id-for-gh-action"
          api-server-url: "${{ secrets.SERVER }}" # fetch from secure secret store, i.e., Vault or GitHub secrets
          ca-crt: "${{ secrets.CA_CRT }}" # fetch from secure secret store, i.e., Vault or GitHub secrets
          id-token-auto-refresh: "true"

      - name: Set short SHA
        run: echo "SHORT_SHA=${GITHUB_SHA::7}" >> $GITHUB_ENV

      - uses: kyma-project/setup-kyma-cli/app-push@v1.1.0
        with:
          name: movies-rest
          code-path: . # relative path of the source code
          build-tag: "${{ env.SHORT_SHA }}" # override with custom image build tag
          container-port: "8080"
          expose: "true"
          istio-inject: "true"
          mount-service-binding-secret: object-store-binding
          kubeconfig: "${{ steps.oidc.outputs.kubeconfig }}"
          env-from-file: .env # relative path of the .env file
          append-output-path: /swagger-ui.html

```

Every push to `main` triggers a fresh build and deploy — no local tooling required.

![GitHub Actions deploy workflow summary showing a successful deploy job and the application URL](../../assets/app-push-gh-action-summary.png)

## Summary

With `kyma app push` you go from source code to a running, externally accessible application with BTP service bindings in a single command. The same deployment can then be moved into a GitHub Actions workflow with zero code changes — just copy the flags into the action inputs.

This is not a throwaway prototype tool. The Deployment, Service, and APIRule it creates are standard Kubernetes resources you can later manage with Helm, Kustomize, or any GitOps tool.
