set -e
apk add nodejs npm
make resolve
make build-linux
cp ./bin/kyma-linux /usr/local/bin/kyma
kyma provision k3d --ci
kyma deploy --ci
make -C "../kyma/tests/fast-integration" "ci"
sleep 10
kyma undeploy --ci --timeout=10m0s
