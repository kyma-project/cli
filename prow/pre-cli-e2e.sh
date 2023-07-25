set -e
apk add nodejs npm
make resolve
make build-linux
cp ./bin/kyma-linux /usr/local/bin/kyma
kyma provision k3d --ci
kyma deploy --ci
#make -C "../kyma/tests/fast-integration" "ci"
kubectl get pvc -A
kyma undeploy --ci --timeout=15m0s
