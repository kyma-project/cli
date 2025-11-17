package modules

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modules/repo"
	"github.com/kyma-project/cli.v3/internal/out"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Enable takes care about enabling kyma module in order:
// 1. add module to the Kyma CR with CustomResourcePolicy set to CreateAndDelete if defaultCR is true and to Ignore in any other case
// 2. if crs array is not empty wait for the module to be ready and add crs to the cluster
func Enable(ctx context.Context, client kube.Client, repo repo.ModuleTemplatesRepository, module, channel string, defaultCR bool, crs ...unstructured.Unstructured) clierror.Error {
	return enable(out.Default, ctx, client, repo, module, channel, defaultCR, crs...)
}

func enable(printer *out.Printer, ctx context.Context, client kube.Client, repo repo.ModuleTemplatesRepository, module, channel string, defaultCR bool, crs ...unstructured.Unstructured) clierror.Error {
	if err := validateModuleAvailability(ctx, client, repo, module); err != nil {
		hints := []string{
			"make sure you provide valid module name and channel (or version)",
			"list available modules by calling the `kyma module catalog` command",
			"pull available modules by calling the `kyma module pull` command",
			"if you want to add a community module, use the `--origin` flag",
		}
		return clierror.Wrap(err, clierror.New("unknown module name", hints...))
	}

	crPolicy := kyma.CustomResourcePolicyIgnore
	if defaultCR {
		crPolicy = kyma.CustomResourcePolicyCreateAndDelete
	}

	printer.Msgfln("adding %s module to the Kyma CR", module)
	err := client.Kyma().EnableModule(ctx, module, channel, crPolicy)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to enable module"))
	}

	clierr := applyCustomCR(printer, ctx, client, module, crs...)
	if clierr != nil {
		return clierr
	}

	printer.Msgfln("%s module enabled", module)
	return nil
}

func applyCustomCR(printer *out.Printer, ctx context.Context, client kube.Client, module string, crs ...unstructured.Unstructured) clierror.Error {
	if len(crs) == 0 {
		// skip if there is nothing to do
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*100)
	defer cancel()

	printer.Msgln("waiting for module to be ready")
	err := client.Kyma().WaitForModuleState(ctx, module, "Ready", "Warning")
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to check module state"))
	}

	for _, cr := range crs {
		printer.Msgfln("applying %s/%s cr", cr.GetNamespace(), cr.GetName())
		err = client.RootlessDynamic().Apply(ctx, &cr, false)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to apply custom cr from path"))
		}
	}

	return nil
}

func validateModuleAvailability(ctx context.Context, client kube.Client, repo repo.ModuleTemplatesRepository, module string) error {
	availableCoreVersions, err := ListAvailableVersions(ctx, client, repo, module, false)
	if err != nil {
		return err
	}

	if len(availableCoreVersions) == 0 {
		return fmt.Errorf("%s not found in the catalog of available modules", module)
	}

	return nil
}
