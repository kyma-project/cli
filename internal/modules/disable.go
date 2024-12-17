package modules

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/watch"
)

// Disable takes care about disabling module whatever if CustomResourcePolicy is set to Ignore or CreateAndDelete
// if CustomResourcePolicy is Ignore then it first deletes module CR and waits for removal
// at the end removes module from the Kyma CR
func Disable(ctx context.Context, client kube.Client, module string) clierror.Error {
	return disable(os.Stdout, ctx, client, module)
}

func disable(writer io.Writer, ctx context.Context, client kube.Client, module string) clierror.Error {
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

	moduleTemplate, err := client.Kyma().GetModuleTemplateForModule(ctx, info.Status.Name, info.Status.Version)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to get ModuleTemplate CR for module"))
	}

	defaultCR := moduleTemplate.Spec.Data
	if len(defaultCR.Object) == 0 {
		// module has no custom CR
		return nil
	}

	list, err := client.RootlessDynamic().List(ctx, &defaultCR)
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
		err = client.RootlessDynamic().Remove(ctx, &moduleCR)
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
