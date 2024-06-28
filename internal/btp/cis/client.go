package cis

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	wwwAuthParser "github.com/gboddin/go-www-authenticate-parser"
	"github.com/kyma-project/cli.v3/internal/btp/auth"
)

type LocalClient struct {
	credentials *auth.CISCredentials
	cis         *httpClient
}

func NewLocalClient(credentials *auth.CISCredentials, token *auth.XSUAAToken) *LocalClient {
	return &LocalClient{
		credentials: credentials,
		cis:         newHTTPClient(token),
	}
}

type oauthTransport struct {
	token *auth.XSUAAToken
}

func (t *oauthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t.token.AccessToken))

	return http.DefaultTransport.RoundTrip(r)
}

type cisError struct {
	Code          int    `json:"code"`
	Message       string `json:"message"`
	Target        string `json:"target"`
	CorrelationID string `json:"correlationID"`
}

type cisErrorResponse struct {
	Error cisError `json:"error"`
}

type requestOptions struct {
	Body    io.Reader
	Headers map[string]string
	Query   map[string]string
}

type httpClient struct {
	client *http.Client
}

func newHTTPClient(token *auth.XSUAAToken) *httpClient {
	return &httpClient{
		client: &http.Client{
			Transport: &oauthTransport{
				token: token,
			},
		},
	}
}

func (c *httpClient) get(url string, options requestOptions) (*http.Response, error) {
	return c.genericRequest(http.MethodGet, url, options)
}

func (c *httpClient) post(url string, options requestOptions) (*http.Response, error) {
	return c.genericRequest(http.MethodPost, url, options)
}

func (c *httpClient) genericRequest(method string, url string, options requestOptions) (*http.Response, error) {
	request, err := http.NewRequest(method, url, options.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %s", err.Error())
	}

	for key, header := range options.Headers {
		request.Header.Add(key, header)
	}

	if len(options.Query) > 0 {
		q := request.URL.Query()
		for key, header := range options.Query {
			q.Add(key, header)
		}
		request.URL.RawQuery = q.Encode()
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to get data from server: %s", err.Error())
	}

	if response.StatusCode >= 400 {
		// error from response (status code higher or equal 400)
		return nil, c.buildResponseError(response)
	}

	return response, nil
}

func (c *httpClient) buildResponseError(response *http.Response) error {
	errorData := cisErrorResponse{}
	err := json.NewDecoder(response.Body).Decode(&errorData)
	if err == io.EOF {
		// error is possibly located in headers
		return c.buildErrorFromHeaders(response)
	}
	if err != nil {
		return fmt.Errorf("failed to decode error response with status '%s': %s", response.Status, err.Error())
	}

	return c.buildErrorFromBody(&errorData)
}

func (c *httpClient) buildErrorFromBody(errorData *cisErrorResponse) error {
	return errors.New(errorData.Error.Message)
}

func (c *httpClient) buildErrorFromHeaders(response *http.Response) error {
	wwwAuthHeaderString := response.Header.Get("Www-Authenticate")
	if wwwAuthHeaderString == "" {
		return fmt.Errorf("failed to parse http error for status: %s", response.Status)
	}

	wwwAuthHeader := wwwAuthParser.Parse(wwwAuthHeaderString)
	return fmt.Errorf("%s: %s", wwwAuthHeader.Params["error"], wwwAuthHeader.Params["error_description"])
}
