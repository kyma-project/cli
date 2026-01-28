package fake

import (
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ExternalParams struct {
	TemplateName string
	ModuleName   string
	Version      string
}

func ExternalModuleTemplate(params *ExternalParams) *entities.ExternalModuleTemplate {
	defaults := defaultExternalParams()

	if params == nil {
		params = defaults
	}

	rawModuleTemplate := kyma.ModuleTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.kyma-project.io/v1beta2",
			Kind:       "ModuleTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: firstNonEmpty(params.TemplateName, defaults.TemplateName),
		},
		Spec: kyma.ModuleTemplateSpec{
			ModuleName: firstNonEmpty(params.ModuleName, defaults.ModuleName),
			Version:    firstNonEmpty(params.Version, defaults.Version),
			Data: unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "operator.kyma-project.io/v1beta2",
					"kind":       "CommunityModule",
					"metadata": map[string]any{
						"name": "community-module-test",
					},
				},
			},
		},
	}

	return entities.NewExternalModuleTemplateFromRaw(&rawModuleTemplate)
}

func defaultExternalParams() *ExternalParams {
	return &ExternalParams{
		TemplateName: "sample-external-community-template-0.0.1",
		ModuleName:   "sample-external-community-module",
		Version:      "0.0.1",
	}
}
