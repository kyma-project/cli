set -e
make resolve
make build-linux
cp ./bin/kyma-linux /usr/local/bin/kyma
kyma provision k3d --ci
kyma deploy --ci
kyma undeploy --ci --timeout=10m0s
