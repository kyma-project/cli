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
echo "Step3: Connecting to remote BTP subaccount"
# -------------------------------------------------------------------------------------
echo "Step4: Create Shared Service Instance Reference"
# -------------------------------------------------------------------------------------
# Enable Docker Registry
echo "Step5: Enable Docker Registry from experimental channel (with persistent BTP based storage)"
../../bin/kyma alpha module add docker-registry --channel experimental --cr-path k8s-resources/exposed-docker-registry.yaml

echo "..waiting for docker registry"
kubectl wait --for condition=Installed dockerregistries.operator.kyma-project.io/default -n kyma-system --timeout=360s
# -------------------------------------------------------------------------------------
echo "Step6: Map bookstore DB"
# -------------------------------------------------------------------------------------
echo "Step7: Push bookstore application (w/o Dockerfile)"
# -------------------------------------------------------------------------------------


exit 0