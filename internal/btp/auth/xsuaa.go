package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
		return nil, fmt.Errorf("failed to build request: %s", err.Error())
	}
	defer request.Body.Close()

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(username, password)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to get token from server: %s", err.Error())
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
		return nil, fmt.Errorf("failed to decode response with Status '%s': %s", response.Status, err.Error())
	}

	return &token, nil
}

func decodeAuthErrorResponse(response *http.Response) error {
	errorData := xsuaaErrorResponse{}
	err := json.NewDecoder(response.Body).Decode(&errorData)
	if err != nil {
		return fmt.Errorf("failed to decode error response: %s", err.Error())
	}
	return fmt.Errorf("error response: %s: %s", response.Status, errorData.ErrorDescription)
}
