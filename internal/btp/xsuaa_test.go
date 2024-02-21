package btp

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
		expectedErr error
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
			want:        nil,
			expectedErr: errors.New("failed to build request: parse \"?\\n?/oauth/token\": net/url: invalid control character in URL"),
		},
		{
			name: "Wrong URL",
			credentials: &CISCredentials{
				UAA: UAA{
					URL: "http://doesnotexist",
				},
			},
			want:        nil,
			expectedErr: errors.New("failed to get token from server: Post \"http://doesnotexist/oauth/token\": dial tcp: lookup doesnotexist: no such host"),
		},
		{
			name: "Error response",
			credentials: &CISCredentials{
				UAA: UAA{
					URL: svrBad.URL,
				},
			},
			want:        nil,
			expectedErr: errors.New("error response: error: description"),
		},
	}
	for _, tt := range tests {
		credentials := tt.credentials
		want := tt.want
		expectedErr := tt.expectedErr

		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOAuthToken(credentials)
			require.Equal(t, expectedErr, err)
			require.Equal(t, want, got)
		})
	}
}

func fixAuthenticationHandler(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token" {
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
		data := errorResponse{
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
