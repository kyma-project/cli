package modules

import (
	"context"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modules/repo"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// GetModuleTemplateFromRemote retrieves a specific module template from remote repositories.
// It searches for a module by name and version in the remote community module catalog.
//
// Parameters:
//   - ctx: Context for the operation
//   - repo: ModuleTemplatesRepository interface for accessing remote modules
//   - moduleName: The name of the module to retrieve
//   - version: The specific version of the module to retrieve
//
// Returns:
//   - *kyma.ModuleTemplate: The found module template
//   - error: Error if module is not found or if there's an issue accessing the repository
//
// Example:
//
//	moduleTemplate, err := GetModuleTemplateFromRemote(ctx, repo, "my-module", "v1.0.0")
func GetModuleTemplateFromRemote(ctx context.Context, repo repo.ModuleTemplatesRepository, moduleName, version string) (*kyma.ModuleTemplate, error) {
	remoteModules, err := repo.ExternalCommunityByNameAndVersion(ctx, moduleName, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get module %s: %v", moduleName, err)
	}

	if len(remoteModules) == 1 {
		return &remoteModules[0], nil
	}

	return nil, fmt.Errorf("module not found in the catalog: try running `module catalog` command to verify available community modules")
}

// PersistModuleTemplateInNamespace saves a module template to a specific namespace in the cluster.
// It converts the module template to an unstructured object and applies it using the dynamic client.
//
// Parameters:
//   - ctx: Context for the operation
//   - client: Kubernetes client for cluster operations
//   - moduleTemplate: The module template to persist
//   - namespace: The target namespace where the module template should be stored
//
// Returns:
//   - error: Error if the module template cannot be converted or applied to the cluster
//
// Example:
//
//	err := PersistModuleTemplateInNamespace(ctx, client, moduleTemplate, "kyma-system")
func PersistModuleTemplateInNamespace(ctx context.Context, client kube.Client, moduleTemplate *kyma.ModuleTemplate, namespace string) error {
	moduleTemplate.Namespace = namespace
	unstructuredModule, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&moduleTemplate)
	if err != nil {
		return err
	}

	return client.RootlessDynamic().Apply(ctx, &unstructured.Unstructured{Object: unstructuredModule}, false)
}
