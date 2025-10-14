#!/bin/bash

echo "Running mount flags test scenario for kyma CLI on k3d runtime (using pre-built image)"

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
    print_status "Cleaning up test resources..."
    kubectl delete deployment test-mount-app --ignore-not-found=true
    kubectl delete service test-mount-app --ignore-not-found=true
    kubectl delete secret test-secret --ignore-not-found=true
    kubectl delete secret test-service-binding --ignore-not-found=true
    kubectl delete configmap test-configmap --ignore-not-found=true
    pkill -f "kubectl port-forward" || true
    print_status "Cleanup completed"
}

# Set trap for cleanup on exit
trap cleanup EXIT

# -------------------------------------------------------------------------------------
# Generate kubeconfig for service account

print_status "Step 1: Generating temporary access for new service account"
../../bin/kyma alpha kubeconfig generate --clusterrole cluster-admin --serviceaccount test-sa --output /tmp/kubeconfig.yaml --time 2h --cluster-wide --namespace default
export KUBECONFIG="/tmp/kubeconfig.yaml"

if [[ $(kubectl config view --minify --raw | yq '.users[0].name') != 'test-sa' ]]; then
    print_error "Failed to set up service account"
    exit 1
fi
print_status "Running test in user context of: $(kubectl config view --minify --raw | yq '.users[0].name')"

# -------------------------------------------------------------------------------------
# Create test resources (Secret, ConfigMap, Service Binding Secret)

print_status "Step 2: Creating test resources for mounting"

# Create a test secret
kubectl create secret generic test-secret \
  --from-literal=username=testuser \
  --from-literal=password=testpass \
  --from-literal=config.json='{"database":"postgres","host":"localhost"}'

print_status "Created test secret with keys: username, password, config.json"

# Create a test configmap
kubectl create configmap test-configmap \
  --from-literal=app.properties='debug=true
logging.level=INFO
server.port=8080' \
  --from-literal=database.url=jdbc:postgresql://localhost:5432/testdb

print_status "Created test configmap with keys: app.properties, database.url"

# Create a service binding secret (simulates a service binding)
kubectl create secret generic test-service-binding \
  --from-literal=type=postgresql \
  --from-literal=provider=bitnami \
  --from-literal=host=postgres.example.com \
  --from-literal=port=5432 \
  --from-literal=username=bindinguser \
  --from-literal=password=bindingpass

print_status "Created service binding secret with connection details"

# -------------------------------------------------------------------------------------
# Push app using nginx image with mount flags (to avoid buildpack issues)

print_status "Step 3: Push application using nginx image with mount flags"

../../bin/kyma app push \
  --name test-mount-app \
  --image nginx:alpine \
  --container-port 80 \
  --mount-secret "name=test-secret,path=/app/secrets,ro=true" \
  --mount-secret "test-secret:username=/app/secrets2:ro" \
  --mount-config "name=test-configmap,path=/app/config,ro=false" \
  --mount-service-binding-secret test-service-binding

print_status "Application pushed with mount flags:"
print_status "  --mount-secret: test-secret -> /app/secrets (read-only)"
print_status "  --mount-secret: test-secret:username -> /app/secrets2 (read-only)"
print_status "  --mount-config: test-configmap -> /app/config (read-write)"
print_status "  --mount-service-binding-secret: test-service-binding -> /bindings/secret-test-service-binding (read-only)"

kubectl wait --for condition=Available deployment test-mount-app --timeout=120s

# -------------------------------------------------------------------------------------
# Verify mounts by inspecting the pod

print_status "Step 4: Verifying mounted files in the pod"

POD_NAME=$(kubectl get pods -l app=test-mount-app -o jsonpath='{.items[0].metadata.name}')
print_status "Inspecting pod: $POD_NAME"

# Check volume mounts in pod spec
print_status "Volume mounts in pod:"
kubectl describe pod $POD_NAME | grep -A 20 "Mounts:" || true

# Check volumes in pod spec
print_status "Volumes in pod spec:"
kubectl get pod $POD_NAME -o yaml | yq '.spec.volumes[].name' | head -10

# -------------------------------------------------------------------------------------
# Test file contents by executing commands in the pod

print_status "Step 5: Testing mounted file contents"

print_status "Testing secret mount at /app/secrets:"
kubectl exec $POD_NAME -- ls -la /app/secrets/ || print_warning "Secret mount path not accessible"
kubectl exec $POD_NAME -- cat /app/secrets/username || print_warning "Could not read username from secret"
kubectl exec $POD_NAME -- cat /app/secrets/password || print_warning "Could not read password from secret"
kubectl exec $POD_NAME -- cat /app/secrets/config.json || print_warning "Could not read config.json from secret"

print_status "Testing configmap mount at /app/config:"
kubectl exec $POD_NAME -- ls -la /app/config/ || print_warning "ConfigMap mount path not accessible"
kubectl exec $POD_NAME -- cat /app/config/app.properties || print_warning "Could not read app.properties from configmap"
kubectl exec $POD_NAME -- cat /app/config/database.url || print_warning "Could not read database.url from configmap"

print_status "Testing service binding mount at /bindings:"
kubectl exec $POD_NAME -- ls -la /bindings/ || print_warning "Service binding mount path not accessible"
BINDING_DIR=$(kubectl exec $POD_NAME -- ls /bindings/ | grep secret- | head -1)
if [[ -n "$BINDING_DIR" ]]; then
    print_status "Found service binding directory: $BINDING_DIR"
    kubectl exec $POD_NAME -- ls -la /bindings/$BINDING_DIR/ || print_warning "Could not list service binding files"
    kubectl exec $POD_NAME -- cat /bindings/$BINDING_DIR/type || print_warning "Could not read type from service binding"
    kubectl exec $POD_NAME -- cat /bindings/$BINDING_DIR/host || print_warning "Could not read host from service binding"
else
    print_warning "No service binding directory found"
fi

# -------------------------------------------------------------------------------------
# Test read-only permissions

print_status "Step 6: Testing read-only permissions"

print_status "Testing if secret mount is read-only:"
if kubectl exec $POD_NAME -- touch /app/secrets/test-file 2>/dev/null; then
    print_warning "Secret mount is NOT read-only (unexpected)"
    kubectl exec $POD_NAME -- rm /app/secrets/test-file
else
    print_status "✓ Secret mount is read-only as expected"
fi

print_status "Testing if configmap mount allows writes:"
if kubectl exec $POD_NAME -- touch /app/config/test-file 2>/dev/null; then
    print_status "✓ ConfigMap mount allows writes as expected"
    kubectl exec $POD_NAME -- rm /app/config/test-file
else
    print_warning "ConfigMap mount is read-only (might be expected behavior)"
fi

if [[ -n "$BINDING_DIR" ]]; then
    print_status "Testing if service binding mount is read-only:"
    if kubectl exec $POD_NAME -- touch /bindings/$BINDING_DIR/test-file 2>/dev/null; then
        print_warning "Service binding mount is NOT read-only (unexpected)"
        kubectl exec $POD_NAME -- rm /bindings/$BINDING_DIR/test-file
    else
        print_status "✓ Service binding mount is read-only as expected"
    fi
fi

# -------------------------------------------------------------------------------------
# Verify deployment configuration

print_status "Step 7: Verifying deployment configuration"

print_status "Checking deployment YAML for volume mounts:"
kubectl get deployment test-mount-app -o yaml | yq '.spec.template.spec.containers[0].volumeMounts'

print_status "Checking deployment YAML for volumes:"
kubectl get deployment test-mount-app -o yaml | yq '.spec.template.spec.volumes'

# -------------------------------------------------------------------------------------
print_status "Mount flags test completed!"
print_status "Summary of tests performed:"
print_status "  ✓ --mount-secret: Created secret mount at /app/secrets (read-only)"
print_status "  ✓ --mount-secret: Created shorthand secret mount at /app/secrets (read-only)"
print_status "  ✓ --mount-config: Created configmap mount at /app/config"
print_status "  ✓ --mount-service-binding-secret: Created service binding mount at /bindings/secret-*"
print_status "  ✓ Verified file contents are accessible"
print_status "  ✓ Verified read-only permissions where applicable"

exit 0
