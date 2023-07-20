set -e
apk add nodejs npm
make resolve
make build-linux
cp ./bin/kyma-linux /usr/local/bin/kyma
kyma provision k3d --ci
kyma deploy --ci
make -C "../kyma/tests/fast-integration" "ci"
#Uncomment when https://github.com/kyma-project/cli/issues/1701 is done
#kyma undeploy --ci --timeout=10m0s
