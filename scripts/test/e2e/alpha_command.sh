#!/usr/bin/env bash
set -e

KYMA_PROJECT_DIR="/home/prow/go/src/github.com/kyma-project"
KLM_SOURCES_DIR="$KYMA_PROJECT_DIR/lifecycle-manager/"
export KCP_KUBECONFIG="$KYMA_PROJECT_DIR/kcp.yaml"
export SKR_KUBECONFIG="$KYMA_PROJECT_DIR/skr.yaml"

#shellcheck source=prow/scripts/lib/log.sh
source "$KYMA_PROJECT_DIR/test-infra/prow/scripts/lib/log.sh"
#shellcheck source=prow/scripts/lib/kyma.sh
source "$KYMA_PROJECT_DIR/test-infra/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "$KYMA_PROJECT_DIR/test-infra/prow/scripts/lib/docker.sh"


#REMOVE below if not used
#shellcheck source=prow/scripts/lib/utils.sh
source "$KYMA_PROJECT_DIR/test-infra/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gardener/gardener.sh
source "$KYMA_PROJECT_DIR/test-infra/prow/scripts/lib/gardener/gardener.sh"



function prereq_install() {
  log::info "Install latest released Kyma CLI"
  kyma::install_cli

  log::info "Install k3d"
  wget -q -O - https://raw.githubusercontent.com/rancher/k3d/main/install.sh | bash

  log::info "Install istioctl"
  export ISTIO_VERSION="1.17.1"
  wget -q "https://github.com/istio/istio/releases/download/${ISTIO_VERSION}/istioctl-${ISTIO_VERSION}-linux-amd64.tar.gz"
  tar -C /usr/local/bin -xzf "istioctl-${ISTIO_VERSION}-linux-amd64.tar.gz"
  export PATH=$PATH:/usr/local/bin/istioctl
  istioctl version --remote=false
  export ISTIOCTL_PATH=/usr/local/bin/istioctl


}

function prereq_test() {
  command -v k3d >/dev/null 2>&1 || { echo >&2 "k3d not found"; exit 1; }
  command -v kyma >/dev/null 2>&1 || { echo >&2 "kyma not found"; exit 1; }
  command -v istioctl >/dev/null 2>&1 || { echo >&2 "istioctl not found"; exit 1; }
  command -v kubectl >/dev/null 2>&1 || { echo >&2 "kubectl not found"; exit 1; }
}

function provision_k3d() {
  log::info "Provisioning k3d cluster"

  k3d version
  kyma provision k3d --name=kcp -p 9080:80@loadbalancer -p 9443:443@loadbalancer --ci
  k3d kubeconfig get kcp > $KCP_KUBECONFIG
  log::success "Kyma K3d cluster provisioned: kcp"

  k3d cluster create skr -p 10080:80@loadbalancer -p 10443:443@loadbalancer
  k3d kubeconfig get skr > $SKR_KUBECONFIG
  log::success "Base K3d cluster provisioned: skr"


  FILE=/etc/hosts
  if [ -f "$FILE" ]; then
      echo "127.0.0.1 k3d-kcp-registry" >> $FILE
  else
      log::error "$FILE does not exist."
      exit 1
  fi

  log::info "/etc/hosts file patched"
}

function installKcpComponents() {
  export KUBECONFIG=$KCP_KUBECONFIG

  istioctl install --set profile=demo -y

  kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.11.0/cert-manager.yaml
}

prereq_install
prereq_test

docker::start

provision_k3d

installKcpComponents

pwd
cd $KLM_SOURCES_DIR
pwd
if make local-deploy-with-watcher IMG=europe-docker.pkg.dev/kyma-project/prod/lifecycle-manager:latest; then
  log::success "KLM deployed successfully"
else
  log::error "Deploy encountered some error, will retry"
  sleep 20
  make local-deploy-with-watcher IMG=europe-docker.pkg.dev/kyma-project/prod/lifecycle-manager:latest
fi

kubectl get crd
kubectl get crd kymas.operator.kyma-project.io -oyaml

cd tests/e2e_test
make test

