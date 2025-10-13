package authorization

import (
	"context"
	"fmt"
	"os"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kubeconfig"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ResourceApplier struct {
	kubeClient kube.Client
}

func NewResourceApplier(kubeClient kube.Client) *ResourceApplier {
	return &ResourceApplier{
		kubeClient: kubeClient,
	}
}

func (ra *ResourceApplier) ApplyResources(ctx context.Context, oidcResource *unstructured.Unstructured, rbacResource *unstructured.Unstructured) clierror.Error {
	if err := ra.applyOIDCResource(ctx, oidcResource); err != nil {
		return err
	}

	if err := ra.applyRBACResource(ctx, rbacResource); err != nil {
		return err
	}

	return nil
}

// applyOIDCResource handles the OIDC resource application with conflict detection
func (ra *ResourceApplier) applyOIDCResource(ctx context.Context, oidcResource *unstructured.Unstructured) clierror.Error {
	resourceName := oidcResource.GetName()

	// Check if OpenIDConnect resource already exists
	existingOIDC, err := ra.kubeClient.Dynamic().Resource(kubeconfig.OpenIdConnectGVR).Get(ctx, resourceName, metav1.GetOptions{})
	if err == nil {
		if err := ra.checkOIDCConflicts(existingOIDC, oidcResource); err != nil {
			return err
		}
	} else if !errors.IsNotFound(err) {
		return clierror.Wrap(err, clierror.New("failed to check existing OpenIDConnect resource"))
	}

	// Apply the resource
	err = ra.kubeClient.RootlessDynamic().Apply(ctx, oidcResource, false)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to apply OpenIDConnect resource"))
	}

	fmt.Printf("OpenIDConnect resource '%s' applied successfully.\n", resourceName)
	return nil
}

// checkOIDCConflicts compares existing and new OIDC resources and handles conflicts
func (ra *ResourceApplier) checkOIDCConflicts(existing, new *unstructured.Unstructured) clierror.Error {
	resourceName := new.GetName()

	fmt.Fprintf(os.Stderr, "Warning: OpenIDConnect resource '%s' already exists.\n", resourceName)

	existingSpec, existingOk := existing.Object["spec"].(map[string]any)
	newSpec, newOk := new.Object["spec"].(map[string]any)

	if !existingOk || !newOk {
		return clierror.New("unable to compare OpenIDConnect resource specifications")
	}

	if !ra.compareOIDCSpecs(existingSpec, newSpec) {
		fmt.Fprintf(os.Stderr, "Error: The existing OpenIDConnect resource has different configuration. Please resolve the conflict manually or use a different name.\n")
		return clierror.New("configuration conflict detected - operation aborted")
	}

	fmt.Printf("OpenIDConnect resource '%s' has identical configuration - proceeding with update.\n", resourceName)
	return nil
}

func (ra *ResourceApplier) compareOIDCSpecs(existing, new map[string]any) bool {
	fields := []string{"issuerURL", "clientID", "usernameClaim", "usernamePrefix"}

	for _, field := range fields {
		if existing[field] != new[field] {
			return false
		}
	}

	existingClaims, ok1 := existing["requiredClaims"].(map[string]any)
	newClaims, ok2 := new["requiredClaims"].(map[string]any)

	if ok1 && ok2 {
		if existingClaims["repository"] != newClaims["repository"] {
			return false
		}
	} else if ok1 != ok2 {
		return false
	}

	return true
}

func (ra *ResourceApplier) applyRBACResource(ctx context.Context, rbacResource *unstructured.Unstructured) clierror.Error {
	resourceName := rbacResource.GetName()
	resourceKind := rbacResource.GetKind()

	err := ra.kubeClient.RootlessDynamic().Apply(ctx, rbacResource, false)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to apply RBAC resource"))
	}

	fmt.Printf("%s '%s' applied successfully.\n", resourceKind, resourceName)

	return nil
}
