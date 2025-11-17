package modules

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modules/repo"
	"github.com/kyma-project/cli.v3/internal/out"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Uninstall takes care of removing the community module from the target Kyma environment.
// It retrieves the module template, gets all associated resources, and deletes them in reverse order.
func Uninstall(ctx context.Context, repo repo.ModuleTemplatesRepository, module string) clierror.Error {
	return uninstall(out.Default, ctx, repo, module)
}

func uninstall(printer *out.Printer, ctx context.Context, repo repo.ModuleTemplatesRepository, module string) clierror.Error {
	printer.Msgfln("removing %s community module from the target Kyma environment", module)

	moduleTemplateToDelete, err := getModuleTemplateToDelete(ctx, repo, module)
	if err != nil {
		return clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to retrieve the module %v", module)))
	}
	moduleResources, err := repo.Resources(ctx, *moduleTemplateToDelete)
	if err != nil {
		return clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to get resources for module %v", moduleTemplateToDelete.Spec.ModuleName)))
	}

	// We want to remove resources in the reversed order
	slices.Reverse(moduleResources)

	removedSuccessfully := true

	for _, resource := range moduleResources {
		resourceWatcher, err := repo.DeleteResourceReturnWatcher(ctx, resource)
		r := unstructured.Unstructured{Object: resource}
		if err != nil {
			removedSuccessfully = false
			printer.Msgfln("failed to delete resource %s (%s): %v", r.GetName(), r.GetKind(), err)
			continue
		}
		printer.Msgfln("waiting for resource deletion: %s (%s)", r.GetName(), r.GetKind())
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*100)
		defer cancel()

		clierr := waitForDeletion(timeoutCtx, resourceWatcher)
		if clierr != nil {
			return clierr
		}
	}

	if removedSuccessfully {
		printer.Msgfln("%s community module successfully removed", module)
	} else {
		printer.Msgfln("some errors occured during the %s community module removal", module)
	}

	return nil
}

func GetRunningResourcesOfCommunityModule(ctx context.Context, repo repo.ModuleTemplatesRepository, module string) ([]string, clierror.Error) {
	moduleTemplateToDelete, err := getModuleTemplateToDelete(ctx, repo, module)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to retrieve the module %v", module)))
	}
	runningResources, err := repo.RunningAssociatedResourcesOfModule(ctx, *moduleTemplateToDelete)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to retrieve running resources of the %s module", module)))
	}

	var runningResourcesNames []string

	for _, resource := range runningResources {
		runningResourcesNames = append(runningResourcesNames, fmt.Sprintf("%s (%s)", resource.GetName(), resource.GetKind()))
	}

	return runningResourcesNames, nil
}

func getModuleTemplateToDelete(ctx context.Context, repo repo.ModuleTemplatesRepository, module string) (*kyma.ModuleTemplate, error) {
	installedModulesWithName, err := repo.CommunityInstalledByName(ctx, module)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve a list of installed community modules: %v", err)
	}
	if len(installedModulesWithName) == 0 {
		return nil, fmt.Errorf("failed to find any version of the module %s", module)
	}
	if len(installedModulesWithName) > 1 {
		return nil, fmt.Errorf("failed to determine module version for %s", module)
	}

	return &installedModulesWithName[0], nil
}
