package communitymodules

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_modulesCatalog(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		response := `
			[
			  {
				"name": "module1",
				"versions": [
				  {
					"version": "1.2.3",
					"repository": "https://repo/path/module1.git",
					"managerPath": "/some/path/module1-controller-manager"
				  },
				  {
					"version": "1.7.0",
					"repository": "https://other/repo/path/module1.git",
					"managerPath": "/other/path/module1-controller-manager"
				  }
				]
			  },
			  {
				"name": "module2",
				"versions": [
				  {
					"version": "4.5.6",
					"repository": "https://repo/path/module2.git",
					"managerPath": "/some/path/module2-manager"
				  }
				]
			  }
			]`
		expectedResult := moduleMap{
			"module1": row{
				Name:       "module1",
				Repository: "https://repo/path/module1.git",
				Version:    "",
				Managed:    "",
			},
			"module2": row{
				Name:       "module2",
				Repository: "https://repo/path/module2.git",
				Version:    "",
				Managed:    "",
			},
		}

		httpServer := httptest.NewServer(http.HandlerFunc(fixHttpResponseHandler(200, response)))
		defer httpServer.Close()
		modules, err := modulesCatalog(httpServer.URL)
		require.Nil(t, err)
		require.Equal(t, expectedResult, modules)
	})

	t.Run("invalid http response", func(t *testing.T) {
		response := ""

		httpServer := httptest.NewServer(http.HandlerFunc(fixHttpResponseHandler(500, response)))
		defer httpServer.Close()
		modules, err := modulesCatalog(httpServer.URL)
		require.Nil(t, modules)
		require.NotNil(t, err)
		require.Contains(t, err.String(), "while handling response")
		require.Contains(t, err.String(), "error response: 500")
	})

	t.Run("invalid json response", func(t *testing.T) {
		response := "invalid json"

		httpServer := httptest.NewServer(http.HandlerFunc(fixHttpResponseHandler(200, response)))
		defer httpServer.Close()
		modules, err := modulesCatalog(httpServer.URL)
		require.Nil(t, modules)
		require.NotNil(t, err)
		require.Contains(t, err.String(), "while handling response")
		require.Contains(t, err.String(), "while unmarshalling")
	})
}

func fixHttpResponseHandler(status int, response string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(response))
	}
}
