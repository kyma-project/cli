package modules

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

	for _, moduleCR := range list.Items {
		fmt.Fprintf(writer, "removing %s/%s CR\n", moduleCR.GetNamespace(), moduleCR.GetName())
		err = client.RootlessDynamic().Remove(ctx, &moduleCR)
		if err != nil && !errors.IsNotFound(err) {
			return clierror.Wrap(err, clierror.New(
				fmt.Sprintf("failed to remove %s/%s cr", moduleCR.GetNamespace(), moduleCR.GetName()),
			))
		}
	}

	for _, moduleCR := range list.Items {
		fmt.Fprintf(writer, "waiting for %s/%s CR to be removed\n", moduleCR.GetNamespace(), moduleCR.GetName())
		clierr := waitForDeletion(ctx, client.RootlessDynamic(), &moduleCR)
		if clierr != nil {
			return clierr
		}
	}

	return nil
}

func waitForDeletion(ctx context.Context, client rootlessdynamic.Interface, obj *unstructured.Unstructured) clierror.Error {
	err := retry.Do(func() error {
		return isObjDeleted(ctx, client, obj)
	},
		retry.Delay(time.Second),
		retry.Context(ctx),
		retry.LastErrorOnly(true),
		retry.DelayType(retry.FixedDelay),
		retry.Attempts(100),
	)
	if err != nil {
		return clierror.Wrap(err, clierror.New(
			fmt.Sprintf("failed to check if %s/%s cr is removed", obj.GetNamespace(), obj.GetName()),
		))
	}

	return nil
}

func isObjDeleted(ctx context.Context, client rootlessdynamic.Interface, obj *unstructured.Unstructured) error {
	_, err := client.Get(ctx, obj)
	if errors.IsNotFound(err) {
		// expected ok
		return nil
	}

	if err != nil {
		return err
	}

	return fmt.Errorf("%s/%s exists on the cluster", obj.GetNamespace(), obj.GetName())
}
