# Fast Prototyping in SAP BTP, Kyma Runtime Using App Push

This tutorial shows how to deploy a single-container application to SAP BTP Kyma runtime in one CLI command using `kyma app push`, then evolve it into an automated GitHub Actions CD pipeline. You don't need to hand-craft YAML, set up a container registry, or configure a CI/CD pipeline upfront — just code, deploy, and iterate.

It is a good fit when you have an app in any language supported by [Cloud Native Buildpacks](https://buildpacks.io/) (Java, Node.js, Go, Python, .NET) and want a clear path from local development to a working prototype in the SAP BTP context — all without writing a Dockerfile.

For this example, use a Spring Boot application that exposes a REST API for managing movies, storing data in BTP Object Store.

## Prerequisites

- SAP BTP, Kyma runtime enabled
- [Kyma CLI](https://help.sap.com/docs/btp/sap-business-technology-platform/kyma-cli?locale=en-US#install-kyma-cli) installed
- [kubectl configured to kubeconfig downloaded from SAP BTP, Kyma runtime](https://developers.sap.com/tutorials/cp-kyma-download-cli.html)
- [Git](https://git-scm.com/downloads) installed
- Check if the Istio, API Gateway, and BTP Operator Kyma modules are added on your runtime. If not, [add them](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module?locale=en-US#adding-a-kyma-module).
  - The Docker Registry community module ([added](https://kyma-project.io/external-content/community-modules/docs/user/README.html#quick-install))

### Clone the Git Repository

1. Clone the `movies-rest` folder from the [kyma-runtime-samples](https://github.com/SAP-samples/kyma-runtime-samples) repository:

    ```Shell/Bash
    git clone --filter=blob:none --no-checkout https://github.com/SAP-samples/kyma-runtime-samples
    cd kyma-runtime-samples
    git sparse-checkout init --no-cone
    git sparse-checkout set 'movies-rest/'
    git checkout
    ```

### Create the Object Store ServiceInstance and ServiceBinding

1. Create the `dev` namespace:

    ```Shell/Bash
    kubectl create namespace dev
    ```

2. Create the Object Store ServiceInstance and ServiceBinding:

    ```Shell/Bash
    kubectl -n dev apply -f - <<EOF
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

3. Wait for the binding to become ready:

    ```Shell/Bash
    kubectl -n dev get servicebinding object-store-binding -w
    ```

    Once the `STATUS` column shows `Ready`, a Kubernetes Secret named `object-store-binding` is created in the namespace with the Object Store credentials.

### Deploy the Application

1. From the `movies-rest` directory, run the following command to push the application:

    ```Shell/Bash
    kyma app push \
      --name movies-rest \
      --namespace dev \
      --code-path . \
      --container-port 8080 \
      --expose \
      --istio-inject=true \
      --mount-service-binding-secret object-store-binding \
      --env-from-file .env
    ```

    What happens under the hood:
    - Source code is built into a container image using [Cloud Native Buildpacks](https://buildpacks.io/) (Paketo). No Dockerfile is required — Buildpacks detect `pom.xml` and automatically build a Java application with the correct JDK.
    - The image is pushed to the in-cluster Docker Registry.
    - A Deployment, Service, and APIRule are created.
    - The Object Store binding secret is mounted at `/bindings/secret-object-store-binding`, and `SERVICE_BINDING_ROOT=/bindings` is set automatically.

    > **Note:** The same approach works for any language supported by Cloud Native Buildpacks — Node.js, Go, Python, .NET, and more.

### Verify the Deployment

1. Once `kyma app push` completes, it prints the app URL:

    ```
    The movies-rest app is available under the
    movies-rest.<CLUSTER_DOMAIN>.kyma.ondemand.com
    ```

    > **Tip:** In quiet mode, the app URL is the only output — useful for capturing it in scripts:
    >
    > ```Shell/Bash
    > APP_URL=$(kyma app push ... --quiet)
    > echo $APP_URL
    > ```

2. Open the interactive Swagger UI in your browser at `https://<APP_URL>/swagger-ui.html` and test the CRUD operations on the movies endpoint.

    The OpenAPI specification is also available at `https://<APP_URL>/v3/api-docs`.

### (Optional) Automate Deployments with GitHub Actions

Once your prototype stabilizes, you can automate deployments on every push to your repository using GitHub Actions.

1. Push your application code to a GitHub repository, for example `https://github.com/<YOUR-ORG>/movies-rest`.

2. Authorize the repository's GitHub Actions workflows to deploy to your Kyma cluster:

    ```Shell/Bash
    kyma alpha authorize repository \
      --client-id my-client-id-for-gh-action \
      --cluster-wide \
      --clusterrole edit \
      --repository <YOUR-ORG>/movies-rest
    ```

    This command configures your Kyma cluster to trust GitHub OIDC tokens issued for the specified repository. The workflow obtains cluster access using a short-lived GitHub OIDC token — no long-lived credentials are stored. The only values you need to keep as secrets are the API server URL and CA certificate, which are connection details rather than credentials.

    > **Note:** The `--clusterrole edit` flag is used here for simplicity. In production, choose the most restrictive ClusterRole that satisfies your workflow's needs. You can also limit authorization to a specific workflow, branch, or environment using `--require-claim`. Run `kyma alpha authorize repository --help` for details.

3. Add the cluster connection details as GitHub Actions secrets. Run the following commands locally to get the values:

    ```Shell/Bash
    kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}'
    kubectl config view --minify --raw -o jsonpath='{.clusters[0].cluster.certificate-authority-data}'
    ```

4. In your GitHub repository, go to **Settings** > **Secrets and variables** > **Actions** > **New repository secret** and add:
    - `SERVER` — the API server URL returned by the first command
    - `CA_CRT` — the base64-encoded CA certificate returned by the second command

5. Create the following GitHub Actions workflow file in your repository at `.github/workflows/deploy.yaml`:

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

          - name: Get kubeconfig
            id: oidc
            uses: kyma-project/setup-kyma-cli/kubeconfig@v1.1.0
            with:
              audience: "my-client-id-for-gh-action"
              api-server-url: "${{ secrets.SERVER }}"
              ca-crt: "${{ secrets.CA_CRT }}"
              id-token-auto-refresh: "true"

          - name: Set short SHA
            run: echo "SHORT_SHA=${GITHUB_SHA::7}" >> $GITHUB_ENV

          - uses: kyma-project/setup-kyma-cli/app-push@v1.1.0
            with:
              name: movies-rest
              namespace: dev
              code-path: .
              build-tag: "${{ env.SHORT_SHA }}"
              container-port: "8080"
              expose: "true"
              istio-inject: "true"
              mount-service-binding-secret: object-store-binding
              kubeconfig: "${{ steps.oidc.outputs.kubeconfig }}"
              env-from-file: .env
              append-output-path: /swagger-ui.html
    ```

6. Every push to the `main` branch now triggers a fresh build and deploy. No local tooling is required after the initial setup.

![GitHub Actions deploy workflow summary showing a successful deploy job and the application URL](../../assets/app-push-gh-action-summary.png)

## Summary

With `kyma app push` you go from source code to a running, externally accessible application with BTP service bindings in a single command. The same deployment can then be moved into a GitHub Actions workflow with zero code changes — just copy the flags into the action inputs.

This is not a throwaway prototype tool. The Deployment, Service, and APIRule it creates are standard Kubernetes resources you can later manage with Helm, Kustomize, or any GitOps tool.
