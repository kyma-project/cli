package templates

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/parameters"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func setExtraValues(u *unstructured.Unstructured, extraValues []parameters.Value) clierror.Error {
	for _, extraValue := range extraValues {
		value := extraValue.GetValue()
		if value == nil {
			// value is not set and has no default value
			continue
		}

		fields := strings.Split(
			// remove optional dot at the beginning of the path
			strings.TrimPrefix(extraValue.GetPath(), "."),
			".",
		)

		err := unstructured.SetNestedField(u.Object, value, fields...)
		if err != nil {
			return clierror.Wrap(err, clierror.New(
				fmt.Sprintf("failed to set value %v for path %s", value, extraValue.GetPath()),
			))
		}
	}

	return nil
}

func commonResourceFlags(resourceScope types.Scope) []types.CustomFlag {
	params := []types.CustomFlag{
		{
			Name:        "name",
			Type:        types.StringCustomFlagType,
			Description: "name of the resource",
			Path:        ".metadata.name",
			Required:    true,
		},
	}
	if resourceScope == types.NamespaceScope {
		params = append(params, types.CustomFlag{
			Name:         "namespace",
			Type:         types.StringCustomFlagType,
			Description:  "resource namespace",
			Path:         ".metadata.namespace",
			DefaultValue: "default",
		})
	}

	return params
}

func getResourceName(scope types.Scope, u *unstructured.Unstructured) string {
	if scope == types.NamespaceScope {
		return fmt.Sprintf("%s/%s", u.GetNamespace(), u.GetName())
	}

	return u.GetName()
}
