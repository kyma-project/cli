package modules

import (
	"context"
	"strings"
	"testing"

	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	modulesfake "github.com/kyma-project/cli.v3/internal/modules/fake"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func getTestCommunityModuleTemplateWithResourcesLink(link string) kyma.ModuleTemplate {
	return kyma.ModuleTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.kyma-project.io/v1beta2",
			Kind:       "ModuleTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "communitymodule-001",
			Namespace: "kyma-system",
		},
		Spec: kyma.ModuleTemplateSpec{
			ModuleName: "communitymodule",
			Version:    "0.0.1",
			Data: unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "operator.kyma-project.io/v1alpha1",
					"kind":       "CommunityModule",
					"metadata": map[string]any{
						"name": "community-module-test",
					},
				},
			},
			Manager: &kyma.Manager{
				Name: "community-module-controller-manager",
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

var (
	resourcesNamespace = `apiVersion: v1
kind: Namespace
metadata:
  name: community-module-system
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
  name: community-module-controller-manager
  namespace: community-module-system
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
			"status": map[string]any{
				"conditions": []any{
					map[string]any{
						"type":   "Available",
						"status": "True",
					},
					map[string]any{
						"type":   "Progressing",
						"status": "True",
					},
				},
			},
		},
	}

	runningManagerMockWithoutVersionLabel = unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]any{
				"namespace": "community-module-system",
				"labels":    map[string]any{},
				"name":      "community-module-controller-manager3",
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
								"image": "http://repo.url/manager:0.0.2",
							},
						},
					},
				},
			},
			"status": map[string]any{
				"conditions": []any{
					map[string]any{
						"type":   "Available",
						"status": "True",
					},
					map[string]any{
						"type":   "Progressing",
						"status": "True",
					},
				},
			},
		},
	}

	runningModuleDeploymentMock = unstructured.Unstructured{
		Object: map[string]any{
			"status": map[string]any{
				"state": "Ready",
			},
		},
	}
)

func TestListInstalled_NoCommunityModules(t *testing.T) {
	fakeClient := &fake.KubeClient{}

	fakeModuleTemplatesRepo := &modulesfake.ModuleTemplatesRepo{
		ReturnCommunity: []kyma.ModuleTemplate{},
	}

	modules, err := listCommunityInstalled(context.Background(), fakeClient, fakeModuleTemplatesRepo, ModulesList{})
	require.NoError(t, err)
	require.Len(t, modules, 0)
}

func TestListInstalled_CommunityModuleInstalledNotRunning(t *testing.T) {
	testServer := getTestHttpServerWithResponse(
		getResourcesUrlResponseYaml(resourcesNamespace, resourcesCRD, resourcesDeployment),
	)
	defer testServer.Close()
	template := getTestCommunityModuleTemplateWithResourcesLink(testServer.URL)

	fakeRootless := &fake.RootlessDynamicClient{
		ReturnGetObj: runningManagerMock,
		ReturnListObjs: &unstructured.UnstructuredList{
			Items: []unstructured.Unstructured{},
		},
	}
	fakeClient := &fake.KubeClient{
		TestRootlessDynamicInterface: fakeRootless,
	}
	fakeModuleTemplatesRepo := &modulesfake.ModuleTemplatesRepo{
		ReturnCommunity:        []kyma.ModuleTemplate{template},
		ReturnInstalledManager: &runningManagerMock,
	}

	modules, err := listCommunityInstalled(context.Background(), fakeClient, fakeModuleTemplatesRepo, ModulesList{})

	require.NoError(t, err)
	require.Len(t, modules, 1)
	require.Equal(t, "communitymodule", modules[0].Name)
	require.Equal(t, "0.0.1", modules[0].InstallDetails.Version)
	require.Equal(t, "NotRunning", modules[0].InstallDetails.ModuleState)
	require.Equal(t, "Ready", modules[0].InstallDetails.InstallationState)
}

func TestListInstalled_CommunityModuleInstalledRunning(t *testing.T) {
	testServer := getTestHttpServerWithResponse(
		getResourcesUrlResponseYaml(resourcesNamespace, resourcesCRD, resourcesDeployment),
	)
	defer testServer.Close()

	template := getTestCommunityModuleTemplateWithResourcesLink(testServer.URL)

	fakeRootless := &fake.RootlessDynamicClient{
		ReturnGetObj: runningManagerMock,
		ReturnListObjs: &unstructured.UnstructuredList{
			Items: []unstructured.Unstructured{
				runningModuleDeploymentMock,
			},
		},
	}
	fakeClient := &fake.KubeClient{
		TestRootlessDynamicInterface: fakeRootless,
	}
	fakeModuleTemplatesRepo := &modulesfake.ModuleTemplatesRepo{
		ReturnCommunity:        []kyma.ModuleTemplate{template},
		ReturnInstalledManager: &runningManagerMock,
	}

	modules, err := listCommunityInstalled(context.Background(), fakeClient, fakeModuleTemplatesRepo, ModulesList{})

	require.NoError(t, err)
	require.Len(t, modules, 1)
	require.Equal(t, "communitymodule", modules[0].Name)
	require.Equal(t, "0.0.1", modules[0].InstallDetails.Version)
	require.Equal(t, "Ready", modules[0].InstallDetails.ModuleState)
	require.Equal(t, "Ready", modules[0].InstallDetails.InstallationState)
}

func TestListInstalled_CommunityModuleVersionFromImage(t *testing.T) {
	testServer := getTestHttpServerWithResponse(
		getResourcesUrlResponseYaml(resourcesNamespace, resourcesCRD, resourcesDeployment),
	)
	defer testServer.Close()

	template := getTestCommunityModuleTemplateWithResourcesLink(testServer.URL)

	fakeRootless := &fake.RootlessDynamicClient{
		ReturnGetObj: runningManagerMockWithoutVersionLabel,
		ReturnListObjs: &unstructured.UnstructuredList{
			Items: []unstructured.Unstructured{
				runningModuleDeploymentMock,
			},
		},
	}
	fakeClient := &fake.KubeClient{
		TestRootlessDynamicInterface: fakeRootless,
	}
	fakeModuleTemplatesRepo := &modulesfake.ModuleTemplatesRepo{
		ReturnCommunity:        []kyma.ModuleTemplate{template},
		ReturnInstalledManager: &runningManagerMockWithoutVersionLabel,
	}

	modules, err := listCommunityInstalled(context.Background(), fakeClient, fakeModuleTemplatesRepo, ModulesList{})

	require.NoError(t, err)
	require.Len(t, modules, 1)
	require.Equal(t, "communitymodule", modules[0].Name)
	require.Equal(t, "0.0.2", modules[0].InstallDetails.Version)
	require.Equal(t, "Ready", modules[0].InstallDetails.ModuleState)
}

func getResourcesUrlResponseYaml(parts ...string) string {
	return strings.Join(parts, "\n---\n")
}
