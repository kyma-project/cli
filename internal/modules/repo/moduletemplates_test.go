package repo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	modulesfake "github.com/kyma-project/cli.v3/internal/modules/fake"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
)

var (
	testCoreModuleTemplate = kyma.ModuleTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.kyma-project.io/v1beta2",
			Kind:       "ModuleTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-module-1",
			Namespace: "test-module-namespace",
			Labels: map[string]string{
				"operator.kyma-project.io/managed-by": "kyma",
			},
		},
		Spec: kyma.ModuleTemplateSpec{
			ModuleName: "test-module",
			Version:    "0.0.1",
		},
	}
	testCommunityModuleTemplate = kyma.ModuleTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.kyma-project.io/v1beta2",
			Kind:       "ModuleTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-module-2",
			Namespace: "test-module-namespace",
		},
		Spec: kyma.ModuleTemplateSpec{
			ModuleName: "test-module",
			Version:    "0.0.2",
		},
	}
	runningManagerMock = unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]any{
				"namespace": "community-module-system",
				"labels": map[string]any{
					"app.kubernetes.io/version": "0.0.1",
				},
				"name": "community-module-controller-manager3",
			},
			"spec": map[string]any{
				"template": map[string]any{
					"spec": map[string]any{
						"containers": []any{
							map[string]any{
								"name":  "proxy",
								"image": "http://repo.url/proxy:2.0.0",
							},
							map[string]any{
								"name":  "manager",
								"image": "http://repo.url/manager:0.0.1",
							},
						},
					},
				},
			},
		},
	}
)

var (
	resourcesNamespace = `apiVersion: v1
kind: Namespace
metadata:
  name: test-module-system
`
	resourcesCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: communitymodules.operator.kyma-project.io
spec:
  group: operator.kyma-project.io
  names:
  kind: CommunityModule
  listKind: CommunityModules
  plural: communitymodules
  singular: communitymodule
  scope: Namespaced
  versions:
  - name: v1alpha1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          testField:
            type: string`

	resourcesDeployment = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-module-controller-manager
  namespace: test-module-system
  labels: 
    app.kubernetes.io/version: 0.0.1
spec:
  template:
    spec:
      containers:
      - command:
        - /manager
      image: http://repo.url/manager:0.0.1
      name: manager`
)

func getTestInstalledCommunityModuleTemplate(link string) kyma.ModuleTemplate {
	return kyma.ModuleTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.kyma-project.io/v1beta2",
			Kind:       "ModuleTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "installed-test-module",
			Namespace: "test-module-namespace",
		},
		Spec: kyma.ModuleTemplateSpec{
			ModuleName:          "test-module",
			Version:             "0.0.3",
			AssociatedResources: nil,
			Data: unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "operator.kyma-project.io/v1beta2",
					"kind":       "CommunityModule",
					"metadata": map[string]any{
						"name": "community-module-test",
					},
				},
			},
			Manager: &kyma.Manager{
				Name: "test-module-controller-manager",
				GroupVersionKind: metav1.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
			},
			Resources: []kyma.Resource{
				{
					Name: "rawManifest",
					Link: link,
				},
			},
		},
	}
}

func TestModuleTemplatesRepo_local(t *testing.T) {
	t.Run("failed to list module templates", func(t *testing.T) {
		fakeKymaClient := fake.KymaClient{
			ReturnErr: errors.New("test-error"),
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{},
			},
		}
		fakeKubeClient := fake.KubeClient{
			TestKymaInterface: &fakeKymaClient,
		}
		repo := NewModuleTemplatesRepo(&fakeKubeClient)

		result, err := repo.local(context.Background())

		require.Len(t, result, 0)
		require.Error(t, err)
		require.Equal(t, err.Error(), "failed to list module templates: test-error")
	})

	t.Run("lists all module templates", func(t *testing.T) {
		fakeKymaClient := fake.KymaClient{
			ReturnErr: nil,
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{
					testCoreModuleTemplate,
					testCommunityModuleTemplate,
				},
			},
		}

		fakeKubeClient := fake.KubeClient{
			TestKymaInterface: &fakeKymaClient,
		}

		repo := NewModuleTemplatesRepo(&fakeKubeClient)

		result, err := repo.local(context.Background())
		require.NoError(t, err)
		require.Len(t, result, 2)
		require.Equal(t, "test-module-1", result[0].ObjectMeta.Name)
		require.Equal(t, "test-module-2", result[1].ObjectMeta.Name)
	})
}

func TestModuleTemplatesRepo_Community(t *testing.T) {
	t.Run("failed to list module templates", func(t *testing.T) {
		fakeKubeClient := &fake.KubeClient{}
		fakeRemoteRepo := &modulesfake.ModuleTemplatesRemoteRepo{
			ReturnCommunity: nil,
			CommunityErr:    errors.New("test-error"),
		}

		repo := NewModuleTemplatesRepoForTests(fakeKubeClient, fakeRemoteRepo)

		result, err := repo.Community(context.Background())

		require.Len(t, result, 0)
		require.Error(t, err)
		require.Equal(t, "test-error", err.Error())
	})

	t.Run("lists community module templates from remote", func(t *testing.T) {
		fakeKubeClient := &fake.KubeClient{}
		fakeRemoteRepo := &modulesfake.ModuleTemplatesRemoteRepo{
			ReturnCommunity: []kyma.ModuleTemplate{
				testCommunityModuleTemplate,
			},
		}

		repo := NewModuleTemplatesRepoForTests(fakeKubeClient, fakeRemoteRepo)

		result, err := repo.Community(context.Background())

		require.NoError(t, err)
		require.Len(t, result, 1)
		require.Equal(t, "test-module", result[0].Spec.ModuleName)
		require.Empty(t, result[0].ObjectMeta.Labels["operator.kyma-project.io/managed-by"])
	})
}

func TestModuleTemplatesRepo_CommunityByName(t *testing.T) {
	t.Run("failed to list module templates", func(t *testing.T) {
		fakeKubeClient := &fake.KubeClient{}
		fakeRemoteRepo := &modulesfake.ModuleTemplatesRemoteRepo{
			ReturnCommunity: nil,
			CommunityErr:    errors.New("test-error"),
		}

		repo := NewModuleTemplatesRepoForTests(fakeKubeClient, fakeRemoteRepo)

		result, err := repo.CommunityByName(context.Background(), "test")

		require.Len(t, result, 0)
		require.Error(t, err)
		require.Equal(t, "test-error", err.Error())
	})

	t.Run("returns only community modules with specific name", func(t *testing.T) {
		fakeKubeClient := &fake.KubeClient{}
		fakeRemoteRepo := &modulesfake.ModuleTemplatesRemoteRepo{
			ReturnCommunity: []kyma.ModuleTemplate{
				testCommunityModuleTemplate,
			},
		}
		repo := NewModuleTemplatesRepoForTests(fakeKubeClient, fakeRemoteRepo)

		result, err := repo.CommunityByName(context.Background(), "test-module")

		require.NoError(t, err)
		require.Len(t, result, 1)
		require.Equal(t, "test-module", result[0].Spec.ModuleName)
		require.Empty(t, result[0].ObjectMeta.Labels["operator.kyma-project.io/managed-by"])
	})

	t.Run("does not return community modules with different names", func(t *testing.T) {
		fakeKubeClient := &fake.KubeClient{}
		fakeRemoteRepo := &modulesfake.ModuleTemplatesRemoteRepo{
			ReturnCommunity: []kyma.ModuleTemplate{
				testCommunityModuleTemplate,
			},
		}
		repo := NewModuleTemplatesRepoForTests(fakeKubeClient, fakeRemoteRepo)

		result, err := repo.CommunityByName(context.Background(), "non-existing-name")

		require.NoError(t, err)
		require.Len(t, result, 0)
	})
}

func TestModuleTemplatesRepo_CommunityInstalledByName(t *testing.T) {
	t.Run("failed to list module templates", func(t *testing.T) {
		fakeKubeClient := &fake.KubeClient{}
		fakeRemoteRepo := &modulesfake.ModuleTemplatesRemoteRepo{
			ReturnCommunity: nil,
			CommunityErr:    errors.New("test-error"),
		}

		repo := NewModuleTemplatesRepoForTests(fakeKubeClient, fakeRemoteRepo)

		result, err := repo.CommunityInstalledByName(context.Background(), "test-module")

		require.Len(t, result, 0)
		require.Error(t, err)
		require.Equal(t, "test-error", err.Error())
	})

	t.Run("lists only installed community modules", func(t *testing.T) {
		moduleResourcesServer := getTestHttpServerWithResponse(
			http.StatusOK,
			getResourcesUrlResponseYaml(resourcesNamespace, resourcesCRD, resourcesDeployment),
		)
		defer moduleResourcesServer.Close()

		testInstalledCommunityModuleTemplate := getTestInstalledCommunityModuleTemplate(moduleResourcesServer.URL)

		fakeKymaClient := fake.KymaClient{
			ReturnErr: nil,
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{
					testCoreModuleTemplate,
				},
			},
		}

		fakeRootless := &fake.RootlessDynamicClient{
			ReturnGetObj: runningManagerMock,
		}

		fakeKubeClient := &fake.KubeClient{
			TestKymaInterface:            &fakeKymaClient,
			TestRootlessDynamicInterface: fakeRootless,
		}

		fakeRemoteRepo := &modulesfake.ModuleTemplatesRemoteRepo{
			ReturnCommunity: []kyma.ModuleTemplate{
				testCommunityModuleTemplate,
				testInstalledCommunityModuleTemplate,
			},
		}

		repo := NewModuleTemplatesRepoForTests(fakeKubeClient, fakeRemoteRepo)

		result, err := repo.CommunityInstalledByName(context.Background(), "test-module")

		require.Len(t, result, 1)
		require.NoError(t, err)
		require.Equal(t, "installed-test-module", result[0].Name)
	})
}

func TestModuleTemplatesRepo_RunningAssociatedResourcesOfModule(t *testing.T) {
	t.Run("fails to list running resources", func(t *testing.T) {
		fakeRootlessDynamicClient := fake.RootlessDynamicClient{
			ReturnErr: errors.New("list-error"),
		}
		fakeKubeClient := fake.KubeClient{
			TestRootlessDynamicInterface: &fakeRootlessDynamicClient,
		}
		repo := NewModuleTemplatesRepo(&fakeKubeClient)
		mod := kyma.ModuleTemplate{
			Spec: kyma.ModuleTemplateSpec{
				AssociatedResources: []metav1.GroupVersionKind{{Group: "g", Version: "v1", Kind: "Kind"}},
			},
		}

		result, err := repo.RunningAssociatedResourcesOfModule(context.Background(), mod)

		require.NoError(t, err)
		require.Len(t, result, 0)
	})

	t.Run("excludes spec.data resource", func(t *testing.T) {
		fakeRootlessDynamicClient := fake.RootlessDynamicClient{
			ReturnListObjs: &unstructured.UnstructuredList{
				Items: []unstructured.Unstructured{
					{
						Object: map[string]any{
							"metadata": map[string]any{
								"name": "res1",
							},
						},
					},
				},
			},
		}

		fakeKubeClient := fake.KubeClient{
			TestRootlessDynamicInterface: &fakeRootlessDynamicClient,
		}
		repo := NewModuleTemplatesRepo(&fakeKubeClient)

		mod := kyma.ModuleTemplate{
			Spec: kyma.ModuleTemplateSpec{
				AssociatedResources: []metav1.GroupVersionKind{
					{Group: "g", Version: "v1", Kind: "Kind"},
					{Group: "g", Version: "v1", Kind: "Operator"},
				},
				Data: unstructured.Unstructured{
					Object: map[string]any{
						"apiVersion": "g/v1",
						"kind":       "Operator",
					},
				},
			},
		}

		resources, err := repo.RunningAssociatedResourcesOfModule(context.Background(), mod)

		require.NoError(t, err)
		require.Len(t, resources, 1)
		require.Equal(t, "res1", resources[0].GetName())
	})

	t.Run("running resources found", func(t *testing.T) {
		fakeRootlessDynamicClient := fake.RootlessDynamicClient{
			ReturnListObjs: &unstructured.UnstructuredList{
				Items: []unstructured.Unstructured{
					{
						Object: map[string]any{
							"metadata": map[string]any{
								"name": "res1",
							},
						},
					},
				},
			},
		}

		fakeKubeClient := fake.KubeClient{
			TestRootlessDynamicInterface: &fakeRootlessDynamicClient,
		}
		repo := NewModuleTemplatesRepo(&fakeKubeClient)

		mod := kyma.ModuleTemplate{
			Spec: kyma.ModuleTemplateSpec{
				AssociatedResources: []metav1.GroupVersionKind{{Group: "g", Version: "v1", Kind: "Kind"}},
			},
		}

		resources, err := repo.RunningAssociatedResourcesOfModule(context.Background(), mod)

		require.NoError(t, err)
		require.Len(t, resources, 1)
		require.Equal(t, "res1", resources[0].GetName())
	})
}

func TestModuleTemplatesRepo_DeleteResourceReturnWatcher(t *testing.T) {
	t.Run("fails to watch resource", func(t *testing.T) {
		fakeRootlessDynamicClient := fake.RootlessDynamicClient{
			ReturnWatchErr: errors.New("watch error"),
		}
		fakeKubeClient := fake.KubeClient{
			TestRootlessDynamicInterface: &fakeRootlessDynamicClient,
		}
		resource := map[string]any{
			"metadata": map[string]any{
				"name": "test-resource",
			},
		}
		repo := NewModuleTemplatesRepo(&fakeKubeClient)

		result, err := repo.DeleteResourceReturnWatcher(context.Background(), resource)

		require.Nil(t, result)
		require.NotNil(t, err)
		require.Error(t, err, "failed to watch resource test-resource: watch error")
	})

	t.Run("fails to delete resource", func(t *testing.T) {
		fakeWatcher := watch.NewFake()
		fakeRootlessDynamicClient := fake.RootlessDynamicClient{
			ReturnWatcher:   fakeWatcher,
			ReturnRemoveErr: errors.New("remove error"),
		}
		fakeKubeClient := fake.KubeClient{
			TestRootlessDynamicInterface: &fakeRootlessDynamicClient,
		}
		resource := map[string]any{
			"metadata": map[string]any{
				"name": "test-resoutce",
			},
		}
		repo := NewModuleTemplatesRepo(&fakeKubeClient)

		result, err := repo.DeleteResourceReturnWatcher(context.Background(), resource)

		require.Nil(t, result)
		require.NotNil(t, err)
		require.Error(t, err, "failed to remove resource test-resource: remove error")
	})

	t.Run("returns watcher", func(t *testing.T) {
		fakeWatcher := watch.NewFake()
		fakeRootlessDynamicClient := fake.RootlessDynamicClient{
			ReturnWatcher: fakeWatcher,
		}
		fakeKubeClient := fake.KubeClient{
			TestRootlessDynamicInterface: &fakeRootlessDynamicClient,
		}
		resource := map[string]any{
			"metadata": map[string]any{
				"name": "test-resource",
			},
		}
		repo := NewModuleTemplatesRepo(&fakeKubeClient)

		result, err := repo.DeleteResourceReturnWatcher(context.Background(), resource)

		require.Nil(t, err)
		require.NotNil(t, result)
		require.Equal(t, result, fakeWatcher)
		require.Contains(t, fakeRootlessDynamicClient.RemovedObjs, unstructured.Unstructured{Object: resource})
	})
}

func getTestHttpServerWithResponse(httpStatus int, response string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(response))
		w.WriteHeader(httpStatus)
	}))
}

func getResourcesUrlResponseYaml(parts ...string) string {
	return strings.Join(parts, "\n---\n")
}
