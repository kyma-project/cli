package btp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

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

	token := XSUAAToken{}
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %s", err.Error())
	}

	return &token, nil
}
