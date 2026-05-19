package repository_test

import (
	"context"
	"fmt"
	"testing"

	kubefake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestModuleCRStateRepository_GetModuleCRState_ReturnsStateFromCR(t *testing.T) {
	moduleTemplate := kyma.ModuleTemplate{}
	moduleTemplate.Spec.Data = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1alpha1",
			"kind":       "APIGateway",
		},
	}

	crList := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"state": "Warning",
					},
				},
			},
		},
	}

	kubeClient := &kubefake.KubeClient{
		TestKymaInterface: &kubefake.KymaClient{
			ReturnModuleTemplate: moduleTemplate,
		},
		TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
			ReturnListObjs: crList,
		},
	}

	module := entities.ModuleInstallation{
		TemplateName:      "api-gateway-template",
		TemplateNamespace: "kyma-system",
	}
	repo := repository.NewModuleCRStateRepository(kubeClient)

	state, err := repo.GetModuleCRState(context.Background(), module)

	require.NoError(t, err)
	require.Equal(t, "Warning", state)
}

func TestModuleCRStateRepository_GetModuleCRState_UnmanagedModule_FindsTemplateByNameAndVersion(t *testing.T) {
	matchingTemplate := kyma.ModuleTemplate{}
	matchingTemplate.Spec.ModuleName = "api-gateway"
	matchingTemplate.Spec.Version = "3.5.1"
	matchingTemplate.Spec.Data = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1alpha1",
			"kind":       "APIGateway",
		},
	}

	crList := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"state": "Ready",
					},
				},
			},
		},
	}

	kubeClient := &kubefake.KubeClient{
		TestKymaInterface: &kubefake.KymaClient{
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{matchingTemplate},
			},
		},
		TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
			ReturnListObjs: crList,
		},
	}

	managed := false
	module := entities.ModuleInstallation{
		Name:    "api-gateway",
		Version: "3.5.1",
		Managed: &managed,
	}
	repo := repository.NewModuleCRStateRepository(kubeClient)

	state, err := repo.GetModuleCRState(context.Background(), module)

	require.NoError(t, err)
	require.Equal(t, "Ready", state)
}

func TestModuleCRStateRepository_GetModuleCRState_ReturnsEmptyOnDiscoveryError(t *testing.T) {
	moduleTemplate := kyma.ModuleTemplate{}
	moduleTemplate.Spec.Data = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1alpha1",
			"kind":       "Eventing",
		},
	}

	kubeClient := &kubefake.KubeClient{
		TestKymaInterface: &kubefake.KymaClient{
			ReturnModuleTemplate: moduleTemplate,
		},
		TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
			ReturnErr: fmt.Errorf("failed to discover API resource using discovery client: resource 'Eventing' in group 'operator.kyma-project.io', and version 'v1alpha1' not registered on cluster"),
		},
	}

	module := entities.ModuleInstallation{
		TemplateName:      "eventing-template",
		TemplateNamespace: "kyma-system",
	}
	repo := repository.NewModuleCRStateRepository(kubeClient)

	state, err := repo.GetModuleCRState(context.Background(), module)

	require.NoError(t, err)
	require.Equal(t, "", state)
}
