#!/usr/bin/env bash

set +o errexit
echo "####################"
echo "kubectl get pods -A"
echo "###################"
kubectl get pods -A

echo "####################"
echo "kubectl get svc -A"
echo "###################"
kubectl get svc -A

echo "####################"
echo "kubectl get apirule -A"
echo "###################"
kubectl get apirule -A

echo "########################################################"
echo "kubectl describe dockerregistry -n kyma-system default"
echo "########################################################"
kubectl describe dockerregistry -n kyma-system default

echo "########################################################"
echo "kubectl describe pod -l app=test-app"
echo "########################################################"
kubectl describe pod -l app=test-app

echo "########################################################"
echo "kubectl logs -l app=test-app"
echo "########################################################"
kubectl logs -l app=test-app
