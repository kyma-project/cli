package repository

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ModuleInstallationStateRepository interface {
	GetInstallationState(ctx context.Context, status kyma.ModuleStatus, spec kyma.Module) (string, error)
}

type moduleInstallationStateRepository struct {
	kubeClient kube.Client
}

func NewModuleInstallationStateRepository(kubeClient kube.Client) ModuleInstallationStateRepository {
	return &moduleInstallationStateRepository{kubeClient: kubeClient}
}

func (r *moduleInstallationStateRepository) GetInstallationState(ctx context.Context, status kyma.ModuleStatus, spec kyma.Module) (string, error) {
	if spec.CustomResourcePolicy == "CreateAndDelete" {
		return status.State, nil
	}

	if spec.Managed != nil && !*spec.Managed {
		return status.State, nil
	}

	moduleTemplate, err := r.kubeClient.Kyma().GetModuleTemplate(ctx, status.Template.GetNamespace(), status.Template.GetName())
	if err != nil {
		if apierrors.IsNotFound(err) {
			return "", nil
		}
		return "", errors.Wrapf(err, "failed to get ModuleTemplate %s/%s", status.Template.GetNamespace(), status.Template.GetName())
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

	statusRaw, ok := result.Object["status"]
	if !ok || statusRaw == nil {
		return "", nil
	}
	status := statusRaw.(map[string]any)
	if state, ok := status["state"]; ok {
		return state.(string), nil
	}

	if conditions, ok := status["conditions"]; ok {
		return getStateFromConditions(conditions.([]any)), nil
	}

	if readyReplicas, ok := status["readyReplicas"]; ok {
		spec := result.Object["spec"].(map[string]any)
		if wantedReplicas, ok := spec["replicas"]; ok {
			return resolveStateFromReplicas(readyReplicas.(int64), wantedReplicas.(int64)), nil
		}
	}

	return "", nil
}

func getStateFromConditions(conditions []interface{}) string {
	for _, condition := range conditions {
		c := condition.(map[string]interface{})
		if c["status"] != "True" {
			continue
		}
		switch c["type"].(string) {
		case "Available":
			return "Ready"
		case "Processing", "Error", "Warning":
			return c["type"].(string)
		}
	}
	return ""
}

func resolveStateFromReplicas(ready, wanted int64) string {
	if ready == wanted {
		return "Ready"
	}
	if ready < wanted {
		return "Processing"
	}
	return "Deleting"
}
