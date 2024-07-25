package rootlessdynamic

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"strings"
)

type Interface interface {
	ApplyMany(context.Context, []unstructured.Unstructured) error
}

type client struct {
	dynamic dynamic.Interface
}

func NewClient(dynamic dynamic.Interface) Interface {
	return &client{
		dynamic: dynamic,
	}
}

func (c *client) ApplyMany(ctx context.Context, objs []unstructured.Unstructured) error {
	for _, resource := range objs {
		group, version := groupVersion(resource.GetAPIVersion())
		gvr := schema.GroupVersionResource{
			Group:    group,
			Version:  version,
			Resource: kindToResource(resource.GetKind()),
		}

		data, err := runtime.Encode(unstructured.UnstructuredJSONScheme, &resource)

		if kind := resource.GetKind(); kind == "CustomResourceDefinition" || kind == "ClusterRole" || kind == "ClusterRoleBinding" {
			_, err = c.dynamic.Resource(gvr).Patch(ctx, resource.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
				FieldManager: "cli",
			})
			if err != nil {
				return err
			}
		} else {
			_, err = c.dynamic.Resource(gvr).Namespace(resource.GetNamespace()).Patch(ctx, resource.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
				FieldManager: "cli",
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func groupVersion(version string) (string, string) {
	split := strings.Split(version, "/")
	if len(split) > 1 {

		return split[0], split[1]
	}
	return "", split[0]
}

func kindToResource(kind string) string {
	return makePlural(strings.ToLower(kind))
}

func makePlural(singular string) string {
	if strings.HasSuffix(singular, "y") {
		return strings.TrimSuffix(singular, "y") + "ies"
	}
	if strings.HasSuffix(singular, "fe") {
		return strings.TrimSuffix(singular, "fe") + "ves"
	}
	if strings.HasSuffix(singular, "s") {
		return singular + "es"
	}
	if strings.HasSuffix(singular, "ss") {
		return singular + "es"
	}
	if strings.HasSuffix(singular, "ch") {
		return singular + "es"
	}
	if strings.HasSuffix(singular, "sh") {
		return singular + "es"
	}
	if strings.HasSuffix(singular, "x") {
		return singular + "es"
	}
	return singular + "s"
}
