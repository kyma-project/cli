.PHONY: prerequisites
prerequisites: 
	terraform -chdir=tf init
	terraform -chdir=tf apply --auto-approve -var-file=.tfvars

.PHONY: e2e-test
e2e-test: |
	DOMAIN=${DOMAIN} ACTIONS_ID_TOKEN_REQUEST_URL=${ACTIONS_ID_TOKEN_REQUEST_URL} ./integration-test-btp.sh

.PHONY: cleanup
cleanup:
	./kyma-cleanup.sh
