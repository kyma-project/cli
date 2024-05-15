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

func GetOAuthToken(grantType, serverURL, username, password string) (*XSUAAToken, clierror.Error) {
	urlBody := url.Values{}
	urlBody.Set("grant_type", grantType)

	request, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/%s", serverURL, authorizationEndpoint),
		strings.NewReader(urlBody.Encode()),
	)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.Message("failed to build request"), clierror.Hints("Make sure the server URL in the config is correct."))
	}
	defer request.Body.Close()

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(username, password)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.Message("failed to get token from server"))
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, decodeAuthErrorResponse(response)
	}

	return decodeAuthSuccessResponse(response)
}

func decodeAuthSuccessResponse(response *http.Response) (*XSUAAToken, clierror.Error) {
	token := XSUAAToken{}
	err := json.NewDecoder(response.Body).Decode(&token)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.MessageF("failed to decode response with Status %s", response.Status))
	}

	return &token, nil
}

func decodeAuthErrorResponse(response *http.Response) clierror.Error {
	errorData := xsuaaErrorResponse{}
	err := json.NewDecoder(response.Body).Decode(&errorData)
	if err != nil {
		return clierror.Wrap(err, clierror.Message("failed to decode error response"))
	}
	return clierror.Wrap(errorData.ErrorDescription, clierror.MessageF("error response: %s", response.Status))
}
