package repository

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ModuleCRStateRepository interface {
	GetModuleCRState(ctx context.Context, module entities.ModuleInstallation) (string, error)
}

type moduleCRStateRepository struct {
	kubeClient kube.Client
}

func NewModuleCRStateRepository(kubeClient kube.Client) ModuleCRStateRepository {
	return &moduleCRStateRepository{kubeClient: kubeClient}
}

func (r *moduleCRStateRepository) GetModuleCRState(ctx context.Context, module entities.ModuleInstallation) (string, error) {
	moduleTemplate, err := r.findModuleTemplate(ctx, module)
	if err != nil {
		return "", err
	}
	if moduleTemplate == nil {
		return "", nil
	}

	data := moduleTemplate.Spec.Data
	if len(data.Object) == 0 {
		return "", nil
	}

	crList, err := r.kubeClient.RootlessDynamic().List(ctx, &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": data.GetAPIVersion(),
			"kind":       data.GetKind(),
		},
	}, &rootlessdynamic.ListOptions{AllNamespaces: true})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}

	return highestStateFromList(crList.Items), nil
}

func (r *moduleCRStateRepository) findModuleTemplate(ctx context.Context, module entities.ModuleInstallation) (*kyma.ModuleTemplate, error) {
	if module.TemplateName != "" {
		mt, err := r.kubeClient.Kyma().GetModuleTemplate(ctx, module.TemplateNamespace, module.TemplateName)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return nil, nil
			}
			return nil, err
		}
		return mt, nil
	}

	templates, err := r.kubeClient.Kyma().ListModuleTemplate(ctx)
	if err != nil {
		return nil, err
	}
	for i, mt := range templates.Items {
		if mt.Spec.ModuleName == module.Name && mt.Spec.Version == module.Version {
			return &templates.Items[i], nil
		}
	}
	return nil, nil
}

func highestStateFromList(items []unstructured.Unstructured) string {
	state := ""
	for _, item := range items {
		statusRaw, ok := item.Object["status"]
		if !ok || statusRaw == nil {
			continue
		}
		status := statusRaw.(map[string]any)
		if crState, ok := status["state"].(string); ok {
			state = highestState(state, crState)
		}
	}
	return state
}

var statesPrecedence = []string{"Ready", "Processing", "Deleting", "Error", "Warning"}

func highestState(a, b string) string {
	for _, s := range statesPrecedence {
		if a == s {
			return a
		}
		if b == s {
			return b
		}
	}
	if a == "" {
		return b
	}
	return a
}

