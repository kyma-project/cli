package templates

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	resourceNameFlag = types.CustomFlag{
		Name:        "name",
		Type:        types.StringCustomFlagType,
		Description: "name of the resource",
		Path:        ".metadata.name",
		Required:    true,
	}

	resourceNamespaceFlag = types.CustomFlag{
		Name:         "namespace",
		Type:         types.StringCustomFlagType,
		Description:  "resource namespace",
		Path:         ".metadata.namespace",
		DefaultValue: "default",
	}
)

func getResourceName(scope types.Scope, u *unstructured.Unstructured) string {
	if scope == types.NamespaceScope {
		return fmt.Sprintf("%s/%s", u.GetNamespace(), u.GetName())
	}

	return u.GetName()
}
