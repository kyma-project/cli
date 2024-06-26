package cis

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-project/cli.v3/internal/btp/auth"
	"github.com/stretchr/testify/require"
)

func Test_oauthTransport_RoundTrip(t *testing.T) {

	t.Parallel()

	t.Run("client bearer authorization", func(t *testing.T) {
		svr := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			require.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		}))
		defer svr.Close()

		req, err := http.NewRequest(http.MethodPost, svr.URL, nil)
		require.NoError(t, err)
		clientTransport := oauthTransport{
			token: &auth.XSUAAToken{
				AccessToken: "token",
			},
		}
		_, err = clientTransport.RoundTrip(req)
		require.NoError(t, err)
	})
}

func Test_GenericRequest(t *testing.T) {

	t.Parallel()

	testEmptyServer := httptest.NewServer(http.HandlerFunc(fixGenericRequestHandler(t, requestOptions{})))
	defer testEmptyServer.Close()

	testServer := httptest.NewServer(http.HandlerFunc(fixGenericRequestHandler(t, fixRequestOptions())))
	defer testServer.Close()

	testErrorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(415)
	}))
	defer testErrorServer.Close()

	t.Run("simple GET request", func(t *testing.T) {
		c := httpClient{client: http.DefaultClient}

		response, err := c.genericRequest(http.MethodGet, testEmptyServer.URL, requestOptions{})

		require.NoError(t, err)
		require.NotNil(t, response)
		require.Equal(t, 200, response.StatusCode)

		_ = response.Body.Close()
	})

	t.Run("simple POST request with additional data", func(t *testing.T) {
		c := httpClient{client: http.DefaultClient}

		response, err := c.genericRequest(http.MethodPost, testServer.URL, fixRequestOptions())

		require.NoError(t, err)
		require.NotNil(t, response)
		require.Equal(t, 200, response.StatusCode)

		_ = response.Body.Close()
	})

	t.Run("build request error becuse of wrong method name", func(t *testing.T) {
		c := httpClient{client: http.DefaultClient}

		response, err := c.genericRequest("DoEsNoTeXiSt)", testServer.URL, requestOptions{})

		require.Equal(t, errors.New("failed to build request: net/http: invalid method \"DoEsNoTeXiSt)\""), err)
		require.Nil(t, response)
	})

	t.Run("cant reach server by URL error", func(t *testing.T) {
		c := httpClient{client: http.DefaultClient}

		response, err := c.genericRequest(http.MethodGet, "does-not-exist", requestOptions{})

		require.Equal(t, errors.New("failed to get data from server: Get \"does-not-exist\": unsupported protocol scheme \"\""), err)
		require.Nil(t, response)
	})

	t.Run("handle 415 response status", func(t *testing.T) {
		c := httpClient{client: http.DefaultClient}

		response, err := c.genericRequest(http.MethodGet, testErrorServer.URL, requestOptions{})

		require.Equal(t, errors.New("failed to parse http error for status: 415 Unsupported Media Type"), err)
		require.Nil(t, response)
	})
}

func Test_httpClient_buildResponseError(t *testing.T) {

	t.Parallel()

	tests := []struct {
		name        string
		response    *http.Response
		expectedErr error
	}{
		{
			name: "build error from status",
			response: &http.Response{
				Status: "Unauthorized",
				Body:   io.NopCloser(strings.NewReader("")),
			},
			expectedErr: errors.New("failed to parse http error for status: Unauthorized"),
		},
		{
			name: "build error from header",
			response: &http.Response{
				Status: "Unauthorized",
				Header: http.Header{
					"Www-Authenticate": []string{"Bearer error=\"error\",error_description=\"description\""},
				},
				Body: io.NopCloser(strings.NewReader("")),
			},
			expectedErr: errors.New("error: description"),
		},
		{
			name: "build error from body",
			response: &http.Response{
				Status: "Unauthorized",
				Body: io.NopCloser(strings.NewReader(`{
					"error": {
						"code": 123,
						"message": "message",
						"target": "target",
						"correlationID": "correlationID"
					}
				}`)),
			},
			expectedErr: errors.New("message"),
		},
		{
			name: "decode response error",
			response: &http.Response{
				Status: "Unauthorized",
				Body:   io.NopCloser(strings.NewReader("[test=value]")),
			},
			expectedErr: errors.New("failed to decode error response with status 'Unauthorized': invalid character 'e' in literal true (expecting 'r')"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &httpClient{}

			err := c.buildResponseError(tt.response)

			require.Equal(t, tt.expectedErr, err)
		})
	}
}

func fixRequestOptions() requestOptions {
	return requestOptions{
		Body: strings.NewReader("test data"),
		Headers: map[string]string{
			"Test-Header": "test-header-value",
		},
		Query: map[string]string{
			"test-query": "test-query-value",
		},
	}
}

func fixGenericRequestHandler(t *testing.T, expectedOptions requestOptions) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		for key, expectedValue := range expectedOptions.Query {
			value, ok := r.URL.Query()[key]
			require.True(t, ok)
			require.Equal(t, expectedValue, value[0])
		}

		for key, expectedValue := range expectedOptions.Headers {
			value, ok := r.Header[key]
			require.True(t, ok)
			require.Equal(t, expectedValue, value[0])
		}

		data := make([]byte, 0)
		expectedData := make([]byte, 0)
		var err error
		if r.Body != nil {
			data, err = io.ReadAll(r.Body)
			require.NoError(t, err)
		}
		if expectedOptions.Body != nil {
			expectedData, err = io.ReadAll(expectedOptions.Body)
			require.NoError(t, err)
		}
		require.Equal(t, expectedData, data)

		w.WriteHeader(200)
	}
}
