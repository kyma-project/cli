package rootlessdynamic

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	apimachinery_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type Interface interface {
	Apply(context.Context, *unstructured.Unstructured) clierror.Error
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

func (c *client) Apply(ctx context.Context, resource *unstructured.Unstructured) clierror.Error {
	group, version := groupVersion(resource.GetAPIVersion())
	apiResource, err := c.discoverAPIResource(group, version, resource.GetKind())
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to discover API resource using discovery client"))
	}

	gvr := &schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: apiResource.Name,
	}

	if apiResource.Namespaced {
		// we should not expect here for all resources to be installed in the kyma-system namespace. passed resources should be defaulted and validated out of the Apply func
		err = applyResource(ctx, c.dynamic.Resource(*gvr).Namespace("kyma-system"), resource)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to apply namespaced resource"))
		}
	} else {
		err = applyResource(ctx, c.dynamic.Resource(*gvr), resource)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to apply cluster-scoped resource"))
		}
	}
	return nil
}

func (c *client) ApplyMany(ctx context.Context, objs []unstructured.Unstructured) clierror.Error {
	for _, resource := range objs {
		err := c.Apply(ctx, &resource)
		if err != nil {
			return err
		}
	}
	return nil
}

// applyResource creates or updates given object
func applyResource(ctx context.Context, resourceInterface dynamic.ResourceInterface, resource *unstructured.Unstructured) error {
	_, err := resourceInterface.Create(ctx, resource, metav1.CreateOptions{
		FieldManager: "cli",
	})
	if apimachinery_errors.IsAlreadyExists(err) {
		_, err = resourceInterface.Update(ctx, resource, metav1.UpdateOptions{
			FieldManager: "cli",
		})
	}

	return err
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
	return nil, fmt.Errorf("resource '%s' in group '%s', and version '%s' not registered on cluster", kind, group, version)
}

func groupVersion(version string) (string, string) {
	split := strings.Split(version, "/")
	if len(split) > 1 {

		return split[0], split[1]
	}
	return "", split[0]
}
