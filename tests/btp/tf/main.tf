terraform {
  required_providers {
    btp = {
      source  = "SAP/btp"
      version = ">= 1.18.1"
    }
  }
}

provider "btp" {
  globalaccount = var.BTP_GLOBAL_ACCOUNT
  cli_server_url = var.BTP_BACKEND_URL
  idp            = var.BTP_CUSTOM_IAS_TENANT
  username = var.BTP_BOT_USER
  password = var.BTP_BOT_PASSWORD
}

resource "btp_subaccount_entitlement" "hana-hdi" {
  subaccount_id = var.BTP_KYMA_SUBACCOUNT_ID
  service_name  = "hana"
  plan_name     = "hdi-shared"
}

# Existing binding for service manager in subaccount where shared objectstore instance is provided
data "btp_subaccount_service_binding" "remote-sm" {
  subaccount_id = var.BTP_OBJECTSTORE_SUBACCOUNT_ID
  name          = "kyma-cli-e2e-test"
}

# Existing binding for hana admin api in subaccount where SAP Hana Cloud instance is provided
data "btp_subaccount_service_binding" "remote-hana-admin" {
  subaccount_id = var.BTP_HANA_SUBACCOUNT_ID
  name          = "kyma-cli-e2e-test"
}

resource "local_sensitive_file" "service-manager-creds" {
  filename = "creds.json"
  content  = data.btp_subaccount_service_binding.remote-sm.credentials
}

resource "local_sensitive_file" "hana-admin-api-creds" {
  filename = "hana-admin-creds.json"
  content  = data.btp_subaccount_service_binding.remote-hana-admin.credentials
}
