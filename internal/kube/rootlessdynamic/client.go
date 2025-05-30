package rootlessdynamic

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

type applyFunc func(context.Context, dynamic.ResourceInterface, *unstructured.Unstructured, bool) error

type Interface interface {
	Get(context.Context, *unstructured.Unstructured) (*unstructured.Unstructured, error)
	List(context.Context, *unstructured.Unstructured, *ListOptions) (*unstructured.UnstructuredList, error)
	Apply(context.Context, *unstructured.Unstructured, bool) error
	ApplyMany(context.Context, []unstructured.Unstructured) error
	Remove(context.Context, *unstructured.Unstructured, bool) error
	RemoveMany(context.Context, []unstructured.Unstructured) error
	WatchSingleResource(context.Context, *unstructured.Unstructured) (watch.Interface, error)
}

type client struct {
	dynamic   dynamic.Interface
	discovery discovery.DiscoveryInterface

	// for testing purposes
	applyFunc applyFunc
}

func NewClient(dynamic dynamic.Interface, discovery discovery.DiscoveryInterface) Interface {
	return NewClientWithApplyFunc(dynamic, discovery, applyResource)
}

func NewClientWithApplyFunc(dynamic dynamic.Interface, discovery discovery.DiscoveryInterface, applyFunc applyFunc) Interface {
	return &client{
		dynamic:   dynamic,
		discovery: discovery,
		applyFunc: applyFunc,
	}
}

func (c *client) Get(ctx context.Context, resource *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	group, version := groupVersion(resource.GetAPIVersion())
	apiResource, err := c.discoverAPIResource(group, version, resource.GetKind())
	if err != nil {
		return nil, fmt.Errorf("failed to discover API resource using discovery client: %w", err)
	}

	gvr := &schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: apiResource.Name,
	}

	if apiResource.Namespaced {
		return c.dynamic.Resource(*gvr).Namespace(resource.GetNamespace()).Get(ctx, resource.GetName(), metav1.GetOptions{})
	}
	return c.dynamic.Resource(*gvr).Get(ctx, resource.GetName(), metav1.GetOptions{})
}

type ListOptions struct {
	AllNamespaces bool
	FieldSelector string
}

func (c *client) List(ctx context.Context, resource *unstructured.Unstructured, opts *ListOptions) (*unstructured.UnstructuredList, error) {
	group, version := groupVersion(resource.GetAPIVersion())
	apiResource, err := c.discoverAPIResource(group, version, resource.GetKind())
	if err != nil {
		return nil, fmt.Errorf("failed to discover API resource using discovery client: %w", err)
	}

	gvr := &schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: apiResource.Name,
	}

	if apiResource.Namespaced && !opts.AllNamespaces && resource.GetNamespace() != "" {
		return c.dynamic.Resource(*gvr).Namespace(getResourceNamespace(resource)).List(ctx, metav1.ListOptions{
			FieldSelector: opts.FieldSelector,
		})
	}

	return c.dynamic.Resource(*gvr).List(ctx, metav1.ListOptions{
		FieldSelector: opts.FieldSelector,
	})
}

func (c *client) Apply(ctx context.Context, resource *unstructured.Unstructured, druRun bool) error {
	group, version := groupVersion(resource.GetAPIVersion())
	apiResource, err := c.discoverAPIResource(group, version, resource.GetKind())
	if err != nil {
		return fmt.Errorf("failed to discover API resource using discovery client: %w", err)
	}

	gvr := &schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: apiResource.Name,
	}

	if apiResource.Namespaced {
		err = c.applyFunc(ctx, c.dynamic.Resource(*gvr).Namespace(getResourceNamespace(resource)), resource, druRun)
		if err != nil {
			return fmt.Errorf("failed to apply namespaced resource: %w", err)
		}
	} else {
		err = c.applyFunc(ctx, c.dynamic.Resource(*gvr), resource, druRun)
		if err != nil {
			return fmt.Errorf("failed to apply cluster-scoped resource: %w", err)
		}
	}
	return nil
}

func (c *client) ApplyMany(ctx context.Context, objs []unstructured.Unstructured) error {
	for _, resource := range objs {
		// TODO: add dryRun support
		err := c.Apply(ctx, &resource, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *client) Remove(ctx context.Context, resource *unstructured.Unstructured, dryRun bool) error {
	group, version := groupVersion(resource.GetAPIVersion())
	apiResource, err := c.discoverAPIResource(group, version, resource.GetKind())
	if err != nil {
		return fmt.Errorf("failed to discover API resource using discovery client: %w", err)
	}

	gvr := &schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: apiResource.Name,
	}

	dryRunOpts := []string{}
	if dryRun {
		dryRunOpts = append(dryRunOpts, "All")
	}

	if apiResource.Namespaced {
		err = c.dynamic.Resource(*gvr).Namespace(getResourceNamespace(resource)).Delete(ctx, resource.GetName(), metav1.DeleteOptions{
			DryRun: dryRunOpts,
		})
		if err != nil {
			return fmt.Errorf("failed to delete namespaced resource %w", err)
		}
	} else {
		err = c.dynamic.Resource(*gvr).Delete(ctx, resource.GetName(), metav1.DeleteOptions{
			DryRun: dryRunOpts,
		})
		if err != nil {
			return fmt.Errorf("failed to delete cluster-scoped resource %w", err)
		}
	}
	return nil
}

func (c *client) RemoveMany(ctx context.Context, objs []unstructured.Unstructured) error {
	for _, resource := range objs {
		// TODO: add dryRun support
		err := c.Remove(ctx, &resource, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *client) WatchSingleResource(ctx context.Context, resource *unstructured.Unstructured) (watch.Interface, error) {
	group, version := groupVersion(resource.GetAPIVersion())
	apiResource, err := c.discoverAPIResource(group, version, resource.GetKind())
	if err != nil {
		return nil, fmt.Errorf("failed to discover API resource using discovery client: %w", err)
	}

	fieldSelector := fmt.Sprintf("metadata.name=%s", resource.GetName())
	gvr := &schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: apiResource.Name,
	}

	if apiResource.Namespaced {
		return c.dynamic.Resource(*gvr).Namespace(getResourceNamespace(resource)).Watch(ctx, metav1.ListOptions{
			FieldSelector: fieldSelector,
		})
	}

	return c.dynamic.Resource(*gvr).Watch(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
}

// applyResource creates or updates given object
func applyResource(ctx context.Context, resourceInterface dynamic.ResourceInterface, resource *unstructured.Unstructured, dryRun bool) error {
	dryRunOpts := []string{}
	if dryRun {
		dryRunOpts = append(dryRunOpts, "All")
	}

	// this function can't be tested because of dynamic.FakeDynamicClient limitations
	_, err := resourceInterface.Apply(ctx, resource.GetName(), resource, metav1.ApplyOptions{
		FieldManager: "cli",
		Force:        true,
		DryRun:       dryRunOpts,
	})

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

// returns resource namespace or kyma-system if empty
func getResourceNamespace(resource *unstructured.Unstructured) string {
	if resource.GetNamespace() != "" {
		return resource.GetNamespace()
	}

	return "kyma-system"
}
