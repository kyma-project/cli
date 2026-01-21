#!/bin/bash
set -e -o pipefail

echo -e "\n--------------------------------------------------------------------------------------\n"
echo "Running kyma integration tests uing connected managed kyma runtime"

echo -e "\n--------------------------------------------------------------------------------------\n"
echo -e "Step1: Generating temporary access for new service account\n"

../../bin/kyma alpha kubeconfig generate --clusterrole cluster-admin --serviceaccount test-sa --output /tmp/kubeconfig.yaml --time 2h --cluster-wide --namespace default

export KUBECONFIG="/tmp/kubeconfig.yaml"
if [[ $(kubectl config view --minify --raw | yq '.users[0].name') != 'test-sa' ]]; then
    exit 1
fi
echo "Running test in user context of: $(kubectl config view --minify --raw | yq '.users[0].name')"

echo "Waiting for KCP to propagate release metadata for kyma modules..."
while ! kubectl get crd modulereleasemetas.operator.kyma-project.io; do echo "Waiting for CRD modulereleasemetas..."; sleep 1; done
kubectl wait --for condition=established crd/modulereleasemetas.operator.kyma-project.io
while ! kubectl get modulereleasemetas.operator.kyma-project.io serverless --namespace kyma-system; do echo "Waiting for serverless release metadata..."; sleep 1; done


echo -e "\n--------------------------------------------------------------------------------------\n"
echo -e "Step2: Manage modules \n"

../../bin/kyma module catalog
../../bin/kyma module add serverless --default-cr
echo "..waiting for serverless module to be installed"
while ! kubectl get crd serverlesses.operator.kyma-project.io; do echo "Waiting for CRD serverless..."; sleep 1; done
kubectl wait --for condition=established crd/serverlesses.operator.kyma-project.io
while ! kubectl get serverlesses.operator.kyma-project.io default --namespace kyma-system; do echo "Waiting for serverless..."; sleep 1; done
kubectl wait --for condition=Installed serverlesses.operator.kyma-project.io/default -n kyma-system --timeout=360s

../../bin/kyma module list

../../bin/kyma module delete serverless --auto-approve
echo "..waiting for serverless module to be deleted"
kubectl wait --for=delete deployment/serverless-operator -n kyma-system --timeout=360s

echo -e "\n--------------------------------------------------------------------------------------\n"
echo -e "Step3: Connecting to a service manager from remote BTP subaccount\n"

# https://help.sap.com/docs/btp/sap-business-technology-platform/namespace-level-mapping?locale=en-US
( cd tf ; curl https://raw.githubusercontent.com/kyma-project/btp-manager/main/hack/create-secret-file.sh | bash -s operator remote-service-manager-credentials )
kubectl create -f tf/btp-access-credentials-secret.yaml || true

echo -e "\n--------------------------------------------------------------------------------------\n"
echo -e "Step4: Create service instance reference to a shared object-store service instance\n"

echo "Waiting for CRD btp operator"
while ! kubectl get crd btpoperators.operator.kyma-project.io; do echo "Waiting for CRD btp operator..."; sleep 1; done
kubectl wait --for condition=established crd/btpoperators.operator.kyma-project.io
while ! kubectl get btpoperators.operator.kyma-project.io btpoperator --namespace kyma-system; do echo "Waiting for btpoperator..."; sleep 1; done
kubectl wait --for condition=Ready btpoperators.operator.kyma-project.io/btpoperator -n kyma-system --timeout=180s


# TODO - change after btp operator commands are extracted as btp module cli extension
../../bin/kyma alpha reference-instance \
    --btp-secret-name remote-service-manager-credentials \
    --namespace kyma-system \
    --offering-name objectstore \
    --plan-selector standard \
    --reference-name object-store-reference
kubectl apply -n kyma-system -f ./k8s-resources/object-store-binding.yaml

while ! kubectl get secret object-store-reference-binding --namespace kyma-system; do echo "Waiting for object-store-reference-binding secret..."; sleep 5; done


# Enable Docker Registry
echo -e "\n--------------------------------------------------------------------------------------\n"
echo -e "Step5: Enable Docker Registry community module (with persistent BTP based storage)\n"
../../bin/kyma module pull docker-registry
../../bin/kyma module add default/docker-registry-0.10.0 --cr-path k8s-resources/custom-docker-registry.yaml --auto-approve

echo "..waiting for docker registry"
kubectl wait --for condition=Installed dockerregistries.operator.kyma-project.io/custom-dr -n kyma-system --timeout=360s

while ! kubectl get secret dockerregistry-config-external --namespace kyma-system; do echo "Waiting for dockerregistry-config-external secret..."; sleep 5; done

dr_external_url=$(../../bin/kyma registry config-external --push-reg-addr)
dr_internal_pull_url=$(../../bin/kyma registry config-internal --pull-reg-addr)
dr_username=$(kubectl get secrets -n kyma-system dockerregistry-config-external -o jsonpath={.data.username} | base64 -d)
dr_password=$(kubectl get secrets -n kyma-system dockerregistry-config-external -o jsonpath={.data.password} | base64 -d)

../../bin/kyma registry config-external --output config.json

echo "Docker Registry enabled (URLs: $dr_external_url, $dr_internal_pull_url)"
echo "config.json for docker CLI access generated"
cat config.json

echo -e "\n--------------------------------------------------------------------------------------\n"
echo -e "Step6: Map SAP Hana DB instance with Kyma runtime\n"

../../bin/kyma alpha hana map --credentials-path tf/hana-admin-creds.json

echo -e "\n--------------------------------------------------------------------------------------\n"
echo -e "Step7: Pack & push hdi-deploy image\n"

docker version

# build hdi-deploy via pack and push it via docker CLI (external url)
pack build hdi-deploy:latest -p sample-http-db-nodejs/hdi-deploy -B paketobuildpacks/builder:base
docker tag hdi-deploy:latest $dr_external_url/hdi-deploy:latest

# check HTTP reachability of registry v2 endpoint before pushing
curl -v https://$dr_external_url/v2/_catalog -u $dr_username:$dr_password --max-time 20 || true

# for test push without docker config use:
docker --config . push $dr_external_url/hdi-deploy:latest

echo -e "\n--------------------------------------------------------------------------------------\n"
echo -e "Step8: Deploy hdi-deploy (hdi instance & binding, run db initialisation)\n"

echo "Initialising db binding..."
kubectl set image -f ./k8s-resources/db/books-hdi-initjob-template.yaml bookstore-db=$dr_internal_pull_url/hdi-deploy:latest --local -o yaml > ./k8s-resources/db/books-hdi-initjob.yaml
kubectl apply -k ./k8s-resources/db
echo "Waiting for hana-init-job to complete..."
kubectl wait --for condition=Complete jobs/hana-hdi-initjob --timeout=360s 
echo "Bookstore db initialised" 

echo -e "\n--------------------------------------------------------------------------------------\n"
echo -e "Step9: Pushing bookstore app\n"

kubectl label namespace default istio-injection=enabled --overwrite

# deploy bookstore app via kyma push

../../bin/kyma app push --name bookstore --expose --container-port 3000 --mount-service-binding-secret hana-hdi-binding --code-path sample-http-db-nodejs/bookstore

echo -e ""
kubectl wait --for condition=Available deployment bookstore --timeout=60s
kubectl wait --for='jsonpath={.status.state}=Ready' apirules.gateway.kyma-project.io/bookstore

echo -e "\n--------------------------------------------------------------------------------------\n"
echo -e "Step10: Verify bookstore app\n"
sleep 5

DOMAIN=$(kubectl get configmaps -n kube-system shoot-info -o=jsonpath='{.data.domain}')
response=$(curl https://bookstore.$DOMAIN/v1/books)
echo "HTTP response from sample app: $response"

if [[ $response != '[{"id":1,"title":"Dune","author":"Frank Herbert"},{"id":2,"title":"Pippi Goes on Board","author":"Astrid Lindgren"}]' ]]; then
    exit 1
fi

# TODO enable after https://github.com/kyma-project/docker-registry/issues/447 is fixed
# echo -e "\n--------------------------------------------------------------------------------------\n"
# echo -e "Step11: Uninstalling community module \n"
# ../../bin/kyma module delete default/docker-registry-0.10.0 

exit 0