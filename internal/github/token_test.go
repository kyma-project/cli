package github

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/stretchr/testify/require"
)

func Test_getGithubToken(t *testing.T) {
	t.Parallel()

	svrGood := httptest.NewServer(http.HandlerFunc(fixGithubHandler(t)))
	defer svrGood.Close()
	svrGoodAudience := httptest.NewServer(http.HandlerFunc(fixGithubAudienceHandler(t, "audience")))
	defer svrGoodAudience.Close()
	svrBad := httptest.NewServer(http.HandlerFunc(fixGithubErrorHandler(t)))
	defer svrBad.Close()

	tests := []struct {
		name         string
		url          string
		requestToken string
		audience     string
		want         string
		wantedErr    clierror.Error
	}{
		{
			name:      "wrong URL",
			url:       "doesnotexist",
			wantedErr: clierror.Wrap(errors.New("Get \"doesnotexist\": unsupported protocol scheme \"\""), clierror.New("failed to get token from GitHub")),
		},
		{
			name: "token request",
			url:  svrGood.URL,
			want: "token",
		},
		{
			name:     "token request with audience",
			url:      svrGoodAudience.URL,
			audience: "audience",
			want:     "token",
		},
		{
			name:      "token request error",
			url:       svrBad.URL,
			wantedErr: clierror.New("Invalid server response: 500"),
		},
	}
	for _, tt := range tests {
		url := tt.url
		requestToken := tt.requestToken
		audience := tt.audience
		want := tt.want
		wantedErr := tt.wantedErr
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetToken(url, requestToken, audience)
			require.Equal(t, wantedErr, err)
			if got != want {
				t.Errorf("GetToken() = %v, want %v", got, want)
			}
		})
	}
}

func fixGithubHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{"value":"token"}`))
		require.NoError(t, err)
	}
}

func fixGithubAudienceHandler(t *testing.T, audience string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, audience, r.URL.Query().Get("audience"))
		_, err := w.Write([]byte(`{"value":"token"}`))
		require.NoError(t, err)
	}
}

func fixGithubErrorHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}
}
