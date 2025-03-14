.PHONY: prerequisites
prerequisites: 
	echo 'BTP_GLOBAL_ACCOUNT = "${BTP_GLOBAL_ACCOUNT}"' >> tf/.tfvars 
	echo 'BTP_BOT_USER = "${BTP_BOT_USER}"' >> tf/.tfvars 
	echo 'BTP_BOT_PASSWORD = "${BTP_BOT_PASSWORD}"' >> tf/.tfvars 
	echo 'BTP_BACKEND_URL = "${BTP_BACKEND_URL}"' >> tf/.tfvars 
	echo 'BTP_CUSTOM_IAS_TENANT = "${BTP_CUSTOM_IAS_TENANT}"' >> tf/.tfvars 
	echo 'BTP_KYMA_SUBACCOUNT_ID = "${BTP_KYMA_SUBACCOUNT_ID}"' >> tf/.tfvars 
	echo 'BTP_OBJECTSTORE_SUBACCOUNT_ID = "${BTP_OBJECTSTORE_SUBACCOUNT_ID}"' >> tf/.tfvars 
	echo 'BTP_HANA_SUBACCOUNT_ID = "${BTP_HANA_SUBACCOUNT_ID}"' >> tf/.tfvars 

.PHONY: e2e-test
e2e-test: |
	terraform -chdir=tf init
	terraform -chdir=tf apply --auto-approve -var-file=.tfvars
	DOMAIN=${DOMAIN} ./integration-test-btp.sh

.PHONY: cleanup
cleanup:
	./kyma-cleanup.sh
