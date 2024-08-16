package cluster

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/communitymodules"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ModuleInfo struct {
	Name    string
	Version string
}

// ParseModules returns ModuleInfo struct based on the string input.
// Can convert 'name' or 'name:version' into struct
func ParseModules(modules []string) []ModuleInfo {
	// TODO: I can't find better place for this function (move it)
	var moduleInfo []ModuleInfo
	for _, module := range modules {
		if module == "" {
			// skip empty strings
			continue
		}

		elems := strings.Split(module, ":")
		info := ModuleInfo{
			Name: elems[0],
		}

		if len(elems) > 1 {
			info.Version = elems[1]
		}

		moduleInfo = append(moduleInfo, info)
	}

	return moduleInfo
}

// ApplySpecifiedModules applies modules to the cluster based on the resources from the community module json (Github)
// if module cr is in the given crs list then it will be applied instead of one from the community module json
func ApplySpecifiedModules(ctx context.Context, client rootlessdynamic.Interface, modules []ModuleInfo, crs []unstructured.Unstructured) clierror.Error {
	available, err := communitymodules.GetAvailableModules()
	if err != nil {
		return err
	}

	return applySpecifiedModules(ctx, client, modules, crs, available)
}

func applySpecifiedModules(ctx context.Context, client rootlessdynamic.Interface, modules []ModuleInfo, crs []unstructured.Unstructured, availableModules communitymodules.Modules) clierror.Error {
	for _, rec := range availableModules {
		moduleInfo := containsModule(rec.Name, modules)
		if moduleInfo == nil {
			continue
		}

		wantedVersion := verifyVersion(*moduleInfo, rec)
		fmt.Printf("Applying %s module manifest\n", rec.Name)

		err := applyGivenObjects(ctx, client, wantedVersion.Resources...)
		if err != nil {
			return err
		}

		if applyGivenCustomCR(ctx, client, rec, crs) {
			fmt.Println("Applying custom CR")
			continue
		}

		fmt.Println("Applying CR")
		err = applyGivenObjects(ctx, client, wantedVersion.CR)
		if err != nil {
			return err
		}
	}
	return nil
}

func containsModule(have string, want []ModuleInfo) *ModuleInfo {
	for _, rec := range want {
		if have == rec.Name {
			return &rec
		}
	}
	return nil
}

func verifyVersion(moduleInfo ModuleInfo, rec communitymodules.Module) communitymodules.Version {
	if moduleInfo.Version != "" {
		for _, version := range rec.Versions {
			if version.Version == moduleInfo.Version {
				// TODO: what if the user passes a version that does not exist?
				// shall we for sure install the latest version?
				fmt.Printf("Version %s found for %s\n", version.Version, rec.Name)
				return version
			}
		}
	}

	fmt.Printf("Using latest version for %s\n", rec.Name)
	return communitymodules.GetLatestVersion(rec.Versions)
}

// applyGivenCustomCR applies custom CR if it exists
func applyGivenCustomCR(ctx context.Context, client rootlessdynamic.Interface, rec communitymodules.Module, config []unstructured.Unstructured) bool {
	for _, obj := range config {
		if strings.EqualFold(obj.GetKind(), strings.ToLower(rec.Name)) {
			client.Apply(ctx, &obj)
			return true
		}
	}
	return false

}

func applyGivenObjects(ctx context.Context, client rootlessdynamic.Interface, resources ...communitymodules.Resource) clierror.Error {
	objects := []unstructured.Unstructured{}
	for _, res := range resources {
		objects = append(objects, unstructured.Unstructured{
			Object: res,
		})
	}

	err := client.ApplyMany(ctx, objects)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to apply module resources"))

	}
	return nil
}
