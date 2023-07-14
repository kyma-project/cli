set -e
apk add nodejs npm
echo "Bump reconciler version used by the Kyma CLI"
make resolve
echo "Run unit-tests for kyma kyma"
make test
echo "Building Kyma CLI"
make build-linux
echo "Committing reconciler bump"
git_status=$(git status --porcelain)
if [[ "${git_status}" != "" ]]; then
  git commit -am 'bump reconciler version'
fi
echo "Copying Kyma binary to usr/local/bin"
cp ./bin/kyma-linux /usr/local/bin/kyma
echo "Provisioning k3d Kubernetes runtime"
kyma provision k3d --ci
echo "Installing Kyma"
kyma deploy --ci
echo "Running fast-integration tests"
make -C "../kyma/tests/fast-integration" "ci"
echo "Uninstalling Kyma"
kyma undeploy --ci --timeout=10m0s
export KYMA_CLI_UNSTABLE_BUCKET=gs://kyma-cli-unstable
export UNSTABLE=true
echo "Publishing new unstable builds to $KYMA_CLI_UNSTABLE_BUCKET"
make ci-main
echo "all done"