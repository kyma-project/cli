package hana

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetID(t *testing.T) {
	authToken := "test-token"
	t.Run("get hana instance ID from server", func(t *testing.T) {
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/inventory/v2/serviceInstances", r.URL.Path)
			require.Equal(t, fmt.Sprintf("Bearer %s", authToken), r.Header.Get("Authorization"))
			require.Equal(t, http.MethodGet, r.Method)

			_, err := w.Write([]byte(`{"data":[{"id":"test-id"}]}`))
			require.NoError(t, err)
			w.WriteHeader(http.StatusOK)
		}))
		defer testServer.Close()

		id, err := getID(testServer.URL, authToken)
		require.Nil(t, err)
		require.Equal(t, "test-id", id)
	})
	t.Run("response body empty", func(t *testing.T) {
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/inventory/v2/serviceInstances", r.URL.Path)
			require.Equal(t, fmt.Sprintf("Bearer %s", authToken), r.Header.Get("Authorization"))
			require.Equal(t, http.MethodGet, r.Method)
			_, err := w.Write([]byte(`{"data":[]}`))
			require.NoError(t, err)
			w.WriteHeader(http.StatusOK)
		}))
		defer testServer.Close()

		id, err := getID(testServer.URL, authToken)
		require.Contains(t, err.String(), "no SAP Hana instances found in the response")
		require.Empty(t, id)
	})
	t.Run("can't create new GET request", func(t *testing.T) {
		id, err := getID(":		", authToken)
		require.Contains(t, err.String(), "failed to create a get SAP Hana instances request")
		require.Empty(t, id)
	})
	t.Run("can't send GET request", func(t *testing.T) {
		id, err := getID("https://localhost", authToken)
		require.Contains(t, err.String(), "failed to get SAP Hana instances")
		require.Empty(t, id)
	})
	t.Run("response status not OK", func(t *testing.T) {
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer testServer.Close()

		id, err := getID(testServer.URL, authToken)
		require.Contains(t, err.String(), "unexpected status code: 500")
		require.Empty(t, id)
	})
	t.Run("failed to unmarshal response", func(t *testing.T) {
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte(`wrong format	{`))
			require.NoError(t, err)
			w.WriteHeader(http.StatusOK)
		}))
		defer testServer.Close()

		id, err := getID(testServer.URL, authToken)
		require.Contains(t, err.String(), "failed to parse SAP Hana instances from the response")
		require.Empty(t, id)
	})
}
