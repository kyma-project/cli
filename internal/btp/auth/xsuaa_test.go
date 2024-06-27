package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/stretchr/testify/require"
)

const (
	correctClientID     = "user"
	correctClientSecret = "zaq12wsx"
	correctToken        = "kusedug"
	supportedGrantType  = "client_credentials"
)

func TestGetOAuthToken(t *testing.T) {
	t.Parallel()

	svrGood := httptest.NewServer(http.HandlerFunc(fixAuthenticationHandler(t)))
	defer svrGood.Close()
	svrBad := httptest.NewServer(http.HandlerFunc(fixAuthenticationErrorHandler(t)))
	defer svrBad.Close()

	tests := []struct {
		name        string
		credentials *CISCredentials
		want        *XSUAAToken
		expectedErr clierror.Error
	}{
		{
			name: "Correct credentials",
			credentials: &CISCredentials{
				UAA: UAA{
					ClientID:     correctClientID,
					ClientSecret: correctClientSecret,
					URL:          svrGood.URL,
				},
			},
			want: &XSUAAToken{
				AccessToken: correctToken,
			},
			expectedErr: nil,
		},
		{
			name: "Incorrect URL",
			credentials: &CISCredentials{
				UAA: UAA{
					URL: "?\n?",
				},
			},
			want: nil,
			expectedErr: clierror.Wrap(
				errors.New("parse \"?\\n?/oauth/token\": net/url: invalid control character in URL"),
				clierror.New("failed to build request", "Make sure the server URL in the config is correct."),
			),
		},
		{
			name: "Wrong URL",
			credentials: &CISCredentials{
				UAA: UAA{
					URL: "doesnotexist",
				},
			},
			want: nil,
			expectedErr: clierror.Wrap(
				errors.New("Post \"doesnotexist/oauth/token\": unsupported protocol scheme \"\""),
				clierror.New("failed to get token from server"),
			),
		},
		{
			name: "Error response",
			credentials: &CISCredentials{
				UAA: UAA{
					URL: svrBad.URL,
				},
			},
			want: nil,
			expectedErr: clierror.Wrap(
				errors.New("description"),
				clierror.New("error response: 401 Unauthorized"),
			),
		},
	}
	for _, tt := range tests {
		credentials := tt.credentials
		want := tt.want
		expectedErr := tt.expectedErr

		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOAuthToken(
				credentials.GrantType,
				credentials.UAA.URL,
				credentials.UAA.ClientID,
				credentials.UAA.ClientSecret,
			)
			require.Equal(t, expectedErr, err)
			require.Equal(t, want, got)
		})
	}
}

func fixAuthenticationHandler(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != fmt.Sprintf("/%s", authorizationEndpoint) {
			w.WriteHeader(404)
			return
		}
		username, password, ok := r.BasicAuth()
		require.True(t, ok)
		require.Equal(t, correctClientID, username)
		require.Equal(t, correctClientSecret, password)

		data := XSUAAToken{
			AccessToken: correctToken,
		}
		response, err := json.Marshal(data)
		require.NoError(t, err)

		_, err = w.Write(response)
		require.NoError(t, err)
	}
}

func fixAuthenticationErrorHandler(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		data := xsuaaErrorResponse{
			Error:            "error",
			ErrorDescription: "description",
		}
		response, err := json.Marshal(data)
		require.NoError(t, err)

		w.WriteHeader(401)
		_, err = w.Write(response)
		require.NoError(t, err)
	}
}
