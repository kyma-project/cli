package hana

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapInstance(t *testing.T) {
	t.Run("map hana instance", func(t *testing.T) {
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/inventory/v2/serviceInstances/test-id/instanceMappings", r.URL.Path)
			require.Equal(t, fmt.Sprintf("Bearer %s", "test-token"), r.Header.Get("Authorization"))
			require.Equal(t, http.MethodPost, r.Method)

			w.WriteHeader(http.StatusOK)
		}))
		defer testServer.Close()

		err := mapInstance(testServer.URL, "test-cluster-id", "test-id", "test-token")
		require.Nil(t, err)
	})
	t.Run("failed to create mapping request", func(t *testing.T) {
		err := mapInstance(":		", "test-cluster-id", "test-id", "test-token")
		require.Contains(t, err.String(), "failed to create mapping request")
	})
	t.Run("failed to post mapping request", func(t *testing.T) {

		err := mapInstance("https://localhost", "test-cluster-id", "test-id", "test-token")
		require.Contains(t, err.String(), "failed to post mapping request")
	})
	t.Run("unexpected status code", func(t *testing.T) {
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer testServer.Close()

		err := mapInstance(testServer.URL, "test-cluster-id", "test-id", "test-token")
		require.Contains(t, err.String(), "unexpected status code")
	})
}
