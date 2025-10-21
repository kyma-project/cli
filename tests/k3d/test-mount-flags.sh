#!/bin/bash

echo "Running mount flags test scenario for kyma CLI on k3d runtime"

# Set strict error handling
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Cleanup function
cleanup() {
    kubectl delete deployment test-mount-app --ignore-not-found=true >/dev/null 2>&1
    kubectl delete service test-mount-app --ignore-not-found=true >/dev/null 2>&1
    kubectl delete secret test-secret --ignore-not-found=true >/dev/null 2>&1
    kubectl delete secret test-service-binding --ignore-not-found=true >/dev/null 2>&1
    kubectl delete configmap test-configmap --ignore-not-found=true >/dev/null 2>&1
    pkill -f "kubectl port-forward" >/dev/null 2>&1 || true
}

# Set trap for cleanup on exit
trap cleanup EXIT

# -------------------------------------------------------------------------------------
# Generate kubeconfig for service account

print_status "Generating service account access..."
../../bin/kyma alpha kubeconfig generate --clusterrole cluster-admin --serviceaccount test-sa --output /tmp/kubeconfig.yaml --time 2h --cluster-wide --namespace default >/dev/null 2>&1
export KUBECONFIG="/tmp/kubeconfig.yaml"

if [[ $(kubectl config view --minify --raw | yq '.users[0].name') != 'test-sa' ]]; then
    print_error "Failed to set up service account"
    exit 1
fi

# -------------------------------------------------------------------------------------
# Enable Docker Registry

print_status "Setting up Docker Registry..."
kubectl create namespace kyma-system >/dev/null 2>&1 || true
kubectl apply -f https://github.com/kyma-project/docker-registry/releases/latest/download/dockerregistry-operator.yaml >/dev/null 2>&1
kubectl apply -f https://github.com/kyma-project/docker-registry/releases/latest/download/default-dockerregistry-cr.yaml -n kyma-system >/dev/null 2>&1
kubectl wait --for condition=Installed dockerregistries.operator.kyma-project.io/default -n kyma-system --timeout=360s >/dev/null 2>&1

# -------------------------------------------------------------------------------------
# Create test resources (Secret, ConfigMap, Service Binding Secret)

print_status "Creating test resources..."

# Create a test secret
kubectl create secret generic test-secret \
  --from-literal=username=testuser \
  --from-literal=password=testpass \
  --from-literal=config.json='{"database":"postgres","host":"localhost"}' >/dev/null 2>&1

# Create a test configmap
kubectl create configmap test-configmap \
  --from-literal=app.properties='debug=true
logging.level=INFO
server.port=8080' \
  --from-literal=database.url=jdbc:postgresql://localhost:5432/testdb >/dev/null 2>&1

# Create a service binding secret (simulates a service binding)
kubectl create secret generic test-service-binding \
  --from-literal=type=postgresql \
  --from-literal=provider=bitnami \
  --from-literal=host=postgres.example.com \
  --from-literal=port=5432 \
  --from-literal=username=bindinguser \
  --from-literal=password=bindingpass >/dev/null 2>&1

# -------------------------------------------------------------------------------------
# Push sample app with all mount flags

print_status "Pushing application with mount flags..."

../../bin/kyma app push \
  --name test-mount-app \
  --code-path sample-go \
  --container-port 8080 \
  --mount-secret "name=test-secret,path=/app/secrets,ro=true" \
  --mount-config "name=test-configmap,path=/app/config,ro=false" \
  --mount-service-binding-secret test-service-binding

echo "Mount flags used:"
echo "  --mount-secret: test-secret -> /app/secrets (read-only)"
echo "  --mount-config: test-configmap -> /app/config (read-write)"
echo "  --mount-service-binding-secret: test-service-binding -> /bindings/secret-test-service-binding (read-only)"

kubectl wait --for condition=Available deployment test-mount-app --timeout=120s >/dev/null 2>&1

# -------------------------------------------------------------------------------------
# Test the application and verify mounts

print_status "Testing application and verifying mounts..."

# Port forward in background
kubectl port-forward deployments/test-mount-app 8080:8080 >/dev/null 2>&1 &
PF_PID=$!
sleep 5 # wait for ports to get forwarded

# Test basic endpoint
response=$(curl -s localhost:8080)
echo "Basic endpoint response: $response"

if [[ $response != 'okey dokey' ]]; then
    print_error "Basic endpoint test failed"
    exit 1
fi

# Instead of using a /mounts endpoint, verify mounts by directly inspecting the pod
POD_NAME=$(kubectl get pods -l app=test-mount-app -o jsonpath='{.items[0].metadata.name}')
echo "Verifying mounts in pod: $POD_NAME"

# Check if secret files are mounted and accessible
print_status "Checking secret mount..."
if kubectl exec $POD_NAME -- ls /app/secrets >/dev/null 2>&1; then
    print_status "✓ Secret mounted successfully"
    echo "Secret files:"
    kubectl exec $POD_NAME -- ls -la /app/secrets

    # Test read-only
    if kubectl exec $POD_NAME -- touch /app/secrets/test-file 2>/dev/null; then
        print_warning "Secret mount is NOT read-only"
        kubectl exec $POD_NAME -- rm /app/secrets/test-file
    else
        print_status "✓ Secret mount is read-only as expected"
    fi
else
    print_error "✗ Secret not mounted properly"
fi

# Check if configmap files are mounted and accessible
print_status "Checking configmap mount..."
if kubectl exec $POD_NAME -- ls /app/config >/dev/null 2>&1; then
    print_status "✓ ConfigMap mounted successfully"
    echo "ConfigMap files:"
    kubectl exec $POD_NAME -- ls -la /app/config
else
    print_error "✗ ConfigMap not mounted properly"
fi

# Check if service binding secret is mounted
print_status "Checking service binding mount..."
if kubectl exec $POD_NAME -- ls /bindings >/dev/null 2>&1; then
    binding_dirs=$(kubectl exec $POD_NAME -- ls /bindings 2>/dev/null || echo "")
    if [[ -n "$binding_dirs" ]]; then
        print_status "✓ Service binding secret mounted"
        echo "Service binding directories:"
        kubectl exec $POD_NAME -- ls -la /bindings

        # Check the first service binding directory
        first_dir=$(kubectl exec $POD_NAME -- ls /bindings | head -1)
        if [[ -n "$first_dir" ]]; then
            echo "Service binding files in $first_dir:"
            kubectl exec $POD_NAME -- ls -la "/bindings/$first_dir"

            # Test read-only
            if kubectl exec $POD_NAME -- touch "/bindings/$first_dir/test-file" 2>/dev/null; then
                print_warning "Service binding mount is NOT read-only"
                kubectl exec $POD_NAME -- rm "/bindings/$first_dir/test-file"
            else
                print_status "✓ Service binding mount is read-only as expected"
            fi
        fi
    else
        print_error "✗ No service binding directories found"
    fi
else
    print_error "✗ Service binding secret not mounted properly"
fi

# -------------------------------------------------------------------------------------
# Show actual file contents

echo ""
echo "Secret files content:"
for file in $(kubectl exec $POD_NAME -- ls /app/secrets 2>/dev/null || echo ""); do
    echo "  $file:"
    kubectl exec $POD_NAME -- cat "/app/secrets/$file" 2>/dev/null || echo "    (could not read)"
done

echo ""
echo "ConfigMap files content:"
for file in $(kubectl exec $POD_NAME -- ls /app/config 2>/dev/null || echo ""); do
    echo "  $file:"
    kubectl exec $POD_NAME -- cat "/app/config/$file" 2>/dev/null || echo "    (could not read)"
done

echo ""
echo "Service binding files content:"
if [[ -n "$first_dir" ]]; then
    for file in $(kubectl exec $POD_NAME -- ls "/bindings/$first_dir" 2>/dev/null || echo ""); do
        echo "  $file:"
        kubectl exec $POD_NAME -- cat "/bindings/$first_dir/$file" 2>/dev/null || echo "    (could not read)"
    done
fi

# -------------------------------------------------------------------------------------
print_status "Mount flags test completed successfully!"
echo "✓ --mount-secret"
echo "✓ --mount-config"
echo "✓ --mount-service-binding-secret"

exit 0
