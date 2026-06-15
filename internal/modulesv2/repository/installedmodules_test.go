package repository_test

import (
	"context"
	"fmt"
	"testing"

	kubefake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func newKubeClient(defaultKyma kyma.Kyma, moduleTemplate kyma.ModuleTemplate, crList *unstructured.UnstructuredList) *kubefake.KubeClient {
	kymaClient := &kubefake.KymaClient{
		ReturnDefaultKyma:    defaultKyma,
		ReturnModuleTemplate: moduleTemplate,
	}
	rootlessDynamic := &kubefake.RootlessDynamicClient{
		ReturnListObjs: crList,
	}
	return &kubefake.KubeClient{
		TestKymaInterface:            kymaClient,
		TestRootlessDynamicInterface: rootlessDynamic,
	}
}

func TestModuleInstallationsRepository_ListInstalledModules_NormalCase(t *testing.T) {
	apiGatewayStatus := kyma.ModuleStatus{Name: "api-gateway", State: "Ready"}
	apiGatewayStatus.Template.SetName("api-gateway-template")
	apiGatewayStatus.Template.SetNamespace("kyma-system")
	defaultKyma := kyma.Kyma{
		Spec: kyma.KymaSpec{
			Modules: []kyma.Module{
				{Name: "api-gateway", CustomResourcePolicy: "CreateAndDelete"},
			},
		},
		Status: kyma.KymaStatus{
			Modules: []kyma.ModuleStatus{apiGatewayStatus},
		},
	}
	moduleTemplate := kyma.ModuleTemplate{
		Spec: kyma.ModuleTemplateSpec{
			Data: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "gateway.kyma-project.io/v1alpha1",
					"kind":       "APIGateway",
				},
			},
		},
	}
	crList := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{Object: map[string]interface{}{"status": map[string]interface{}{"state": "Ready"}}},
		},
	}
	kubeClient := newKubeClient(defaultKyma, moduleTemplate, crList)
	repo := repository.NewModuleInstallationsRepository(kubeClient)

	result, err := repo.ListInstalledModules(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 1)
	module := result[0]
	require.Equal(t, "api-gateway", module.Name)
	require.Equal(t, "Ready", module.ModuleState)
	require.Equal(t, "CreateAndDelete", module.CustomResourcePolicy)
}

func TestModuleInstallationsRepository_ListInstalledModules_ModuleBeingAdded(t *testing.T) {
	defaultKyma := kyma.Kyma{
		Spec: kyma.KymaSpec{
			Modules: []kyma.Module{
				{Name: "api-gateway", CustomResourcePolicy: "CreateAndDelete"},
			},
		},
		Status: kyma.KymaStatus{},
	}
	kubeClient := newKubeClient(defaultKyma, kyma.ModuleTemplate{}, nil)
	repo := repository.NewModuleInstallationsRepository(kubeClient)

	result, err := repo.ListInstalledModules(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 1)
	module := result[0]
	require.Equal(t, "api-gateway", module.Name)
	require.Equal(t, "", module.ModuleState)
	require.Equal(t, "CreateAndDelete", module.CustomResourcePolicy)
}

func TestModuleInstallationsRepository_ListInstalledModules_ModuleBeingDeleted(t *testing.T) {
	deletingStatus := kyma.ModuleStatus{Name: "api-gateway", State: "Deleting"}
	deletingStatus.Template.SetName("api-gateway-template")
	deletingStatus.Template.SetNamespace("kyma-system")
	defaultKyma := kyma.Kyma{
		Spec: kyma.KymaSpec{},
		Status: kyma.KymaStatus{
			Modules: []kyma.ModuleStatus{deletingStatus},
		},
	}
	moduleTemplate := kyma.ModuleTemplate{
		Spec: kyma.ModuleTemplateSpec{
			Data: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "gateway.kyma-project.io/v1alpha1",
					"kind":       "APIGateway",
				},
			},
		},
	}
	crList := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{Object: map[string]interface{}{"status": map[string]interface{}{"state": "Deleting"}}},
		},
	}
	kubeClient := newKubeClient(defaultKyma, moduleTemplate, crList)
	repo := repository.NewModuleInstallationsRepository(kubeClient)

	result, err := repo.ListInstalledModules(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 1)
	module := result[0]
	require.Equal(t, "api-gateway", module.Name)
	require.Equal(t, "Deleting", module.KymaModuleState)
}

func TestModuleInstallationsRepository_ListInstalledModules_SetsInstallationStateForCreateAndDelete(t *testing.T) {
	warningStatus := kyma.ModuleStatus{Name: "api-gateway", State: "Warning"}
	warningStatus.Template.SetName("api-gateway-template")
	warningStatus.Template.SetNamespace("kyma-system")
	defaultKyma := kyma.Kyma{
		Spec: kyma.KymaSpec{
			Modules: []kyma.Module{
				{Name: "api-gateway", CustomResourcePolicy: "CreateAndDelete"},
			},
		},
		Status: kyma.KymaStatus{
			Modules: []kyma.ModuleStatus{warningStatus},
		},
	}
	moduleTemplate := kyma.ModuleTemplate{
		Spec: kyma.ModuleTemplateSpec{
			Data: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "gateway.kyma-project.io/v1alpha1",
					"kind":       "APIGateway",
				},
			},
		},
	}
	crList := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{Object: map[string]interface{}{"status": map[string]interface{}{"state": "Warning"}}},
		},
	}
	kubeClient := newKubeClient(defaultKyma, moduleTemplate, crList)
	repo := repository.NewModuleInstallationsRepository(kubeClient)

	result, err := repo.ListInstalledModules(context.Background())

	require.NoError(t, err)
	module := result[0]
	require.Equal(t, "Warning", module.InstallationState)
}

func TestModuleInstallationsRepository_ListInstalledModules_ModuleCRState_ReturnsStateFromCR(t *testing.T) {
	apiGatewayStatus := kyma.ModuleStatus{Name: "api-gateway", State: "Ready"}
	apiGatewayStatus.Template.SetName("api-gateway-template")
	apiGatewayStatus.Template.SetNamespace("kyma-system")
	defaultKyma := kyma.Kyma{
		Spec: kyma.KymaSpec{
			Modules: []kyma.Module{{Name: "api-gateway"}},
		},
		Status: kyma.KymaStatus{
			Modules: []kyma.ModuleStatus{apiGatewayStatus},
		},
	}
	moduleTemplate := kyma.ModuleTemplate{
		Spec: kyma.ModuleTemplateSpec{
			Data: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "operator.kyma-project.io/v1alpha1",
					"kind":       "APIGateway",
				},
			},
		},
	}
	crList := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{Object: map[string]interface{}{"status": map[string]interface{}{"state": "Warning"}}},
		},
	}
	kubeClient := newKubeClient(defaultKyma, moduleTemplate, crList)
	repo := repository.NewModuleInstallationsRepository(kubeClient)

	result, err := repo.ListInstalledModules(context.Background())

	require.NoError(t, err)
	module := result[0]
	require.Equal(t, "Warning", module.ModuleState)
}

func TestModuleInstallationsRepository_ListInstalledModules_ModuleCRState_UnmanagedModule_FindsTemplateByNameAndVersion(t *testing.T) {
	managed := false
	matchingTemplate := kyma.ModuleTemplate{}
	matchingTemplate.Spec.ModuleName = "api-gateway"
	matchingTemplate.Spec.Version = "3.5.1"
	matchingTemplate.Spec.Data = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1alpha1",
			"kind":       "APIGateway",
		},
	}
	defaultKyma := kyma.Kyma{
		Spec: kyma.KymaSpec{
			Modules: []kyma.Module{{Name: "api-gateway", Managed: &managed}},
		},
		Status: kyma.KymaStatus{
			Modules: []kyma.ModuleStatus{
				{Name: "api-gateway", Version: "3.5.1"},
			},
		},
	}
	crList := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{Object: map[string]interface{}{"status": map[string]interface{}{"state": "Ready"}}},
		},
	}
	kymaClient := &kubefake.KymaClient{
		ReturnDefaultKyma: defaultKyma,
		ReturnModuleTemplateList: kyma.ModuleTemplateList{
			Items: []kyma.ModuleTemplate{matchingTemplate},
		},
	}
	kubeClient := &kubefake.KubeClient{
		TestKymaInterface: kymaClient,
		TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
			ReturnListObjs: crList,
		},
	}
	repo := repository.NewModuleInstallationsRepository(kubeClient)

	result, err := repo.ListInstalledModules(context.Background())

	require.NoError(t, err)
	module := result[0]
	require.Equal(t, "Ready", module.ModuleState)
}

func TestModuleInstallationsRepository_ListInstalledModules_ModuleCRState_ReturnsEmptyOnDiscoveryError(t *testing.T) {
	eventingStatus := kyma.ModuleStatus{Name: "eventing", State: "Ready"}
	eventingStatus.Template.SetName("eventing-template")
	eventingStatus.Template.SetNamespace("kyma-system")
	defaultKyma := kyma.Kyma{
		Spec: kyma.KymaSpec{
			Modules: []kyma.Module{{Name: "eventing"}},
		},
		Status: kyma.KymaStatus{
			Modules: []kyma.ModuleStatus{eventingStatus},
		},
	}
	moduleTemplate := kyma.ModuleTemplate{
		Spec: kyma.ModuleTemplateSpec{
			Data: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "operator.kyma-project.io/v1alpha1",
					"kind":       "Eventing",
				},
			},
		},
	}
	kubeClient := &kubefake.KubeClient{
		TestKymaInterface: &kubefake.KymaClient{
			ReturnDefaultKyma:    defaultKyma,
			ReturnModuleTemplate: moduleTemplate,
		},
		TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
			ReturnErr: fmt.Errorf("failed to discover API resource using discovery client: resource 'Eventing' in group 'operator.kyma-project.io', and version 'v1alpha1' not registered on cluster"),
		},
	}
	repo := repository.NewModuleInstallationsRepository(kubeClient)

	result, err := repo.ListInstalledModules(context.Background())

	require.NoError(t, err)
	module := result[0]
	require.Equal(t, "", module.ModuleState)
}

func TestModuleInstallationsRepository_ListInstalledModules_InstallationState_ManagedModuleUsesManagerState(t *testing.T) {
	managed := true
	apiGatewayStatus := kyma.ModuleStatus{Name: "api-gateway", State: "Warning"}
	apiGatewayStatus.Template.SetName("api-gateway-template")
	apiGatewayStatus.Template.SetNamespace("kyma-system")
	defaultKyma := kyma.Kyma{
		Spec: kyma.KymaSpec{
			Modules: []kyma.Module{{Name: "api-gateway", Managed: &managed}},
		},
		Status: kyma.KymaStatus{
			Modules: []kyma.ModuleStatus{apiGatewayStatus},
		},
	}
	moduleTemplate := kyma.ModuleTemplate{
		Spec: kyma.ModuleTemplateSpec{
			Manager: &kyma.Manager{
				GroupVersionKind: managerGVK("apps", "v1", "Deployment"),
				Name:             "api-gateway-manager",
				Namespace:        "kyma-system",
			},
		},
	}
	managerObj := unstructured.Unstructured{
		Object: map[string]interface{}{
			"status": map[string]interface{}{
				"readyReplicas": int64(1),
			},
			"spec": map[string]interface{}{
				"replicas": int64(1),
			},
		},
	}
	kubeClient := &kubefake.KubeClient{
		TestKymaInterface: &kubefake.KymaClient{
			ReturnDefaultKyma:    defaultKyma,
			ReturnModuleTemplate: moduleTemplate,
		},
		TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
			ReturnGetObj: managerObj,
		},
	}
	repo := repository.NewModuleInstallationsRepository(kubeClient)

	result, err := repo.ListInstalledModules(context.Background())

	require.NoError(t, err)
	module := result[0]
	require.Equal(t, "Ready", module.InstallationState)
}

func managerGVK(group, version, kind string) metav1.GroupVersionKind {
	return metav1.GroupVersionKind{Group: group, Version: version, Kind: kind}
}

func TestModuleInstallationsRepository_ListInstalledCommunityModules_ReturnsModuleFromCommunityTemplate(t *testing.T) {
	communityTemplate := kyma.ModuleTemplate{}
	communityTemplate.SetName("docker-registry")
	communityTemplate.SetNamespace("default")
	communityTemplate.Spec.ModuleName = "docker-registry"
	communityTemplate.Spec.Version = "0.10.0"
	communityTemplate.Spec.Manager = &kyma.Manager{
		GroupVersionKind: managerGVK("apps", "v1", "Deployment"),
		Name:             "docker-registry-manager",
		Namespace:        "default",
	}
	kubeClient := &kubefake.KubeClient{
		TestKymaInterface: &kubefake.KymaClient{
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{communityTemplate},
			},
		},
		TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{},
	}
	repo := repository.NewModuleInstallationsRepository(kubeClient)

	result, err := repo.ListInstalledCommunityModules(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 1)
	module := result[0]
	require.Equal(t, "docker-registry", module.Name)
	require.Equal(t, "default", module.Namespace)
	require.Equal(t, "0.10.0", module.Version)
}

func TestModuleInstallationsRepository_ListInstalledCommunityModules_SetsInstallationStateFromManager(t *testing.T) {
	communityTemplate := kyma.ModuleTemplate{}
	communityTemplate.SetName("docker-registry")
	communityTemplate.SetNamespace("default")
	communityTemplate.Spec.ModuleName = "docker-registry"
	communityTemplate.Spec.Version = "0.10.0"
	communityTemplate.Spec.Manager = &kyma.Manager{
		GroupVersionKind: managerGVK("apps", "v1", "Deployment"),
		Name:             "docker-registry-manager",
		Namespace:        "default",
	}
	managerObj := unstructured.Unstructured{
		Object: map[string]interface{}{
			"status": map[string]interface{}{
				"readyReplicas": int64(1),
			},
			"spec": map[string]interface{}{
				"replicas": int64(1),
			},
		},
	}
	kubeClient := &kubefake.KubeClient{
		TestKymaInterface: &kubefake.KymaClient{
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{communityTemplate},
			},
		},
		TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
			ReturnGetObj: managerObj,
		},
	}
	repo := repository.NewModuleInstallationsRepository(kubeClient)

	result, err := repo.ListInstalledCommunityModules(context.Background())

	require.NoError(t, err)
	module := result[0]
	require.Equal(t, "Ready", module.InstallationState)
}

func TestModuleInstallationsRepository_ListInstalledCommunityModules_SkipsTemplateWithoutManager(t *testing.T) {
	templateWithoutManager := kyma.ModuleTemplate{}
	templateWithoutManager.SetName("docker-registry")
	templateWithoutManager.SetNamespace("default")
	templateWithoutManager.Spec.ModuleName = "docker-registry"
	templateWithoutManager.Spec.Version = "0.10.0"
	// no Spec.Manager
	kubeClient := &kubefake.KubeClient{
		TestKymaInterface: &kubefake.KymaClient{
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{templateWithoutManager},
			},
		},
		TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{},
	}
	repo := repository.NewModuleInstallationsRepository(kubeClient)

	result, err := repo.ListInstalledCommunityModules(context.Background())

	require.NoError(t, err)
	require.Empty(t, result)
}

func TestModuleInstallationsRepository_ListInstalledCommunityModules_SetsModuleStateFromCR(t *testing.T) {
	communityTemplate := kyma.ModuleTemplate{}
	communityTemplate.SetName("docker-registry")
	communityTemplate.SetNamespace("default")
	communityTemplate.Spec.ModuleName = "docker-registry"
	communityTemplate.Spec.Version = "0.10.0"
	communityTemplate.Spec.Manager = &kyma.Manager{
		GroupVersionKind: managerGVK("apps", "v1", "Deployment"),
		Name:             "docker-registry-manager",
		Namespace:        "default",
	}
	communityTemplate.Spec.Data = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1alpha1",
			"kind":       "DockerRegistry",
		},
	}
	crList := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{Object: map[string]interface{}{"status": map[string]interface{}{"state": "Ready"}}},
		},
	}
	kubeClient := &kubefake.KubeClient{
		TestKymaInterface: &kubefake.KymaClient{
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{communityTemplate},
			},
		},
		TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
			ReturnListObjs: crList,
		},
	}
	repo := repository.NewModuleInstallationsRepository(kubeClient)

	result, err := repo.ListInstalledCommunityModules(context.Background())

	require.NoError(t, err)
	module := result[0]
	require.Equal(t, "Ready", module.ModuleState)
}
