.PHONY: prerequisites
prerequisites: 
	@test -s tf/.tfvars || \
	echo 'BTP_GLOBAL_ACCOUNT = "${BTP_GLOBAL_ACCOUNT}"\n\
	BTP_BOT_USER = "${BTP_BOT_USER}"\n\
	BTP_BOT_PASSWORD = "${BTP_BOT_PASSWORD}"\n\
	BTP_BACKEND_URL = "${BTP_BACKEND_URL}"\n\
	BTP_CUSTOM_IAS_TENANT = "${BTP_CUSTOM_IAS_TENANT}"\n\
	BTP_KYMA_SUBACCOUNT_ID = "${BTP_KYMA_SUBACCOUNT_ID}"\n\
	BTP_OBJECTSTORE_SUBACCOUNT_ID = "${BTP_OBJECTSTORE_SUBACCOUNT_ID}"\n\
	BTP_HANA_SUBACCOUNT_ID = "${BTP_HANA_SUBACCOUNT_ID}"' >> tf/.tfvars 
	terraform -chdir=tf init
	terraform -chdir=tf apply --auto-approve -var-file=.tfvars

.PHONY: e2e-test
e2e-test: |
	DOMAIN=${DOMAIN} ./integration-test-btp.sh

.PHONY: cleanup
cleanup:
	./kyma-cleanup.sh
