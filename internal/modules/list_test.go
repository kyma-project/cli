package modules

import (
	"context"
	"testing"

	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgo_fake "k8s.io/client-go/discovery/fake"
	dynamic_fake "k8s.io/client-go/dynamic/fake"
	clientgo_testing "k8s.io/client-go/testing"
)

var (
	testModuleTemplate1 = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleTemplate",
			"metadata": map[string]interface{}{
				"name":      "serverless-1",
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"moduleName": "serverless",
				"version":    "0.0.1",
				"info": map[string]interface{}{
					"repository": "url-1",
				},
			},
			"status": map[string]interface{}{
				"state": "ready",
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
						"name":    "serverless",
						"managed": false,
					},
					map[string]interface{}{
						"name":    "keda",
						"managed": true,
					},
				},
			},
			"status": map[string]interface{}{
				"modules": []interface{}{
					map[string]interface{}{
						"name":    "serverless",
						"version": "0.0.1",
						"state":   "Ready",
					},
					map[string]interface{}{
						"name":    "keda",
						"version": "0.2",
					},
				},
			},
		},
	}

	testManagedModuleList = []Module{
		{
			Name: "keda",
			InstallDetails: ModuleInstallDetails{
				Managed: ManagedTrue,
				Channel: "fast",
				Version: "0.2",
			},
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
		},
		{
			Name: "serverless",
			InstallDetails: ModuleInstallDetails{
				Managed: ManagedFalse,
				Channel: "fast",
				Version: "0.0.1",
				State:   "Ready",
			},
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
		},
	}

	testModuleDataResourceList = &metav1.APIResourceList{
		GroupVersion: "operator.kyma-project.io/v1alpha1",
		APIResources: []metav1.APIResource{
			{
				Group:        "operator.kyma-project.io",
				Version:      "v1alpha1",
				Kind:         "Serverless",
				SingularName: "serverless",
				Name:         "serverlesses",
				Namespaced:   true,
			},
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

	testModuleManagerResourceList = &metav1.APIResourceList{
		GroupVersion: "apps/v1",
		APIResources: []metav1.APIResource{
			{
				Group:        "apps",
				Version:      "v1",
				Kind:         "Deployment",
				SingularName: "deployment",
				Name:         "deployments",
				Namespaced:   true,
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
	replicasCountOne        int64 = 1
	replicasCountTwo        int64 = 2
	testDeploymentDataReady       = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "serverless-replicas-ready",
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"replicas": replicasCountOne,
			},
			"status": map[string]interface{}{
				"readyReplicas": replicasCountOne,
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
				"replicas": replicasCountTwo,
			},
			"status": map[string]interface{}{
				"readyReplicas": replicasCountOne,
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
				"replicas": replicasCountOne,
			},
			"status": map[string]interface{}{
				"readyReplicas": replicasCountTwo,
			},
		},
	}
)

func TestList(t *testing.T) {
	t.Run("list modules from cluster without Kyma CR", func(t *testing.T) {
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

		modules, err := List(context.Background(), fakeClient)

		require.NoError(t, err)
		require.Equal(t, ModulesList(testModuleList), modules)
	})

	t.Run("list managed modules from cluster", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(kyma.GVRModuleTemplate.GroupVersion())
		scheme.AddKnownTypes(kyma.GVRModuleReleaseMeta.GroupVersion())
		scheme.AddKnownTypes(kyma.GVRKyma.GroupVersion())
		dynamicClient := dynamic_fake.NewSimpleDynamicClient(scheme,
			&testModuleTemplate1,
			&testModuleTemplate2,
			&testModuleTemplate3,
			&testModuleTemplate4,
			&testReleaseMeta1,
			&testReleaseMeta2,
			&testKymaCR,
		)

		fakeRootless := &fake.RootlessDynamicClient{}

		fakeClient := &fake.KubeClient{
			TestKymaInterface:            kyma.NewClient(dynamicClient),
			TestRootlessDynamicInterface: fakeRootless,
		}

		modules, err := List(context.Background(), fakeClient)

		require.NoError(t, err)
		require.Equal(t, ModulesList(testManagedModuleList), modules)
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

		modules, err := List(context.Background(), fakeClient)

		require.NoError(t, err)
		require.Equal(t, ModulesList(testModuleList), modules)
	})
}

func TestModuleStatus(t *testing.T) {
	tests := []struct {
		name            string
		kyma            *kyma.Kyma
		moduleTemplate  kyma.ModuleTemplate
		moduleOrManager *unstructured.Unstructured
		wantedErr       error
		expectedStatus  string
	}{
		{
			name: "module is Ready, from default Kyma",
			kyma: &kyma.Kyma{
				Status: kyma.KymaStatus{
					Modules: []kyma.ModuleStatus{
						{
							Name:  "serverless",
							State: "Ready",
						},
					},
				},
			},
			moduleTemplate: kyma.ModuleTemplate{
				Spec: kyma.ModuleTemplateSpec{
					ModuleName: "serverless",
				},
			},
			expectedStatus: "Ready",
			wantedErr:      nil,
		}, {
			name: "module is in Error state, from Kyma",
			kyma: &kyma.Kyma{
				Status: kyma.KymaStatus{
					Modules: []kyma.ModuleStatus{
						{
							Name:  "serverless",
							State: "Error",
						},
					},
				},
			},
			moduleTemplate: kyma.ModuleTemplate{
				Spec: kyma.ModuleTemplateSpec{
					ModuleName: "serverless",
				},
			},
			expectedStatus: "Error",
			wantedErr:      nil,
		},
		{
			name: "module is Ready, from moduleTemplate.spec.data",
			kyma: &kyma.Kyma{
				Status: kyma.KymaStatus{
					Modules: []kyma.ModuleStatus{
						{
							Name:  "serverless-1",
							State: "",
						},
					},
				},
			},
			moduleTemplate: kyma.ModuleTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "serverless",
				},
				Spec: kyma.ModuleTemplateSpec{
					Data: unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "operator.kyma-project.io/v1alpha1",
							"kind":       "Serverless",
							"metadata": map[string]interface{}{
								"name":      "serverless-1",
								"namespace": "kyma-system",
							},
						},
					},
				},
			},
			expectedStatus: "Ready",
			wantedErr:      nil,
		},
		{
			name: "module is Ready, from moduleTemplate.spec.manager, state",
			kyma: &kyma.Kyma{
				Status: kyma.KymaStatus{
					Modules: []kyma.ModuleStatus{
						{
							Name:  "serverless-state",
							State: "",
						},
					},
				},
			},
			moduleTemplate: kyma.ModuleTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "serverless",
				},
				Spec: kyma.ModuleTemplateSpec{
					Manager: &kyma.Manager{
						GroupVersionKind: metav1.GroupVersionKind{
							Group:   "apps",
							Version: "v1",
							Kind:    "Deployment",
						},
						Name:      "serverless-state",
						Namespace: "kyma-system",
					},
				},
			},
			expectedStatus: "Ready",
			wantedErr:      nil,
		},
		{
			name: "module is Ready, from moduleTemplate.spec.manager, conditions",
			kyma: &kyma.Kyma{
				Status: kyma.KymaStatus{
					Modules: []kyma.ModuleStatus{
						{
							Name:  "serverless-1",
							State: "",
						},
					},
				},
			},
			moduleTemplate: kyma.ModuleTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "serverless",
				},
				Spec: kyma.ModuleTemplateSpec{
					Manager: &kyma.Manager{
						GroupVersionKind: metav1.GroupVersionKind{
							Group:   "apps",
							Version: "v1",
							Kind:    "Deployment",
						},
						Name:      "serverless-conditions",
						Namespace: "kyma-system",
					},
				},
			},
			expectedStatus: "Ready",
			wantedErr:      nil,
		},
		{
			name: "module is Ready, from moduleTemplate.spec.manager, readyReplicas",
			kyma: &kyma.Kyma{
				Status: kyma.KymaStatus{
					Modules: []kyma.ModuleStatus{
						{
							Name:  "serverless-1",
							State: "",
						},
					},
				},
			},
			moduleTemplate: kyma.ModuleTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "serverless",
				},
				Spec: kyma.ModuleTemplateSpec{
					Manager: &kyma.Manager{
						GroupVersionKind: metav1.GroupVersionKind{
							Group:   "apps",
							Version: "v1",
							Kind:    "Deployment",
						},
						Name:      "serverless-replicas-ready",
						Namespace: "kyma-system",
					},
				},
			},
			expectedStatus: "Ready",
			wantedErr:      nil,
		},
		{
			name: "module is Processing, from moduleTemplate.spec.manager, readyReplicas",
			kyma: &kyma.Kyma{
				Status: kyma.KymaStatus{
					Modules: []kyma.ModuleStatus{
						{
							Name:  "serverless-2",
							State: "",
						},
					},
				},
			},
			moduleTemplate: kyma.ModuleTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "serverless",
				},
				Spec: kyma.ModuleTemplateSpec{
					Manager: &kyma.Manager{
						GroupVersionKind: metav1.GroupVersionKind{
							Group:   "apps",
							Version: "v1",
							Kind:    "Deployment",
						},
						Name:      "serverless-replicas-processing",
						Namespace: "kyma-system",
					},
				},
			},
			expectedStatus: "Processing",
			wantedErr:      nil,
		},
		{
			name: "module is Deleting, from moduleTemplate.spec.manager, readyReplicas",
			kyma: &kyma.Kyma{
				Status: kyma.KymaStatus{
					Modules: []kyma.ModuleStatus{
						{
							Name:  "serverless",
							State: "",
						},
					},
				},
			},
			moduleTemplate: kyma.ModuleTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "serverless",
				},
				Spec: kyma.ModuleTemplateSpec{
					Manager: &kyma.Manager{
						GroupVersionKind: metav1.GroupVersionKind{
							Group:   "apps",
							Version: "v1",
							Kind:    "Deployment",
						},
						Name:      "serverless-replicas-deleting",
						Namespace: "kyma-system",
					},
				},
			},
			expectedStatus: "Deleting",
			wantedErr:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			scheme.AddKnownTypes(GVRDeployment.GroupVersion())
			scheme.AddKnownTypes(GVRServerless.GroupVersion())
			apiResources := []*metav1.APIResourceList{
				testModuleDataResourceList, testModuleManagerResourceList,
			}
			dynamicFake := dynamic_fake.NewSimpleDynamicClient(scheme, &testServerless, &testDeploymentDataState, &testDeploymentDataConditions, &testDeploymentDataReady, &testDeploymentDataProcessing, &testDeploymentDataDeleting)

			fakeRootless := rootlessdynamic.NewClient(
				dynamicFake,
				&clientgo_fake.FakeDiscovery{
					Fake: &clientgo_testing.Fake{
						Resources: apiResources,
					},
				},
			)

			fakeClient := &fake.KubeClient{
				TestRootlessDynamicInterface: fakeRootless,
			}
			state, err := getModuleState(context.Background(), fakeClient, tt.moduleTemplate, tt.kyma)
			require.Equal(t, tt.wantedErr, err)
			require.Equal(t, tt.expectedStatus, state)
		})
	}
}
