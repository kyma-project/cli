package cis

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/cli.v3/internal/btp/auth"
	"github.com/kyma-project/cli.v3/internal/clierror"
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
		credentials    *auth.CISCredentials
		token          *auth.XSUAAToken
		pe             *ProvisionEnvironment
		wantedResponse *ProvisionResponse
		expectedErr    clierror.Error
	}{
		{
			name: "Correct data",
			credentials: &auth.CISCredentials{
				Endpoints: auth.Endpoints{
					ProvisioningServiceURL: svrGood.URL,
				},
			},
			token: &auth.XSUAAToken{},
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
			credentials: &auth.CISCredentials{
				Endpoints: auth.Endpoints{
					ProvisioningServiceURL: "?\n?",
				},
			},
			token: &auth.XSUAAToken{},
			pe: &ProvisionEnvironment{
				Name: "name",
			},
			wantedResponse: nil,
			expectedErr:    clierror.New("failed to build request: parse \"?\\n?/provisioning/v1/environments\": net/url: invalid control character in URL"),
		},
		{
			name: "Wrong URL",
			credentials: &auth.CISCredentials{
				Endpoints: auth.Endpoints{
					ProvisioningServiceURL: "doesnotexist",
				},
			},
			token: &auth.XSUAAToken{},
			pe: &ProvisionEnvironment{
				Name: "name",
			},
			wantedResponse: nil,
			expectedErr:    clierror.New("failed to get data from server: Post \"doesnotexist/provisioning/v1/environments\": unsupported protocol scheme \"\""),
		},
		{
			name: "Error response",
			credentials: &auth.CISCredentials{
				Endpoints: auth.Endpoints{
					ProvisioningServiceURL: svrBad.URL,
				},
			},
			token: &auth.XSUAAToken{},
			pe: &ProvisionEnvironment{
				Name: "name",
			},
			wantedResponse: nil,
			expectedErr:    clierror.New("error"),
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
			if expectedErr != nil {
				require.Equal(t, expectedErr, err)
			}
			require.Equal(t, wantedResponse, response)
		})
	}
}
