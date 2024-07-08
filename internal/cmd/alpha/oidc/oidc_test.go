package oidc

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-test/deep"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd/api"
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
			wantedErr: clierror.Wrap(errors.New("Get \"doesnotexist\": unsupported protocol scheme \"\""), clierror.New("failed to get token from Github")),
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
			got, err := getGithubToken(url, requestToken, audience)
			require.Equal(t, wantedErr, err)
			if got != want {
				t.Errorf("getGithubToken() = %v, want %v", got, want)
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

func Test_createKubeconfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		kubeconfig *api.Config
		token      string
		want       *api.Config
	}{
		{
			name: "change kubeconfig authInfo",
			kubeconfig: &api.Config{
				Clusters: map[string]*api.Cluster{
					"cluster": {},
				},
				AuthInfos: map[string]*api.AuthInfo{
					"user": {
						Username:  "user",
						ClientKey: "remove",
					},
				},
				Contexts: map[string]*api.Context{
					"context": {
						AuthInfo: "user",
					},
				},
				CurrentContext: "context",
			},
			token: "token",
			want: &api.Config{
				Kind:       "Config",
				APIVersion: "v1",
				Clusters: map[string]*api.Cluster{
					"cluster": {},
				},
				AuthInfos: map[string]*api.AuthInfo{
					"user": {
						Token: "token",
					},
				},
				Contexts: map[string]*api.Context{
					"context": {
						AuthInfo: "user",
					},
				},
				CurrentContext: "context",
			},
		},
	}

	for _, tt := range tests {
		kubeconfig := tt.kubeconfig
		token := tt.token
		want := tt.want
		t.Run(tt.name, func(t *testing.T) {
			got := createKubeconfig(kubeconfig, token)
			if diff := deep.Equal(got, want); diff != nil {
				t.Errorf("createKubeconfig() = %s", diff)
			}
		})
	}
}
