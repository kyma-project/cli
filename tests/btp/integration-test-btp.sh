#!/bin/bash

echo "Running kyma integration tests uing connected managed kyma runtime"

# -------------------------------------------------------------------------------------
echo "Step1: Generating temporary access for new service account"

../../bin/kyma alpha kubeconfig generate --clusterrole cluster-admin --serviceaccount test-sa --output /tmp/kubeconfig.yaml --time 2h

export KUBECONFIG="/tmp/kubeconfig.yaml"
if [[ $(kubectl config view --minify --raw | yq '.users[0].name') != 'test-sa' ]]; then
    exit 1
fi
echo "Running test in user context of: $(kubectl config view --minify --raw | yq '.users[0].name')"
# -------------------------------------------------------------------------------------
echo "Step2: List modules"
../../bin/kyma alpha module list

# -------------------------------------------------------------------------------------
echo "Step3: Connecting to a service manager from remote BTP subaccount"

kubectl create secret generic remote-service-manager-credentials --namespace kyma-system --from-env-file sm.env

# -------------------------------------------------------------------------------------
echo "Step4: Create service instance reference to a shared object-store service instance"

echo "Waiting for CRD btp operator"
while ! kubectl get crd btpoperators.operator.kyma-project.io; do echo "Waiting for CRD btp operator..."; sleep 1; done
kubectl wait --for condition=established crd/btpoperators.operator.kyma-project.io
while ! kubectl get btpoperators.operator.kyma-project.io btpoperator --namespace kyma-system; do echo "Waiting for btpoperator..."; sleep 1; done
kubectl wait --for condition=Ready btpoperators.operator.kyma-project.io/btpoperator -n kyma-system --timeout=180s


# TODO - change after btp operator commands are extracted as btp module cli extension
../../bin/kyma@v3 alpha reference-instance \
    --btp-secret-name remote-service-manager-credentials \
    --namespace kyma-system \
    --offering-name objectstore \
    --plan-selector standard \
    --reference-name object-store-reference
kubectl apply -n kyma-system -f ./k8s-resources/object-store-binding.yaml

while ! kubectl get secret object-store-reference-binding --namespace kyma-system; do echo "Waiting for object-store-reference-binding secret..."; sleep 5; done

# -------------------------------------------------------------------------------------
# Enable Docker Registry
echo "Step5: Enable Docker Registry from experimental channel (with persistent BTP based storage)"
../../bin/kyma alpha module add docker-registry --channel experimental --cr-path k8s-resources/exposed-docker-registry.yaml

echo "..waiting for docker registry"
kubectl wait --for condition=Installed dockerregistries.operator.kyma-project.io/custom-dr -n kyma-system --timeout=360s

dr_external_url=$(../../bin/kyma@v3 alpha registry config --externalurl)

# TODO new cli command, for example
# dr_internal_pull_url=$(../../bin/kyma@v3 alpha registry config --internalurl)
dr_internal_pull_url=$(kubectl get dockerregistries.operator.kyma-project.io -n kyma-system custom-dr -ojsonpath={.status.internalAccess.pullAddress})

../../bin/kyma@v3 alpha registry config --output config.json

echo "Docker Registry enabled \n\t(URLs: $dr_external_url, $dr_internal_pull_url)"
echo "\t config.json for docker CLI access generated"
# -------------------------------------------------------------------------------------
echo "Step6: Map SAP Hana DB instance with Kyma runtime"

../../bin/kyma@v3 alpha hana map --credentials-path hana-admin-api-binding.json

# -------------------------------------------------------------------------------------
echo "Step7: Push bookstore application (w/o Dockerfile)"

# build hdi-deploy via pack and push it via docker CLI (external url)
pack build hdi-deploy:latest -p sample-http-db-nodejs/hdi-deploy -B paketobuildpacks/builder:base
docker tag hdi-deploy:latest $dr_external_url/hdi-deploy:latest
docker --config . push $dr_external_url/hdi-deploy:latest

# continue



# -------------------------------------------------------------------------------------
# echo "Cleanup"

# kubectl delete dockerregistries.operator.kyma-project.io -n kyma-system custom-dr
# ../../bin/kyma@v3 alpha module delete docker-registry

# kubectl delete servicebindings.services.cloud.sap.com -n kyma-system object-store-reference-binding
# kubectl delete serviceinstances.services.cloud.sap.com -n kyma-system object-store-reference
# kubectl delete secret -n kyma-system remote-service-manager-credentials 

# TODO new command ?
# ../../bin/kyma@v3 alpha hana unmap --credentials-path hana-admin-api-binding.json
# -------------------------------------------------------------------------------------


exit 0