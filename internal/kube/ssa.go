package kube

import (
	"context"
	"errors"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/cli-runtime/pkg/resource"

	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var ErrClientObjectConversionFailed = errors.New("ctrlClient object conversion failed")

type SSA interface {
	Run(context.Context, []*resource.Info) error
}

type ConcurrentDefaultSSA struct {
	clnt      ctrlClient.Client
	owner     ctrlClient.FieldOwner
	versioner runtime.GroupVersioner
	force     bool
}

func ConcurrentSSA(clnt ctrlClient.Client, owner ctrlClient.FieldOwner, force bool) *ConcurrentDefaultSSA {
	return &ConcurrentDefaultSSA{
		clnt: clnt, owner: owner, force: force,
		versioner: schema.GroupVersions(clnt.Scheme().PrioritizedVersionsAllGroups()),
	}
}

func (c *ConcurrentDefaultSSA) Run(ctx context.Context, resources []*resource.Info) error {
	ssaStart := time.Now()
	logger := log.FromContext(ctx, "owner", c.owner)
	logger.V(2).Info("ServerSideApply", "resources", len(resources))

	// The Runtime Complexity of this Branch is N as only ServerSideApplier Apply is required
	results := make(chan error, len(resources))
	for i := range resources {
		i := i
		go c.serverSideApply(ctx, resources[i], results)
	}

	var errsFromApply []error
	for i := 0; i < len(resources); i++ {
		if err := <-results; err != nil {
			errsFromApply = append(errsFromApply, err)
		}
	}

	ssaFinish := time.Since(ssaStart)

	if errsFromApply != nil {
		return fmt.Errorf("ServerSideApply failed (after %s): %w", ssaFinish, errors.Join(errsFromApply...))
	}
	logger.V(2).Info("ServerSideApply finished", "time", ssaFinish)
	return nil
}

func (c *ConcurrentDefaultSSA) serverSideApply(
	ctx context.Context,
	resource *resource.Info,
	results chan error,
) {
	start := time.Now()
	logger := log.FromContext(ctx, "owner", c.owner)

	// this converts unstructured to typed objects if possible, leveraging native APIs
	resource.Object = c.convertUnstructuredToTyped(resource.Object, resource.Mapping)

	logger.V(3).Info(
		fmt.Sprintf("apply %s", resource.ObjectName()),
	)

	results <- c.serverSideApplyResourceInfo(ctx, resource)

	logger.V(3).Info(
		fmt.Sprintf("apply %s finished", resource.ObjectName()),
		"time", time.Since(start),
	)
}

func (c *ConcurrentDefaultSSA) serverSideApplyResourceInfo(ctx context.Context, info *resource.Info) error {
	obj, isTyped := info.Object.(ctrlClient.Object)
	if !isTyped {
		return fmt.Errorf(
			"%s is not a valid ctrlClient-go object: %w", info.ObjectName(), ErrClientObjectConversionFailed,
		)
	}

	obj.SetManagedFields(nil)
	obj.SetResourceVersion("")

	opts := []ctrlClient.PatchOption{c.owner}
	if c.force {
		opts = append(opts, ctrlClient.ForceOwnership)
	}

	err := c.clnt.Patch(ctx, obj, ctrlClient.Apply, opts...)
	if err != nil {
		return fmt.Errorf(
			"patch for %s failed: %w", info.ObjectName(), err,
		)
	}

	return nil
}

// convertWithMapper converts the given object with the optional provided
// RESTMapping. If no mapping is provided, the default schema versioner is used.
//

func (c *ConcurrentDefaultSSA) convertUnstructuredToTyped(
	obj runtime.Object, mapping *meta.RESTMapping,
) runtime.Object {
	gv := c.versioner
	if mapping != nil {
		gv = mapping.GroupVersionKind.GroupVersion()
	}
	if obj, err := c.clnt.Scheme().UnsafeConvertToVersion(obj, gv); err == nil {
		return obj
	}
	return obj
}
