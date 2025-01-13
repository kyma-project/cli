#!/bin/bash

echo "Running basic test scenario for kyma@v3 CLI on k3d runtime"

# -------------------------------------------------------------------------------------
# Generate kubeconfig for service account

echo "Step1: Generating temporary access for new service account"
../../bin/kyma@v3 alpha access --clusterrole cluster-admin --name test-sa --output /tmp/kubeconfig.yaml --time 2h
export KUBECONFIG="/tmp/kubeconfig.yaml"
if [[ $(kubectl config view --minify --raw | yq '.users[0].name') != 'test-sa' ]]; then
    exit 1
fi
echo "Running test in user context of: $(kubectl config view --minify --raw | yq '.users[0].name')"

# -------------------------------------------------------------------------------------
# Enable Docker Registry

echo "Step2: Enable latest Docker Registry release"
kubectl create namespace kyma-system || true
kubectl apply -f https://github.com/kyma-project/docker-registry/releases/latest/download/dockerregistry-operator.yaml
kubectl apply -f https://github.com/kyma-project/docker-registry/releases/latest/download/default-dockerregistry-cr.yaml -n kyma-system
echo "..waiting for docker registry"
kubectl wait --for condition=Installed dockerregistries.operator.kyma-project.io/default -n kyma-system --timeout=360s

# -------------------------------------------------------------------------------------
# Push sample go app

echo "Step3: Push sample Go application (tests/k3d/sample-go)"
../../bin/kyma@v3 alpha app push --name test-app --code-path sample-go
kubectl wait --for condition=Available deployment test-app --timeout=60s
kubectl port-forward deployments/test-app 8080:8080 &
sleep 3 # wait for ports to get forwarded
response=$(curl localhost:8080)
echo "HTTP response from sample app: $response"

if [[ $response != 'okey dokey' ]]; then
    exit 1
fi

# -------------------------------------------------------------------------------------
# Cleanup

echo "Step4: Cleanup"
kubectl delete deployment test-app

# -------------------------------------------------------------------------------------
exit 0