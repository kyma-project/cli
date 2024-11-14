package hana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-project/cli.v3/internal/clierror"
)

func MapInstance(baseURL, clusterID, hanaID, token string) clierror.Error {
	return mapInstance(fmt.Sprintf("https://%s", baseURL), clusterID, hanaID, token)
}

func mapInstance(baseURL, clusterID, hanaID, token string) clierror.Error {
	client := &http.Client{}

	requestData := HanaMapping{
		Platform:  "kubernetes",
		PrimaryID: clusterID,
	}

	requestString, err := json.Marshal(requestData)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to marshal mapping request"))
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/inventory/v2/serviceInstances/%s/instanceMappings", baseURL, hanaID), bytes.NewBuffer(requestString))
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create mapping request"))
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(request)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to post mapping request"))
	}

	// server sends status Created when mapping is created, and 200 if it already exists
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return clierror.Wrap(fmt.Errorf("status code: %d", resp.StatusCode), clierror.New("unexpected status code"))
	}

	return nil
}
