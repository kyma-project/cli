set -e
apk add nodejs npm
echo "Bump reconciler version used by the Kyma CLI"
make resolve
make build-linux
cp ./bin/kyma-linux /usr/local/bin/kyma
kyma provision k3d --ci
kyma deploy --ci
make -C "../kyma/tests/fast-integration" "ci"
kyma undeploy --ci --timeout=10m0s
