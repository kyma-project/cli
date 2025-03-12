terraform {
  required_providers {
    btp = {
      source  = "SAP/btp"
      version = ">= 1.6.0"
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

data "btp_subaccount_service_binding" "kyma-cli-e2e-test" {
  subaccount_id = "ad8a3201-39b4-4c6e-a67a-9a29a2099d2d"
  name          = "kyma-cli-e2e-test"
}

resource "local_sensitive_file" "creds" {
  filename = "creds.json"
  content  = data.btp_subaccount_service_binding.kyma-cli-e2e-test.credentials
}
