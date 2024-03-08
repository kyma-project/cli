package cis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const entitlementsEndpoint = "entitlements/v1/subaccountServicePlans"

type EntitlementsSubaccountServicePlans struct {
	SubaccountServicePlans SubaccountServicePlans `json:"subaccountServicePlans"`
}

type SubaccountServicePlans struct {
	AssignmentInfo  AssignmentInfo `json:"assignmentInfo"`
	ServiceName     string         `json:"serviceName"`
	ServicePlanName string         `json:"servicePlanName"`
}

type AssignmentInfo struct {
	//Amount         int       `json:"amount"`
	Enable bool `json:"enable"`
	//Resources      Resources `json:"resources"`
	SubaccountGUID string `json:"subaccountGUID"`
}

//type Resources struct {
//	//ResourceData          ResourceData `json:"resourceData"`
//	ResourceName          string `json:"resourceName"`
//	ResourceProvider      string `json:"resourceProvider"`
//	ResourceTechnicalName string `json:"resourceTechnicalName"`
//	ResourceType          string `json:"resourceType"`
//}

func (c *CentralClient) Entitlements(servicePlans *EntitlementsSubaccountServicePlans) (*http.Response, error) {
	reqData, err := json.Marshal(servicePlans)
	if err != nil {
		return nil, err
	}

	provisionURL := fmt.Sprintf("%s/%s", c.credentials.Endpoints.EntitlementsServiceURL, entitlementsEndpoint)
	options := requestOptions{
		Body: bytes.NewBuffer(reqData),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	response, err := c.cis.put(provisionURL, options)
	if err != nil {
		return nil, fmt.Errorf("failed to provision: %s", err.Error())
	}
	defer response.Body.Close()

	return response, nil
}
