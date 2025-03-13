package github

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-project/cli.v3/internal/clierror"
)

func GetToken(url, requestToken, audience string) (string, clierror.Error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("failed to create request"))
	}
	if audience != "" {
		q := request.URL.Query()
		q.Add("audience", audience)
		request.URL.RawQuery = q.Encode()
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", requestToken))
	request.Header.Add("Accept", "application/json; api-version=2.0")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("failed to get token from Github"))
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return "", clierror.New(fmt.Sprintf("Invalid server response: %d", response.StatusCode))
	}

	tokenData := struct {
		Count int    `json:"count"`
		Value string `json:"value"`
	}{}
	err = json.NewDecoder(response.Body).Decode(&tokenData)
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("failed to decode token response"))
	}
	return tokenData.Value, nil
}
