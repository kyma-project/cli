package rootlessdynamic

import (
	"context"
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"strings"
)

type Interface interface {
	ApplyMany(context.Context, []unstructured.Unstructured) error
}

type client struct {
	dynamic   dynamic.Interface
	discovery discovery.DiscoveryInterface
}

func NewClient(dynamic dynamic.Interface, restConfig *rest.Config) (Interface, error) {
	discovery, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &client{
		dynamic:   dynamic,
		discovery: memory.NewMemCacheClient(discovery),
	}, nil
}

// TODO: Add a script to test applying default resources
func (c *client) ApplyMany(ctx context.Context, objs []unstructured.Unstructured) error {
	for _, resource := range objs {
		gvr, err := c.discoverGVR(resource.GetAPIVersion(), resource.GetKind())
		if err != nil {
			return err
		}

		data, err := runtime.Encode(unstructured.UnstructuredJSONScheme, &resource)
		if err != nil {
			return err
		}

		if kind := resource.GetKind(); kind == "CustomResourceDefinition" || kind == "ClusterRole" || kind == "ClusterRoleBinding" {
			_, err = c.dynamic.Resource(*gvr).Patch(ctx, resource.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
				FieldManager: "cli",
			})
			if err != nil {
				return err
			}
		} else {
			_, err = c.dynamic.Resource(*gvr).Namespace("kyma-system").Patch(ctx, resource.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
				FieldManager: "cli",
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *client) discoverGVR(apiVersion, kind string) (*schema.GroupVersionResource, error) {
	group, version := groupVersion(apiVersion)
	groupVersion := schema.GroupVersion{Group: group, Version: version}

	apiResourceList, err := c.discovery.ServerResourcesForGroupVersion(groupVersion.String())
	if err != nil {
		return nil, err
	}

	for _, apiResource := range apiResourceList.APIResources {
		if apiResource.Kind == kind {
			return &schema.GroupVersionResource{
				Group:    group,
				Version:  version,
				Resource: apiResource.Name,
			}, nil

		}
	}
	return nil, errors.New("Resource " + kind + " in apiVersion " + apiVersion + " not registered on cluster")
}

func groupVersion(version string) (string, string) {
	split := strings.Split(version, "/")
	if len(split) > 1 {

		return split[0], split[1]
	}
	return "", split[0]
}
