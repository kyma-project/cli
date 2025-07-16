package modules

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	modulesfake "github.com/kyma-project/cli.v3/internal/modules/fake"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamic_fake "k8s.io/client-go/dynamic/fake"
	"k8s.io/utils/ptr"
)

var (
	testModuleTemplate1 = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleTemplate",
			"metadata": map[string]interface{}{
				"name":      "serverless-1",
				"namespace": "kyma-system",
				"labels": map[string]interface{}{
					"operator.kyma-project.io/managed-by": "kyma",
				},
			},
			"spec": map[string]interface{}{
				"moduleName": "serverless",
				"version":    "0.0.1",
				"data":       testServerless.Object,
				"info": map[string]interface{}{
					"repository": "url-1",
				},
			},
		},
	}

	testModuleTemplate2 = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleTemplate",
			"metadata": map[string]interface{}{
				"name":      "serverless-2",
				"namespace": "kyma-system",
				"labels": map[string]interface{}{
					"operator.kyma-project.io/managed-by": "kyma",
				},
			},
			"spec": map[string]interface{}{
				"moduleName": "serverless",
				"version":    "0.0.2",
				"info": map[string]interface{}{
					"repository": "url-2",
				},
			},
		},
	}

	testModuleTemplate3 = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleTemplate",
			"metadata": map[string]interface{}{
				"name":      "keda-1",
				"namespace": "kyma-system",
				"labels": map[string]interface{}{
					"operator.kyma-project.io/managed-by": "kyma",
				},
			},
			"spec": map[string]interface{}{
				"moduleName": "keda",
				"version":    "0.1",
				"info": map[string]interface{}{
					"repository": "url-3",
				},
			},
		},
	}

	testModuleTemplate4 = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleTemplate",
			"metadata": map[string]interface{}{
				"name":      "keda-2",
				"namespace": "kyma-system",
				"labels": map[string]interface{}{
					"operator.kyma-project.io/managed-by": "kyma",
				},
			},
			"spec": map[string]interface{}{
				"moduleName": "keda",
				"version":    "0.2",
			},
		},
	}

	// corrupted one - without spec
	testModuleTemplate5 = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleTemplate",
			"metadata": map[string]interface{}{
				"name":      "keda-3",
				"namespace": "kyma-system",
				"labels": map[string]interface{}{
					"operator.kyma-project.io/managed-by": "kyma",
				},
			},
		},
	}

	testReleaseMeta1 = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleReleaseMeta",
			"metadata": map[string]interface{}{
				"name":      "serverless",
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"moduleName": "serverless",
				"channels": []interface{}{
					map[string]interface{}{
						"version": "0.0.1",
						"channel": "fast",
					},
				},
			},
		},
	}

	testReleaseMeta2 = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleReleaseMeta",
			"metadata": map[string]interface{}{
				"name":      "keda",
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"moduleName": "keda",
				"channels": []interface{}{
					map[string]interface{}{
						"version": "0.1",
						"channel": "regular",
					},
					map[string]interface{}{
						"version": "0.2",
						"channel": "fast",
					},
				},
			},
		},
	}

	testKymaCR = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "Kyma",
			"metadata": map[string]interface{}{
				"name":      kyma.DefaultKymaName,
				"namespace": kyma.DefaultKymaNamespace,
			},
			"spec": map[string]interface{}{
				"channel": "fast",
				"modules": []interface{}{
					map[string]interface{}{
						"name":                 "serverless",
						"managed":              true,
						"customResourcePolicy": "Ignore",
					},
					map[string]interface{}{
						"name":    "keda",
						"managed": false,
					},
				},
			},
			"status": map[string]interface{}{
				"modules": []interface{}{
					map[string]interface{}{
						"name":     "serverless",
						"version":  "0.0.1",
						"channel":  "fast",
						"state":    "Ready",
						"template": testServerless.Object,
					},
					map[string]interface{}{
						"name":    "keda",
						"version": "0.2",
						"channel": "fast",
						"state":   "Unmanaged",
					},
				},
			},
		},
	}

	testInstalledModuleList = []Module{
		{
			Name: "serverless",
			InstallDetails: ModuleInstallDetails{
				Managed:              ManagedTrue,
				Channel:              "fast",
				Version:              "0.0.1",
				ModuleState:          "Ready",
				InstallationState:    "Ready",
				CustomResourcePolicy: "Ignore",
			},
		},
		{
			Name: "keda",
			InstallDetails: ModuleInstallDetails{
				Managed:              ManagedFalse,
				Channel:              "fast",
				Version:              "0.2",
				ModuleState:          "Unmanaged",
				InstallationState:    "Unmanaged",
				CustomResourcePolicy: "CreateAndDelete",
			},
		},
	}

	testModuleList = []Module{
		{
			Name: "keda",
			Versions: []ModuleVersion{
				{
					Repository: "url-3",
					Version:    "0.1",
					Channel:    "regular",
				},
				{
					Version: "0.2",
					Channel: "fast",
				},
			},
			CommunityModule: false,
		},
		{
			Name: "serverless",
			Versions: []ModuleVersion{
				{
					Repository: "url-1",
					Version:    "0.0.1",
					Channel:    "fast",
				},
				{
					Repository: "url-2",
					Version:    "0.0.2",
				},
			},
			CommunityModule: false,
		},
	}

	GVRServerless = schema.GroupVersionResource{
		Group:    "operator.kyma-project.io",
		Version:  "v1alpha1",
		Resource: "serverlesses",
	}

	testServerless = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1alpha1",
			"kind":       "Serverless",
			"metadata": map[string]interface{}{
				"name":      "serverless-1",
				"namespace": "kyma-system",
			},
			"status": map[string]interface{}{
				"state": "Ready",
			},
		},
	}

	GVRDeployment = schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}

	testDeploymentDataState = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "serverless-state",
				"namespace": "kyma-system",
			},
			"status": map[string]interface{}{
				"state": "Ready",
			},
		},
	}

	testDeploymentDataConditions = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "serverless-conditions",
				"namespace": "kyma-system",
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Available",
						"status": "True",
					},
				},
			},
		},
	}

	testDeploymentDataConditionsProcessing = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "serverless-conditions-processing",
				"namespace": "kyma-system",
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Processing",
						"status": "True",
					},
				},
			},
		},
	}

	testDeploymentDataReady = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "serverless-replicas-ready",
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"replicas": int64(1),
			},
			"status": map[string]interface{}{
				"readyReplicas": int64(1),
			},
		},
	}

	testDeploymentDataProcessing = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "serverless-replicas-processing",
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"replicas": int64(2),
			},
			"status": map[string]interface{}{
				"readyReplicas": int64(1),
			},
		},
	}

	testDeploymentDataDeleting = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "serverless-replicas-deleting",
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"replicas": int64(1),
			},
			"status": map[string]interface{}{
				"readyReplicas": int64(2),
			},
		},
	}
)

func TestListInstalled(t *testing.T) {
	t.Run("list managed modules from cluster", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(kyma.GVRModuleTemplate.GroupVersion())
		scheme.AddKnownTypes(kyma.GVRModuleReleaseMeta.GroupVersion())
		scheme.AddKnownTypes(kyma.GVRKyma.GroupVersion())
		dynamicClient := dynamic_fake.NewSimpleDynamicClient(scheme,
			&testModuleTemplate1,
			&testKymaCR,
		)

		fakeRootless := &fake.RootlessDynamicClient{
			ReturnGetObj: testServerless,
		}

		fakeClient := &fake.KubeClient{
			TestKymaInterface:            kyma.NewClient(dynamicClient),
			TestRootlessDynamicInterface: fakeRootless,
		}
		fakeModuleTemplatesRepo := &modulesfake.ModuleTemplatesRepo{}

		modules, err := ListInstalled(context.Background(), fakeClient, fakeModuleTemplatesRepo)

		require.NoError(t, err)
		require.Equal(t, ModulesList(testInstalledModuleList), modules)
	})
}

func TestListCatalog(t *testing.T) {
	t.Run("list modules catalog from cluster", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(kyma.GVRModuleTemplate.GroupVersion())
		scheme.AddKnownTypes(kyma.GVRModuleReleaseMeta.GroupVersion())
		dynamicClient := dynamic_fake.NewSimpleDynamicClient(scheme,
			&testModuleTemplate1,
			&testModuleTemplate2,
			&testModuleTemplate3,
			&testModuleTemplate4,
			&testReleaseMeta1,
			&testReleaseMeta2,
		)

		fakeRootless := &fake.RootlessDynamicClient{}

		fakeClient := &fake.KubeClient{
			TestKymaInterface:            kyma.NewClient(dynamicClient),
			TestRootlessDynamicInterface: fakeRootless,
		}

		modules, err := ListCatalog(context.Background(), fakeClient)

		require.NoError(t, err)
		require.Equal(t, ModulesList(testModuleList), modules)
	})

	t.Run("ignore corrupted ModuleTemplate", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(kyma.GVRModuleTemplate.GroupVersion())
		scheme.AddKnownTypes(kyma.GVRModuleReleaseMeta.GroupVersion())
		dynamicClient := dynamic_fake.NewSimpleDynamicClient(scheme,
			&testModuleTemplate1,
			&testModuleTemplate2,
			&testModuleTemplate3,
			&testModuleTemplate4,
			&testModuleTemplate5, // corrupted ModuleTemplate
			&testReleaseMeta1,
			&testReleaseMeta2,
		)

		fakeRootless := &fake.RootlessDynamicClient{}

		fakeClient := &fake.KubeClient{
			TestKymaInterface:            kyma.NewClient(dynamicClient),
			TestRootlessDynamicInterface: fakeRootless,
		}

		modules, err := ListCatalog(context.Background(), fakeClient)

		require.NoError(t, err)
		require.Equal(t, ModulesList(testModuleList), modules)
	})
}

func TestModuleInstallationStatus(t *testing.T) {
	for _, tt := range []struct {
		name         string
		moduleSpec   *kyma.Module
		moduleStatus kyma.ModuleStatus
		client       kube.Client

		expectedState string
		expectedError string
	}{
		{
			name:       "module is under deletion",
			moduleSpec: nil,
			moduleStatus: kyma.ModuleStatus{
				State: "Deleting",
			},
			expectedState: "Deleting",
		},
		{
			name: "module CR is managed by klm",
			moduleSpec: &kyma.Module{
				CustomResourcePolicy: "CreateAndDelete",
			},
			moduleStatus: kyma.ModuleStatus{
				State: "Ready",
			},
			expectedState: "Ready",
		},
		{
			name: "module in unmanaged",
			moduleSpec: &kyma.Module{
				Managed: ptr.To(false),
			},
			moduleStatus: kyma.ModuleStatus{
				State: "Unmanaged",
			},
			expectedState: "Unmanaged",
		},
		{
			name:       "ModuleTemplate not found error",
			moduleSpec: &kyma.Module{},
			moduleStatus: kyma.ModuleStatus{
				Template: testModuleTemplate1,
			},
			client: func() kube.Client {
				return &fake.KubeClient{
					TestKymaInterface: &fake.KymaClient{
						ReturnGetModuleTemplateErr: errors.New("not found"),
					},
				}
			}(),
			expectedError: "failed to get ModuleTemplate kyma-system/serverless-1: not found",
		},
		{
			name:       "get state from module CR",
			moduleSpec: &kyma.Module{},
			moduleStatus: kyma.ModuleStatus{
				Template: testModuleTemplate1,
			},
			client: func() kube.Client {
				return &fake.KubeClient{
					TestKymaInterface: &fake.KymaClient{
						ReturnModuleTemplate: kyma.ModuleTemplate{
							Spec: kyma.ModuleTemplateSpec{
								Data: testServerless,
							},
						},
					},
					TestRootlessDynamicInterface: &fake.RootlessDynamicClient{
						ReturnGetObj: testServerless,
					},
				}
			}(),
			expectedState: "Ready",
		},
		{
			name:       "get ready state from module managers conditions",
			moduleSpec: &kyma.Module{},
			moduleStatus: kyma.ModuleStatus{
				Template: testModuleTemplate1,
			},
			client:        fixKymaClientForManager(testDeploymentDataConditions),
			expectedState: "Ready",
		},
		{
			name:       "get ready state from module managers state",
			moduleSpec: &kyma.Module{},
			moduleStatus: kyma.ModuleStatus{
				Template: testModuleTemplate1,
			},
			client:        fixKymaClientForManager(testDeploymentDataState),
			expectedState: "Ready",
		},
		{
			name:       "get processing state from module manager",
			moduleSpec: &kyma.Module{},
			moduleStatus: kyma.ModuleStatus{
				Template: testModuleTemplate1,
			},
			client:        fixKymaClientForManager(testDeploymentDataConditionsProcessing),
			expectedState: "Processing",
		},
		{
			name:       "get processing state from module managers replicas",
			moduleSpec: &kyma.Module{},
			moduleStatus: kyma.ModuleStatus{
				Template: testModuleTemplate1,
			},
			client:        fixKymaClientForManager(testDeploymentDataProcessing),
			expectedState: "Processing",
		},
		{
			name:       "get ready state from module managers replicas",
			moduleSpec: &kyma.Module{},
			moduleStatus: kyma.ModuleStatus{
				Template: testModuleTemplate1,
			},
			client:        fixKymaClientForManager(testDeploymentDataReady),
			expectedState: "Ready",
		},
		{
			name:       "get deleting state from module managers replicas",
			moduleSpec: &kyma.Module{},
			moduleStatus: kyma.ModuleStatus{
				Template: testModuleTemplate1,
			},
			client:        fixKymaClientForManager(testDeploymentDataDeleting),
			expectedState: "Deleting",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			state, err := getModuleInstallationState(context.Background(), tt.client, tt.moduleStatus, tt.moduleSpec)
			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
			}
			require.Equal(t, tt.expectedState, state)
		})
	}
}

func fixKymaClientForManager(manager unstructured.Unstructured) kube.Client {
	return &fake.KubeClient{
		TestKymaInterface: &fake.KymaClient{
			ReturnModuleTemplate: kyma.ModuleTemplate{
				Spec: kyma.ModuleTemplateSpec{
					Manager: &kyma.Manager{
						GroupVersionKind: metav1.GroupVersionKind(manager.GroupVersionKind()),
						Namespace:        manager.GetNamespace(),
						Name:             manager.GetName(),
					},
				},
			},
		},
		TestRootlessDynamicInterface: &fake.RootlessDynamicClient{
			ReturnGetObj: manager,
		},
	}
}
