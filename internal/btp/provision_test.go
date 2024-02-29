package btp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func fixProvisionHandler(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != fmt.Sprintf("/%s", provisionEndpoint) {
			w.WriteHeader(404)
			return
		}
		request := ProvisionEnvironment{}
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		resp := ProvisionResponse{
			Name: request.Name,
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		w.WriteHeader(202)
		_, err = w.Write(data)
		require.NoError(t, err)
	}
}

func fixProvisionErrorHandler(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		data := cisErrorResponse{
			Error: cisError{
				Message: "error",
			},
		}
		response, err := json.Marshal(data)
		require.NoError(t, err)

		w.WriteHeader(401)
		_, err = w.Write(response)
		require.NoError(t, err)
	}
}

func TestCISClient_Provision(t *testing.T) {
	t.Parallel()

	svrGood := httptest.NewServer(http.HandlerFunc(fixProvisionHandler(t)))
	defer svrGood.Close()
	svrBad := httptest.NewServer(http.HandlerFunc(fixProvisionErrorHandler(t)))
	defer svrBad.Close()

	tests := []struct {
		name           string
		credentials    *CISCredentials
		token          *XSUAAToken
		pe             *ProvisionEnvironment
		wantedResponse *ProvisionResponse
		expectedErr    error
	}{
		{
			name: "Correct data",
			credentials: &CISCredentials{
				Endpoints: Endpoints{
					ProvisioningServiceURL: svrGood.URL,
				},
			},
			token: &XSUAAToken{},
			pe: &ProvisionEnvironment{
				Name: "name",
			},
			wantedResponse: &ProvisionResponse{
				Name: "name",
			},
			expectedErr: nil,
		},
		{
			name: "Incorrect URL",
			credentials: &CISCredentials{
				Endpoints: Endpoints{
					ProvisioningServiceURL: "?\n?",
				},
			},
			token: &XSUAAToken{},
			pe: &ProvisionEnvironment{
				Name: "name",
			},
			wantedResponse: nil,
			expectedErr:    errors.New("failed to provision: failed to build request: parse \"?\\n?/provisioning/v1/environments\": net/url: invalid control character in URL"),
		},
		{
			name: "Wrong URL",
			credentials: &CISCredentials{
				Endpoints: Endpoints{
					ProvisioningServiceURL: "http://doesnotexist",
				},
			},
			token: &XSUAAToken{},
			pe: &ProvisionEnvironment{
				Name: "name",
			},
			wantedResponse: nil,
			expectedErr:    errors.New("failed to provision: failed to get data from server: Post \"http://doesnotexist/provisioning/v1/environments\": dial tcp: lookup doesnotexist: no such host"),
		},
		{
			name: "Error response",
			credentials: &CISCredentials{
				Endpoints: Endpoints{
					ProvisioningServiceURL: svrBad.URL,
				},
			},
			token: &XSUAAToken{},
			pe: &ProvisionEnvironment{
				Name: "name",
			},
			wantedResponse: nil,
			expectedErr:    errors.New("failed to provision: error"),
		},
	}
	for _, tt := range tests {
		pe := tt.pe
		credentials := tt.credentials
		token := tt.token
		wantedResponse := tt.wantedResponse
		expectedErr := tt.expectedErr
		t.Run(tt.name, func(t *testing.T) {
			c := NewLocalClient(credentials, token)

			response, err := c.Provision(pe)
			require.Equal(t, expectedErr, err)
			require.Equal(t, wantedResponse, response)
		})
	}
}
