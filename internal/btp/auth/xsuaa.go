package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
)

const authorizationEndpoint = "oauth/token"

type xsuaaErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type XSUAAToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	JTI         string `json:"jti"`
}

func GetOAuthToken(grantType, serverURL, username, password string) (*XSUAAToken, error) {
	urlBody := url.Values{}
	urlBody.Set("grant_type", grantType)

	request, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/%s", serverURL, authorizationEndpoint),
		strings.NewReader(urlBody.Encode()),
	)
	if err != nil {
		return nil, &clierror.Error{Message: "failed to build request", Details: err.Error(), Hints: []string{"Make sure the server URL in the config is correct."}}
	}
	defer request.Body.Close()

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(username, password)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, &clierror.Error{Message: "failed to get token from server", Details: err.Error()}
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, decodeAuthErrorResponse(response)
	}

	return decodeAuthSuccessResponse(response)
}

func decodeAuthSuccessResponse(response *http.Response) (*XSUAAToken, error) {
	token := XSUAAToken{}
	err := json.NewDecoder(response.Body).Decode(&token)
	if err != nil {
		return nil, &clierror.Error{Message: fmt.Sprintf("failed to decode response with Status %s", response.Status), Details: err.Error()}
	}

	return &token, nil
}

func decodeAuthErrorResponse(response *http.Response) error {
	errorData := xsuaaErrorResponse{}
	err := json.NewDecoder(response.Body).Decode(&errorData)
	if err != nil {
		return &clierror.Error{Message: "failed to decode error response", Details: err.Error()}
	}
	return &clierror.Error{Message: fmt.Sprintf("error response: %s", response.Status), Details: errorData.ErrorDescription}
}
