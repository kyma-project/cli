package cis

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kyma-project/cli.v3/internal/clierror"
)

const environmentsEndpoint = "provisioning/v1/environments"

type Labels struct {
	APIServerURL  string `json:"APIServerURL"`
	KubeconfigURL string `json:"KubeconfigURL"`
	Name          string `json:"Name"`
}

type environmentInstances struct {
	EnvironmentInstances []ProvisionResponse `json:"environmentInstances"`
}

func (c *LocalClient) GetKymaKubeconfig() (string, clierror.Error) {
	provisionURL := fmt.Sprintf("%s/%s", c.credentials.Endpoints.ProvisioningServiceURL, environmentsEndpoint)

	response, err := c.cis.get(provisionURL, requestOptions{})
	if err != nil {
		// TODO: finish
		return "", clierror.New(err.Error())
	}

	defer response.Body.Close()

	return decodeKubeconfig(response)

}

func decodeKubeconfig(response *http.Response) (string, clierror.Error) {
	envInstances := environmentInstances{}
	err := json.NewDecoder(response.Body).Decode(&envInstances)
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("failed to decode response"))
	}

	// we assume there can be only one Kyma environment in the BTP subaccount
	for _, env := range envInstances.EnvironmentInstances {
		if env.EnvironmentType == "kyma" {
			// parse labels to get kubeconfig URL
			labels := Labels{}
			err := json.Unmarshal([]byte(env.Labels), &labels)
			if err != nil {
				return "", clierror.Wrap(err, clierror.New("failed to unmarshal labels"))
			}

			kubeconfig, err := getKubeconfig(labels.KubeconfigURL)
			if err != nil {
				return "", clierror.Wrap(err, clierror.New("failed to get kubeconfig"))
			}
			return kubeconfig, nil
		}
	}

	return "", clierror.New("no Kyma environment found")
}

func getKubeconfig(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	kubeconfig, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(kubeconfig), nil
}
