package modules

import (
	"context"
	"strings"
	"testing"

	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamic_fake "k8s.io/client-go/dynamic/fake"
)

func getTestCommunityModuleTemplateWithResourcesLink(link string) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleTemplate",
			"metadata": map[string]any{
				"name":      "communitymodule-001",
				"namespace": "kyma-system",
			},
			"spec": map[string]any{
				"moduleName": "communitymodule",
				"version":    "0.0.1",
				"data": map[string]any{
					"apiVersion": "operator.kyma-project.io/v1alpha1",
					"kind":       "CommunityModule",
					"metadata": map[string]any{
						"name": "community-module-test",
					},
				},
				"manager": map[string]any{
					"name":    "community-module-controller-manager",
					"group":   "apps",
					"version": "v1",
					"kind":    "Deployment",
				},
				"resources": []any{
					map[string]any{
						"name": "rawManifest",
						"link": link,
					},
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

	emptyKymaCustomResource = unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "Kyma",
			"metadata": map[string]any{
				"name":      kyma.DefaultKymaName,
				"namespace": kyma.DefaultKymaNamespace,
			},
			"spec":   map[string]any{},
			"status": map[string]any{},
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
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(kyma.GVRKyma.GroupVersion())
	scheme.AddKnownTypes(kyma.GVRModuleTemplate.GroupVersion())
	scheme.AddKnownTypes(kyma.GVRModuleReleaseMeta.GroupVersion())

	gvrToListKind := map[schema.GroupVersionResource]string{
		kyma.GVRModuleTemplate:    "ModuleTemplateList",
		kyma.GVRKyma:              "KymaList",
		kyma.GVRModuleReleaseMeta: "ModuleReleaseMetaList",
	}

	dynamicClient := dynamic_fake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		gvrToListKind,
		&emptyKymaCustomResource,
	)

	fakeRootless := &fake.RootlessDynamicClient{}
	fakeClient := &fake.KubeClient{
		TestKymaInterface:            kyma.NewClient(dynamicClient),
		TestRootlessDynamicInterface: fakeRootless,
	}

	modules, err := ListInstalled(context.Background(), fakeClient)
	require.NoError(t, err)
	require.Len(t, modules, 0)
}

func TestListInstalled_CommunityModuleInstalledNotRunning(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(kyma.GVRKyma.GroupVersion())
	scheme.AddKnownTypes(kyma.GVRModuleTemplate.GroupVersion())
	scheme.AddKnownTypes(kyma.GVRModuleReleaseMeta.GroupVersion())
	gvrToListKind := map[schema.GroupVersionResource]string{
		kyma.GVRModuleTemplate:    "ModuleTemplateList",
		kyma.GVRKyma:              "KymaList",
		kyma.GVRModuleReleaseMeta: "ModuleReleaseMetaList",
	}

	testServer := getTestHttpServerWithResponse(
		getResourcesUrlResponseYaml(resourcesNamespace, resourcesCRD, resourcesDeployment),
	)
	defer testServer.Close()

	template := getTestCommunityModuleTemplateWithResourcesLink(testServer.URL)

	dynamicClient := dynamic_fake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		gvrToListKind,
		&emptyKymaCustomResource,
		&template,
	)

	fakeRootless := &fake.RootlessDynamicClient{
		ReturnGetObj: runningManagerMock,
		ReturnListObjs: &unstructured.UnstructuredList{
			Items: []unstructured.Unstructured{},
		},
	}
	fakeClient := &fake.KubeClient{
		TestKymaInterface:            kyma.NewClient(dynamicClient),
		TestRootlessDynamicInterface: fakeRootless,
	}

	modules, err := ListInstalled(context.Background(), fakeClient)
	require.NoError(t, err)
	require.Len(t, modules, 1)
	require.Equal(t, "communitymodule", modules[0].Name)
	require.Equal(t, "0.0.1", modules[0].InstallDetails.Version)
	require.Equal(t, "NotRunning", modules[0].InstallDetails.State)
}

func TestListInstalled_CommunityModuleInstalledRunning(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(kyma.GVRKyma.GroupVersion())
	scheme.AddKnownTypes(kyma.GVRModuleTemplate.GroupVersion())
	scheme.AddKnownTypes(kyma.GVRModuleReleaseMeta.GroupVersion())
	gvrToListKind := map[schema.GroupVersionResource]string{
		kyma.GVRModuleTemplate:    "ModuleTemplateList",
		kyma.GVRKyma:              "KymaList",
		kyma.GVRModuleReleaseMeta: "ModuleReleaseMetaList",
	}

	testServer := getTestHttpServerWithResponse(
		getResourcesUrlResponseYaml(resourcesNamespace, resourcesCRD, resourcesDeployment),
	)
	defer testServer.Close()

	template := getTestCommunityModuleTemplateWithResourcesLink(testServer.URL)

	dynamicClient := dynamic_fake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		gvrToListKind,
		&emptyKymaCustomResource,
		&template,
	)

	fakeRootless := &fake.RootlessDynamicClient{
		ReturnGetObj: runningManagerMock,
		ReturnListObjs: &unstructured.UnstructuredList{
			Items: []unstructured.Unstructured{
				runningModuleDeploymentMock,
			},
		},
	}
	fakeClient := &fake.KubeClient{
		TestKymaInterface:            kyma.NewClient(dynamicClient),
		TestRootlessDynamicInterface: fakeRootless,
	}

	modules, err := ListInstalled(context.Background(), fakeClient)
	require.NoError(t, err)
	require.Len(t, modules, 1)
	require.Equal(t, "communitymodule", modules[0].Name)
	require.Equal(t, "0.0.1", modules[0].InstallDetails.Version)
	require.Equal(t, "Ready", modules[0].InstallDetails.State)
}

func TestListInstalled_CommunityModuleVersionFromImage(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(kyma.GVRKyma.GroupVersion())
	scheme.AddKnownTypes(kyma.GVRModuleTemplate.GroupVersion())
	scheme.AddKnownTypes(kyma.GVRModuleReleaseMeta.GroupVersion())
	gvrToListKind := map[schema.GroupVersionResource]string{
		kyma.GVRModuleTemplate:    "ModuleTemplateList",
		kyma.GVRKyma:              "KymaList",
		kyma.GVRModuleReleaseMeta: "ModuleReleaseMetaList",
	}

	testServer := getTestHttpServerWithResponse(
		getResourcesUrlResponseYaml(resourcesNamespace, resourcesCRD, resourcesDeployment),
	)
	defer testServer.Close()

	template := getTestCommunityModuleTemplateWithResourcesLink(testServer.URL)

	dynamicClient := dynamic_fake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		gvrToListKind,
		&emptyKymaCustomResource,
		&template,
	)

	fakeRootless := &fake.RootlessDynamicClient{
		ReturnGetObj: runningManagerMockWithoutVersionLabel,
		ReturnListObjs: &unstructured.UnstructuredList{
			Items: []unstructured.Unstructured{
				runningModuleDeploymentMock,
			},
		},
	}
	fakeClient := &fake.KubeClient{
		TestKymaInterface:            kyma.NewClient(dynamicClient),
		TestRootlessDynamicInterface: fakeRootless,
	}

	modules, err := ListInstalled(context.Background(), fakeClient)
	require.NoError(t, err)
	require.Len(t, modules, 1)
	require.Equal(t, "communitymodule", modules[0].Name)
	require.Equal(t, "0.0.2", modules[0].InstallDetails.Version)
	require.Equal(t, "Ready", modules[0].InstallDetails.State)
}

func getResourcesUrlResponseYaml(parts ...string) string {
	return strings.Join(parts, "\n---\n")
}
