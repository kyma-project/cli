package communitymodules

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	kube_fake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamic_fake "k8s.io/client-go/dynamic/fake"
	k8s_fake "k8s.io/client-go/kubernetes/fake"
)

func Test_modulesCatalog(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		expectedResult := moduleMap{
			"module1": row{
				Name:          "module1",
				Repository:    "https://repo2/path/module1.git",
				LatestVersion: "1.7.0",
				Version:       "",
				Channel:       "",
			},
			"module2": row{
				Name:          "module2",
				Repository:    "https://repo/path/module2.git",
				LatestVersion: "4.5.6",
				Version:       "",
				Channel:       "",
			},
		}

		httpServer := httptest.NewServer(http.HandlerFunc(
			fixHttpResponseHandler(200, fixCommunityModulesResponse())))
		defer httpServer.Close()
		modules, err := modulesCatalog(httpServer.URL)
		require.Nil(t, err)
		require.Equal(t, expectedResult, modules)
	})

	t.Run("invalid http response", func(t *testing.T) {
		httpServer := httptest.NewServer(http.HandlerFunc(fixHttpResponseHandler(500, "")))
		defer httpServer.Close()
		modules, err := modulesCatalog(httpServer.URL)
		require.Nil(t, modules)
		require.NotNil(t, err)
		require.Contains(t, err.String(), "while handling response")
		require.Contains(t, err.String(), "error response: 500")
	})

	t.Run("invalid json response", func(t *testing.T) {
		httpServer := httptest.NewServer(http.HandlerFunc(fixHttpResponseHandler(200, "invalid json")))
		defer httpServer.Close()
		modules, err := modulesCatalog(httpServer.URL)
		require.Nil(t, modules)
		require.NotNil(t, err)
		require.Contains(t, err.String(), "while handling response")
		require.Contains(t, err.String(), "while unmarshalling")
	})
	t.Run("greatest version is latest", func(t *testing.T) {
		expectedResult := moduleMap{
			"module1": row{
				Name:          "module1",
				Repository:    "https://repo2/path/module1.git",
				LatestVersion: "1.7.0",
				Version:       "",
				Channel:       "",
			},
			"module2": row{
				Name:          "module2",
				Repository:    "https://repo/path/module2.git",
				LatestVersion: "4.5.6",
				Version:       "",
				Channel:       "",
			},
		}

		httpServer := httptest.NewServer(http.HandlerFunc(fixHttpResponseHandler(200, fixCommunityModulesResponse())))
		defer httpServer.Close()
		modules, err := modulesCatalog(httpServer.URL)
		require.Nil(t, err)
		require.Equal(t, expectedResult, modules)
	})
}

func Test_ManagedModules(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		expectedResult := moduleMap{
			"module1": row{
				Name:    "module1",
				Channel: "fast",
			},
			"module2": row{
				Name:    "module2",
				Channel: "fast",
			},
			"module3": row{
				Name:    "module3",
				Channel: "regular",
			},
		}

		testKyma := fixTestKyma()
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRKyma.GroupVersion(), testKyma)
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme, testKyma)
		kubeClient := &kube_fake.FakeKubeClient{
			TestKubernetesInterface: nil,
			TestKymaInterface:       kyma.NewClient(dynamic),
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
			TestKymaInterface:       kyma.NewClient(dynamic),
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
		expectedResult := moduleMap{
			"module1": row{
				Name:    "module1",
				Version: "1.7.0",
			},
			"module2": row{
				Name:    "module2",
				Version: "outdated moduleVersion, latest is 4.5.6",
			},
		}

		httpServer := httptest.NewServer(http.HandlerFunc(
			fixHttpResponseHandler(200, fixCommunityModulesResponse())))
		defer httpServer.Close()

		staticClient := k8s_fake.NewSimpleClientset(
			fixTestDeployment("module1-controller-manager", "kyma-system", "1.7.0"),
			fixTestDeployment("module2-manager", "kyma-system", "6.7.8"), // outdated
			fixTestDeployment("other-deployment", "kyma-system", "1.2.3"))
		kubeClient := &kube_fake.FakeKubeClient{
			TestKubernetesInterface: staticClient,
			TestDynamicInterface:    nil,
		}

		kymaConfig := cmdcommon.KymaConfig{
			Ctx: context.Background(),
		}

		modules, err := installedModules(
			httpServer.URL,
			cmdcommon.KubeClientConfig{
				Kubeconfig: "",
				KubeClient: kubeClient,
			}, kymaConfig)

		assert.Equal(t, expectedResult, modules)
		assert.Nil(t, err)
	})
	t.Run("only one installed", func(t *testing.T) {
		expectedResult := moduleMap{
			"module2": row{
				Name:    "module2",
				Version: "4.5.6",
			},
		}

		httpServer := httptest.NewServer(http.HandlerFunc(
			fixHttpResponseHandler(200, fixCommunityModulesResponse())))
		defer httpServer.Close()

		staticClient := k8s_fake.NewSimpleClientset(
			fixTestDeployment("module2-manager", "kyma-system", "4.5.6"),
			fixTestDeployment("other-deployment", "kyma-system", "1.2.3"))
		kubeClient := &kube_fake.FakeKubeClient{
			TestKubernetesInterface: staticClient,
			TestDynamicInterface:    nil,
		}

		kymaConfig := cmdcommon.KymaConfig{
			Ctx: context.Background(),
		}

		modules, err := installedModules(
			httpServer.URL,
			cmdcommon.KubeClientConfig{
				Kubeconfig: "",
				KubeClient: kubeClient,
			}, kymaConfig)

		assert.Equal(t, expectedResult, modules)
		assert.Nil(t, err)
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
			"name": "default",
			"namespace": "kyma-system"
		  },
		  "status": {
			"modules": [
			  {
				"name": "module1",
				"channel": "fast"
			  },
			  {
				"name": "module3",
				"channel": "regular"
			  },
			  {
				"name": "module2",
				"channel": "fast"
			  }
			]
		  }
		}
		`)
	f := make(map[string]interface{})
	_ = json.Unmarshal(b, &f)
	u := &unstructured.Unstructured{Object: f}
	return u
}

func fixTestDeployment(name, namespace, imageTag string) *v1.Deployment {
	return &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: fmt.Sprintf("localhost:5000/some-project/some-image-name:%s", imageTag),
						},
					},
				},
			},
		},
	}
}

func fixCommunityModulesResponse() string {
	return `
	[
	  {
		"name": "module1",
		"versions": [
		  {
			"version": "1.5.3",
			"repository": "https://repo1/path/module1.git",
			"managerPath": "/some/path1/module1-controller-manager"
		  },
		  {
			"version": "1.7.0",
			"repository": "https://repo2/path/module1.git",
			"managerPath": "/other/path2/module1-controller-manager"
		  },
		  {
			"version": "1.3.4",
			"repository": "https://repo3/path/module1.git",
			"managerPath": "/some/path3/module1-controller-manager"
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
}

func Test_GetLatestVersion(t *testing.T) {
	t.Run("simple versions", func(t *testing.T) {
		result := GetLatestVersion([]Version{
			{
				Version: "1.5.3",
			},
			{
				Version: "1.7.1",
			},
			{
				Version: "1.4.3",
			},
		})
		assert.Equal(t, Version{
			Version: "1.7.1",
		}, result)
	})
	t.Run("v prefix", func(t *testing.T) {
		result := GetLatestVersion([]Version{
			{
				Version: "v1.5.3",
			},
			{
				Version: "v1.7.1",
			},
			{
				Version: "1.4.3",
			},
		})
		assert.Equal(t, Version{
			Version: "v1.7.1",
		}, result)
	})
	t.Run("with suffix", func(t *testing.T) {
		result := GetLatestVersion([]Version{
			{
				Version: "1.5.3-experimental",
			},
			{
				Version: "1.7.1-dev",
			},
			{
				Version: "1.7.1",
			},
			{
				Version: "1.4.3",
			},
		})
		assert.Equal(t, Version{
			Version: "1.7.1",
		}, result)
	})
}

func Test_chooseRepository(t *testing.T) {
	t.Run("version has repository", func(t *testing.T) {
		result := chooseRepository(Module{
			Name:       "module1",
			Repository: "https://repo1/path/module1.git",
		}, Version{
			Version:     "1.5.3",
			Repository:  "https://repo2/path/module1.git",
			ManagerPath: "/some/path1/module1-controller-manager",
		})
		assert.Equal(t, "https://repo2/path/module1.git", result)
	})
	t.Run("module has repository", func(t *testing.T) {
		result := chooseRepository(Module{
			Name:       "module1",
			Repository: "https://repo1/path/module1.git",
		}, Version{
			Version:     "1.5.3",
			Repository:  "",
			ManagerPath: "/some/path1/module1-controller-manager",
		})
		assert.Equal(t, "https://repo1/path/module1.git", result)
	})
	t.Run("Both have repository", func(t *testing.T) {
		result := chooseRepository(Module{
			Name:       "module1",
			Repository: "https://repo1/path/module1.git",
		}, Version{
			Version:     "1.5.3",
			Repository:  "https://repo2/path/module1.git",
			ManagerPath: "/some/path1/module1-controller-manager",
		})
		assert.Equal(t, "https://repo2/path/module1.git", result)
	})
	t.Run("no repository", func(t *testing.T) {
		result := chooseRepository(Module{
			Name:       "module1",
			Repository: "",
		}, Version{
			Version:     "1.5.3",
			Repository:  "",
			ManagerPath: "/some/path1/module1-controller-manager",
		})
		assert.Equal(t, "Unknown", result)
	})
}
