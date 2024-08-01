package rootlessdynamic

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyma-project/cli.v3/internal/clierror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"strings"
)

type Interface interface {
	Apply(context.Context, unstructured.Unstructured) clierror.Error
	ApplyMany(context.Context, []unstructured.Unstructured) clierror.Error
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
		discovery: discovery,
	}, nil
}

func (c *client) Apply(ctx context.Context, resource unstructured.Unstructured) clierror.Error {
	group, version := groupVersion(resource.GetAPIVersion())
	apiResource, err := c.discoverAPIResource(group, version, resource.GetKind())
	if err != nil {
		return clierror.Wrap(err, clierror.New("Failed to discover API resource using discovery client"))
	}

	gvr := &schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: apiResource.Name,
	}
	fmt.Println(resource.GetKind())
	data, err := runtime.Encode(unstructured.UnstructuredJSONScheme, &resource)
	if err != nil {
		return clierror.Wrap(err, clierror.New("Failed to encode resource"))
	}

	if apiResource.Namespaced {
		_, err = c.dynamic.Resource(*gvr).Namespace("kyma-system").Patch(ctx, resource.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
			FieldManager: "cli",
		})
		if err != nil {
			return clierror.Wrap(err, clierror.New("Failed to apply namespaced resource"))
		}
	} else {
		_, err = c.dynamic.Resource(*gvr).Patch(ctx, resource.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
			FieldManager: "cli",
		})
		if err != nil {
			return clierror.Wrap(err, clierror.New("Failed to apply cluster-scoped resource"))
		}
	}
	return nil
}

// TODO: Add a script to test applying default resources
func (c *client) ApplyMany(ctx context.Context, objs []unstructured.Unstructured) clierror.Error {
	for _, resource := range objs {
		err := c.Apply(ctx, resource)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *client) discoverAPIResource(group, version, kind string) (*metav1.APIResource, error) {
	groupVersion := schema.GroupVersion{Group: group, Version: version}

	apiResourceList, err := c.discovery.ServerResourcesForGroupVersion(groupVersion.String())
	if err != nil {
		return nil, err
	}

	for _, apiResource := range apiResourceList.APIResources {
		if apiResource.Kind == kind {
			return &apiResource, nil
		}
	}
	return nil, errors.New("Resource " + kind + " in group " + group + " and version " + version + " not registered on cluster")
}

func groupVersion(version string) (string, string) {
	split := strings.Split(version, "/")
	if len(split) > 1 {

		return split[0], split[1]
	}
	return "", split[0]
}
