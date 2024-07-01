package communitymodules

import (
	"context"
	"encoding/json"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	kube_fake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamic_fake "k8s.io/client-go/dynamic/fake"
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
				Name:          "module1",
				Repository:    "https://repo/path/module1.git",
				LatestVersion: "1.2.3",
				Version:       "",
				Managed:       "",
			},
			"module2": row{
				Name:          "module2",
				Repository:    "https://repo/path/module2.git",
				LatestVersion: "4.5.6",
				Version:       "",
				Managed:       "",
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

func Test_ManagedModules(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		expectedResult := moduleMap{
			"module1": row{
				Name:    "module1",
				Managed: "True",
			},
			"module2": row{
				Name:    "module2",
				Managed: "True",
			},
			"module3": row{
				Name:    "module3",
				Managed: "True",
			},
		}

		testKyma := fixTestKyma()

		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRKyma.GroupVersion(), testKyma)
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme, testKyma)

		kubeClient := &kube_fake.FakeKubeClient{
			TestKubernetesInterface: nil,
			TestDynamicInterface:    dynamic,
		}

		kymaConfig := cmdcommon.KymaConfig{
			Ctx: context.Background(),
		}

		modules, err := ManagedModules(cmdcommon.KubeClientConfig{
			Kubeconfig: "",
			KubeClient: kubeClient,
		}, kymaConfig)

		assert.Equal(t, expectedResult, modules)
		assert.Nil(t, err)
	})
	t.Run("kyma cr not found", func(t *testing.T) {
		expectedResult := moduleMap{}

		testKyma := fixTestKyma()

		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRKyma.GroupVersion(), testKyma)
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme)

		kubeClient := &kube_fake.FakeKubeClient{
			TestKubernetesInterface: nil,
			TestDynamicInterface:    dynamic,
		}

		kymaConfig := cmdcommon.KymaConfig{
			Ctx: context.Background(),
		}

		modules, err := ManagedModules(cmdcommon.KubeClientConfig{
			Kubeconfig: "",
			KubeClient: kubeClient,
		}, kymaConfig)

		assert.Equal(t, expectedResult, modules)
		assert.Nil(t, err)
	})
}

func Test_installedModules(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
	})
}

func fixHttpResponseHandler(status int, response string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(response))
	}
}

func fixTestKyma() *unstructured.Unstructured {
	b := []byte(`
		{
		  "apiVersion": "operator.kyma-project.io/v1beta2",
		  "kind": "Kyma",
		  "metadata": {
			"managedFields": [
			  {
				"fieldsV1": {
				  "f:spec": {
					"f:modules": {
					  ".": {},
					  "k:{\"name\":\"module1\"}": {
						".": {},
						"f:customResourcePolicy": {},
						"f:name": {}
					  },
					  "k:{\"name\":\"module3\"}": {
						".": {},
						"f:customResourcePolicy": {},
						"f:name": {}
					  },
					  "k:{\"name\":\"module2\"}": {
						".": {},
						"f:customResourcePolicy": {},
						"f:name": {}
					  }
					}
				  }
				}
			  }
			],
			"name": "default",
			"namespace": "kyma-system"
		  }
		}
		`)
	f := make(map[string]interface{})
	_ = json.Unmarshal(b, &f)
	u := &unstructured.Unstructured{Object: f}
	return u
}
