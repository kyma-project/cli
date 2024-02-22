package btp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type errorResponse struct {
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

func GetOAuthToken(credentials *CISCredentials) (*XSUAAToken, error) {
	urlBody := url.Values{}
	urlBody.Set("grant_type", credentials.GrantType)

	request, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/oauth/token", credentials.UAA.URL),
		strings.NewReader(urlBody.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %s", err.Error())
	}
	defer request.Body.Close()

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(
		credentials.UAA.ClientID,
		credentials.UAA.ClientSecret,
	)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to get token from server: %s", err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, decodeAuthErrorResponse(resp)
	}

	return decodeAuthSuccessResponse(resp)
}

func decodeAuthSuccessResponse(resp *http.Response) (*XSUAAToken, error) {
	token := XSUAAToken{}
	err := json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %s", err.Error())
	}

	return &token, nil
}

func decodeAuthErrorResponse(resp *http.Response) error {
	errorData := errorResponse{}
	err := json.NewDecoder(resp.Body).Decode(&errorData)
	if err != nil {
		return fmt.Errorf("failed to decode error response: %s", err.Error())
	}
	return fmt.Errorf("error response: %s: %s", errorData.Error, errorData.ErrorDescription)
}
