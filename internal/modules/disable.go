package modules

import (
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"time"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	"github.com/kyma-project/cli.v3/internal/modules/repo"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
)

// Disable takes care about disabling module whatever if CustomResourcePolicy is set to Ignore or CreateAndDelete
// if CustomResourcePolicy is Ignore then it first deletes module CR and waits for removal
// at the end removes module from the Kyma CR
func Disable(ctx context.Context, client kube.Client, repo repo.ModuleTemplatesRepository, module string, community bool) clierror.Error {
	if community {
		return disableCommunity(os.Stdout, ctx, repo, module)
	}
	return disableCore(os.Stdout, ctx, client, module)
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

func disableCommunity(writer io.Writer, ctx context.Context, repo repo.ModuleTemplatesRepository, module string) clierror.Error {
	fmt.Fprintf(writer, "removing %s community module from the cluster\n", module)

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
			fmt.Fprintf(writer, "failed to delete resource %s (%s): %v\n", r.GetName(), r.GetKind(), err)
			continue
		}
		fmt.Fprintf(writer, "waiting for resource deletion: %s (%s)\n", r.GetName(), r.GetKind())
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*100)
		defer cancel()

		clierr := waitForDeletion(timeoutCtx, resourceWatcher)
		if clierr != nil {
			return clierr
		}
	}

	if removedSuccessfully {
		fmt.Fprintf(writer, "%s community module successfully removed\n", module)
	} else {
		fmt.Fprintf(writer, "some errors occured during the %s community module removal\n", module)
	}

	return nil
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

func disableCore(writer io.Writer, ctx context.Context, client kube.Client, module string) clierror.Error {
	clierr := removeModuleCR(writer, ctx, client, module)
	if clierr != nil {
		return clierr
	}

	fmt.Fprintf(writer, "removing %s module from the Kyma CR\n", module)
	err := client.Kyma().DisableModule(ctx, module)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to disable module"))
	}

	fmt.Fprintf(writer, "%s module disabled\n", module)
	return nil
}

func removeModuleCR(writer io.Writer, ctx context.Context, client kube.Client, module string) clierror.Error {
	info, err := client.Kyma().GetModuleInfo(ctx, module)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to get module info from the Kyma CR"))
	}

	if info.Spec.CustomResourcePolicy == kyma.CustomResourcePolicyCreateAndDelete {
		// lifecycle-manager is managing module cr on its own and we can't remove it manually
		return nil
	}

	moduleTemplate, err := client.Kyma().GetModuleTemplateForModule(ctx, info.Status.Name, info.Status.Channel)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to get ModuleTemplate CR for module"))
	}

	defaultCR := moduleTemplate.Spec.Data
	if len(defaultCR.Object) == 0 {
		// module has no custom CR
		return nil
	}

	list, err := client.RootlessDynamic().List(ctx, &defaultCR, &rootlessdynamic.ListOptions{
		AllNamespaces: true,
	})
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to list module CRs"))
	}

	watchers := make([]watch.Interface, len(list.Items))
	for i, moduleCR := range list.Items {
		watchers[i], err = client.RootlessDynamic().WatchSingleResource(ctx, &moduleCR)
		if err != nil {
			return clierror.Wrap(err, clierror.New(
				fmt.Sprintf("failed to watch resource %s/%s", moduleCR.GetNamespace(), moduleCR.GetName()),
			))
		}
		defer watchers[i].Stop()
	}

	for _, moduleCR := range list.Items {
		fmt.Fprintf(writer, "removing %s/%s CR\n", moduleCR.GetNamespace(), moduleCR.GetName())
		err = client.RootlessDynamic().Remove(ctx, &moduleCR, false)
		if err != nil && !errors.IsNotFound(err) {
			return clierror.Wrap(err, clierror.New(
				fmt.Sprintf("failed to remove %s/%s cr", moduleCR.GetNamespace(), moduleCR.GetName()),
			))
		}
	}

	for i, moduleCR := range list.Items {
		fmt.Fprintf(writer, "waiting for %s/%s CR to be removed\n", moduleCR.GetNamespace(), moduleCR.GetName())
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*100)
		defer cancel()

		clierr := waitForDeletion(timeoutCtx, watchers[i])
		if clierr != nil {
			return clierr
		}
	}

	return nil
}

func waitForDeletion(ctx context.Context, watcher watch.Interface) clierror.Error {
	for {
		select {
		case <-ctx.Done():
			return clierror.Wrap(ctx.Err(), clierror.New("context timeout"))
		case event := <-watcher.ResultChan():
			if event.Type == watch.Deleted {
				return nil
			}
		}
	}
}
