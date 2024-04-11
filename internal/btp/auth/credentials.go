package auth

import (
	"encoding/json"
	"os"

	"github.com/kyma-project/cli.v3/internal/clierror"
)

type CISCredentials struct {
	Endpoints       Endpoints `json:"endpoints"`
	GrantType       string    `json:"grant_type"`
	SapCloudService string    `json:"sap.cloud.service"`
	UAA             UAA       `json:"uaa"`
}

type Endpoints struct {
	AccountsServiceURL          string `json:"accounts_service_url"`
	CloudAutomationURL          string `json:"cloud_automation_url"`
	EntitlementsServiceURL      string `json:"entitlements_service_url"`
	EventsServiceURL            string `json:"events_service_url"`
	ExternalProviderRegistryURL string `json:"external_provider_registry_url"`
	MetadataServiceURL          string `json:"metadata_service_url"`
	OrderProcessingURL          string `json:"order_processing_url"`
	ProvisioningServiceURL      string `json:"provisioning_service_url"`
	SaasRegistryServiceURL      string `json:"saas_registry_service_url"`
}

type UAA struct {
	APIurl            string `json:"apiurl"`
	ClientID          string `json:"clientid"`
	ClientSecret      string `json:"clientsecret"`
	CredentialType    string `json:"credential-type"`
	IdentityZone      string `json:"identityzone"`
	IdentityZoneID    string `json:"identityzoneid"`
	SBurl             string `json:"sburl"`
	ServiceInstanceID string `json:"serviceInstanceId"`
	SubAccountID      string `json:"subaccountid"`
	TenantID          string `json:"tenantid"`
	TenantMode        string `json:"tenantmode"`
	UAADomain         string `json:"uaadomain"`
	URL               string `json:"url"`
	VerificationKey   string `json:"verificationkey"`
	XSAppname         string `json:"xsappname"`
	XSMasterAppName   string `json:"xsmasterappname"`
	ZoneID            string `json:"zoneid"`
}

func LoadCISCredentials(path string) (*CISCredentials, error) {
	credentialsBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, &clierror.Error{Message: "failed to read credentials file", Details: err.Error(), Hints: []string{"Make sure the path to the credentials file is correct."}}
	}

	credentials := CISCredentials{}
	err = json.Unmarshal(credentialsBytes, &credentials)
	if err != nil {
		return nil, &clierror.Error{Message: "failed to unmarshal file data", Details: err.Error(), Hints: []string{"Make sure the credentials file is in the correct format."}}
	}

	return &credentials, nil
}
