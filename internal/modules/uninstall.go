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
func Uninstall(ctx context.Context, repo repo.ModuleTemplatesRepository, moduleTemplate *kyma.ModuleTemplate) clierror.Error {
	return uninstall(out.Default, ctx, repo, moduleTemplate)
}

func uninstall(printer *out.Printer, ctx context.Context, repo repo.ModuleTemplatesRepository, moduleTemplate *kyma.ModuleTemplate) clierror.Error {
	moduleName := moduleTemplate.Spec.ModuleName
	printer.Msgfln("removing %s community module from the target Kyma environment", moduleName)

	associatedResources, err := repo.RunningAssociatedResourcesOfModule(ctx, *moduleTemplate)
	if err != nil {
		return clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to get resources for the module %v: %v", moduleName, err)))
	}

	moduleResources, err := repo.Resources(ctx, *moduleTemplate)
	if err != nil {
		return clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to get resources for the module %v: %v", moduleName, err)))
	}

	// We want to remove resources in the reversed order
	slices.Reverse(moduleResources)
	slices.Reverse(associatedResources)

	moduleResourcesUnstruct := []unstructured.Unstructured{}
	for _, mr := range moduleResources {
		moduleResourcesUnstruct = append(moduleResourcesUnstruct, unstructured.Unstructured{Object: mr})
	}

	resourcesToDelete := slices.Concat(moduleResourcesUnstruct, associatedResources)

	removedSuccessfully := true
	for _, resource := range resourcesToDelete {
		resourceWatcher, err := repo.DeleteResourceReturnWatcher(ctx, resource)
		if err != nil {
			removedSuccessfully = false
			printer.Msgfln("failed to delete resource %s (%s): %v", resource.GetName(), resource.GetKind(), err)
			continue
		}
		printer.Msgfln("waiting for resource deletion: %s (%s)", resource.GetName(), resource.GetKind())
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*100)
		defer cancel()

		clierr := waitForDeletion(timeoutCtx, resourceWatcher)
		if clierr != nil {
			return clierr
		}
	}

	if removedSuccessfully {
		printer.Msgfln("%s community module successfully removed", moduleName)
	} else {
		printer.Msgfln("some errors occured during the %s community module removal", moduleName)
	}

	return nil
}

func GetRunningResourcesOfCommunityModule(ctx context.Context, repo repo.ModuleTemplatesRepository, moduleTemplate kyma.ModuleTemplate) ([]string, clierror.Error) {
	runningResources, err := repo.RunningUserDefinedResourcesOfModule(ctx, moduleTemplate)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to retrieve running resources of the %s module", moduleTemplate.Spec.ModuleName)))
	}

	var runningResourcesNames []string

	for _, resource := range runningResources {
		runningResourcesNames = append(runningResourcesNames, fmt.Sprintf("%s (%s)", resource.GetName(), resource.GetKind()))
	}

	return runningResourcesNames, nil
}
