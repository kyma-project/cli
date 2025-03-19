.PHONY: prerequisites
prerequisites: 
	terraform -chdir=tf init
	terraform -chdir=tf apply --auto-approve -var-file=.tfvars

.PHONY: e2e-test
e2e-test: |
	DOMAIN=${DOMAIN} ./integration-test-btp.sh

.PHONY: cleanup
cleanup:
	./kyma-cleanup.sh
