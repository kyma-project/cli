#!/bin/bash

# -------------------------------------------------------------------------------------
echo "Cleanup"

kubectl delete apirules.gateway.kyma-project.io  bookstore
kubectl delete deployments.apps bookstore
kubectl delete svc bookstore

kubectl delete job hana-hdi-initjob

kubectl delete dockerregistries.operator.kyma-project.io -n kyma-system custom-dr
../../bin/kyma alpha module delete docker-registry
kubectl delete servicebindings.services.cloud.sap.com -A --all
kubectl delete serviceinstances.services.cloud.sap.com -A --all
kubectl delete secret -n kyma-system remote-service-manager-credentials

rm tf/btp-access-credentials-secret.yaml || true
rm tf/hana-admin-creds.json || true
rm tf/creds.json || true

rm tf/terraform.tfstate.backup || true
rm tf/terraform.tfstate || true
rm config.json || true


# TODO new command ?
# ../../bin/kyma alpha hana unmap --credentials-path hana-admin-api-binding.json
# -------------------------------------------------------------------------------------

exit 0