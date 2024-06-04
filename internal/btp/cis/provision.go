package cis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
)

const provisionEndpoint = "provisioning/v1/environments"

type ProvisionEnvironment struct {
	// Description     string         `json:"description,omitempty"`
	EnvironmentType string `json:"environmentType"`
	// LandscapeLabel  string         `json:"landscapeLabel,omitempty"`
	Name string `json:"name"`
	// Origin          string         `json:"origin,omitempty"`
	Parameters KymaParameters `json:"parameters"`
	PlanName   string         `json:"planName"`
	// ServiceName     string         `json:"serviceName,omitempty"`
	// TechnicalKey    string         `json:"technicalKey,omitempty"`
	User string `json:"user"`
}

type KymaParameters struct {
	Administrators *[]string                 `json:"administrators,omitempty"`
	AutoScalerMax  *int                      `json:"autoScalerMax,omitempty"`
	AutoScalerMin  *int                      `json:"autoScalerMin,omitempty"`
	MachineType    *string                   `json:"machineType,omitempty"`
	Modules        *KymaParametersModules    `json:"modules,omitempty"`
	Name           string                    `json:"name"`
	Networking     *KymaParametersNetworking `json:"networking,omitempty"`
	Oidc           *KymaParametersOidc       `json:"oidc,omitempty"`
	Region         string                    `json:"region"`
}

type KymaParametersModules struct {
	Default bool `json:"default,omitempty"`
	List    []struct {
		Name                 string `json:"name"`
		Channel              string `json:"channel,omitempty"`
		CustomResourcePolicy string `json:"customResourcePolicy,omitempty"`
	} `json:"list,omitempty"`
}

type KymaParametersNetworking struct {
	Nodes string `json:"nodes"`
}

type KymaParametersOidc struct {
	ClientID      string   `json:"clientID"`
	GroupsClaim   string   `json:"groupsClaim,omitempty"`
	IssuerURL     string   `json:"issuerURL"`
	SigningAlgs   []string `json:"signingAlgs,omitempty"`
	UsernameClaim string   `json:"usernameClaim,omitempty"`
}

type ProvisionResponse struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	BrokerID          string `json:"brokerId"`
	GlobalAccountGUID string `json:"globalAccountGUID"`
	SubaccountGUID    string `json:"subaccountGUID"`
	TenantID          string `json:"tenantId"`
	ServiceID         string `json:"serviceId"`
	PlanID            string `json:"planId"`
	DashboardURL      string `json:"dashboardUrl"`
	Operation         string `json:"operation"`
	Parameters        string `json:"parameters"`
	Labels            string `json:"labels"`
	// CustomLabels      struct {} `json:"customLabels"`
	Type            string `json:"type"`
	Status          string `json:"status"`
	EnvironmentType string `json:"environmentType"`
	PlatformID      string `json:"platformId"`
	CreatedDate     int64  `json:"createdDate"`
	ModifiedDate    int64  `json:"modifiedDate"`
	State           string `json:"state"`
	StateMessage    string `json:"stateMessage"`
	ServiceName     string `json:"serviceName"`
	PlanName        string `json:"planName"`
}

func (c *LocalClient) Provision(pe *ProvisionEnvironment) (*ProvisionResponse, clierror.Error) {
	reqData, err := json.Marshal(pe)
	if err != nil {
		return nil, clierror.New(err.Error())
	}

	provisionURL := fmt.Sprintf("%s/%s", c.credentials.Endpoints.ProvisioningServiceURL, provisionEndpoint)
	options := requestOptions{
		Body: bytes.NewBuffer(reqData),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	response, err := c.cis.post(provisionURL, options)
	if err != nil {
		var hints []string

		if strings.Contains(err.Error(), "404 Not Found") {
			hints = append(hints, "check if the CIS config file contains correct provisioning_service_url endpoint")
		}
		if strings.Contains(err.Error(), "insufficient_scope") {
			hints = append(hints, "check if the CIS instance plan is set to local")
		}
		if strings.Contains(err.Error(), "User is unauthorized for this operation") {
			hints = append(hints, "check if subaccount have enabled Kyma entitlement with correct plan")
		}

		return nil, clierror.New(err.Error(), hints...)
	}
	defer response.Body.Close()

	return decodeProvisionSuccessResponse(response)
}

func decodeProvisionSuccessResponse(response *http.Response) (*ProvisionResponse, clierror.Error) {
	provisionResponse := ProvisionResponse{}
	err := json.NewDecoder(response.Body).Decode(&provisionResponse)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("failed to decode response"))
	}

	return &provisionResponse, nil
}
