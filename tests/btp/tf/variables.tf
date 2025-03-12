variable "BTP_GLOBAL_ACCOUNT" {
  type        = string
  description = "Global account name"
  default     = "global-account-guid"
}

variable "BTP_BOT_USER" {
  type        = string
  description = "Bot account name"
  default     = "email@domain.com"
}

variable "BTP_BOT_PASSWORD" {
  type        = string
  description = "Bot account password"
  default     = "password"
  sensitive = true
}

variable "BTP_BACKEND_URL" {
  type        = string
  description = "BTP backend URL"
  default     = "https://cli.btp.cloud.sap"
}

variable "BTP_CUSTOM_IAS_TENANT" {
  type        = string
  description = "Custom IAS tenant"
  default     = "custom-tenant"
}

variable "BTP_KYMA_SUBACCOUNT_ID" {
  type        = string
  description = "Subaccount ID"
}

variable "BTP_OBJECTSTORE_SUBACCOUNT_ID" {
  type        = string
  description = "Subaccount ID"
}

variable "BTP_HANA_SUBACCOUNT_ID" {
  type        = string
  description = "Subaccount ID"
}