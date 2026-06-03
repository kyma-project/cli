package repository

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type moduleInstallationStateFetcher struct {
	kubeClient kube.Client
}

func (r *moduleInstallationStateFetcher) GetInstallationState(ctx context.Context, module entities.ModuleInstallation) (string, error) {
	moduleTemplate, err := r.kubeClient.Kyma().GetModuleTemplate(ctx, module.TemplateNamespace, module.TemplateName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return "", nil
		}
		return "", errors.Wrapf(err, "failed to get ModuleTemplate %s/%s", module.TemplateNamespace, module.TemplateName)
	}

	return getResourceState(ctx, r.kubeClient, moduleTemplate.Spec.Manager)
}

func getResourceState(ctx context.Context, client kube.Client, manager *kyma.Manager) (string, error) {
	if manager == nil {
		return "", nil
	}
	namespace := "kyma-system"
	if manager.Namespace != "" {
		namespace = manager.Namespace
	}

	apiVersion := manager.Group + "/" + manager.Version
	unstruct := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       manager.Kind,
			"metadata": map[string]interface{}{
				"name":      manager.Name,
				"namespace": namespace,
			},
		},
	}

	result, err := client.RootlessDynamic().Get(ctx, &unstruct)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}

	return extractStateFromObject(result), nil
}
