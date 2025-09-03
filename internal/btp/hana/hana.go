package hana

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kyma-project/cli.v3/internal/clierror"
)

func GetID(baseURL, token string) (string, clierror.Error) {
	return getID(fmt.Sprintf("https://%s", baseURL), token)
}

func getID(baseURL, token string) (string, clierror.Error) {
	client := &http.Client{}
	requestGet, err := http.NewRequest("GET", fmt.Sprintf("%s/inventory/v2/serviceInstances", baseURL), nil)
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("failed to create a get SAP Hana instances request"))
	}
	requestGet.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	response, err := client.Do(requestGet)
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("failed to get SAP Hana instances"))
	}

	if response.StatusCode != http.StatusOK {
		return "", clierror.New(fmt.Sprintf("unexpected status code: %d", response.StatusCode))
	}
	hanaInstanceBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("failed to read SAP Hana instances from the response"))
	}
	hanaInstance := HanaInstance{}
	err = json.Unmarshal(hanaInstanceBytes, &hanaInstance)
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("failed to parse SAP Hana instances from the response"))
	}
	if len(hanaInstance.Data) == 0 {
		return "", clierror.New("no SAP Hana instances found in the response")
	}
	return hanaInstance.Data[0].ID, nil
}
