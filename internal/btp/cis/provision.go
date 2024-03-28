package cis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

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
	Name   string `json:"name"`
	Region string `json:"region"`
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

func (c *LocalClient) Provision(pe *ProvisionEnvironment) (*ProvisionResponse, *clierror.Error) {
	reqData, err := json.Marshal(pe)
	if err != nil {
		return nil, &clierror.Error{Message: err.Error()}
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
		return nil, &clierror.Error{Message: err.Error()}
	}
	defer response.Body.Close()

	return decodeProvisionSuccessResponse(response)
}

func decodeProvisionSuccessResponse(response *http.Response) (*ProvisionResponse, *clierror.Error) {
	provisionResponse := ProvisionResponse{}
	err := json.NewDecoder(response.Body).Decode(&provisionResponse)
	if err != nil {
		return nil, &clierror.Error{Message: "failed to decode response", Details: err.Error()}
	}

	return &provisionResponse, nil
}
