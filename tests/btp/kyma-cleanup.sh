#!/bin/bash

# -------------------------------------------------------------------------------------
echo "Cleanup"

kubectl delete apirules.gateway.kyma-project.io  bookstore
kubectl delete deployments.apps bookstore
kubectl delete svc bookstore

kubectl delete job hana-hdi-initjob

../../bin/kyma module delete default/docker-registry-0.10.0 --auto-approve
kubectl delete servicebindings.services.cloud.sap.com -A --all
kubectl delete serviceinstances.services.cloud.sap.com -A --all
kubectl delete secret -n kyma-system remote-service-manager-credentials

rm tf/btp-access-credentials-secret.yaml || true



# TODO new command ?
# ../../bin/kyma alpha hana unmap --credentials-path hana-admin-api-binding.json
# -------------------------------------------------------------------------------------

exit 0