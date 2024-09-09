package cluster

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/communitymodules"
	"github.com/kyma-project/cli.v3/internal/kube/resources"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	"k8s.io/apimachinery/pkg/api/errors"
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
func ApplySpecifiedModules(ctx context.Context, client rootlessdynamic.Interface, desiredModules []ModuleInfo, crs []unstructured.Unstructured) clierror.Error {
	available, err := communitymodules.GetAvailableModules()
	if err != nil {
		return err
	}

	modules, err := downloadSpecifiedModules(desiredModules, available)
	if err != nil {
		return err
	}

	return applySpecifiedModules(ctx, client, modules, crs)
}

func RemoveSpecifiedModules(ctx context.Context, client rootlessdynamic.Interface, desiredModules []ModuleInfo) clierror.Error {
	available, err := communitymodules.GetAvailableModules()
	if err != nil {
		return err
	}

	modules, err := downloadSpecifiedModules(desiredModules, available)
	if err != nil {
		return err
	}

	return removeSpecifiedModules(ctx, client, modules)
}

type moduleDetails struct {
	name      string
	version   string
	cr        unstructured.Unstructured
	resources []unstructured.Unstructured
}

func downloadSpecifiedModules(desiredModules []ModuleInfo, availableModules communitymodules.Modules) ([]moduleDetails, clierror.Error) {
	modules := []moduleDetails{}
	for _, module := range availableModules {
		moduleInfo := containsModule(module.Name, desiredModules)
		if moduleInfo == nil {
			// module is not specified
			continue
		}

		desiredVersion := getDesiredVersion(*moduleInfo, module.Versions)

		crURL := desiredVersion.CrYaml
		cr, err := downloadUnstructuredList(crURL)
		if err != nil {
			return nil, clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to download cr from '%s' url", crURL)))
		}

		manifestURL := desiredVersion.DeploymentYaml
		resources, err := downloadUnstructuredList(manifestURL)
		if err != nil {
			return nil, clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to download manifest from '%s' url", manifestURL)))
		}

		modules = append(modules, moduleDetails{
			name:      module.Name,
			version:   desiredVersion.Version,
			cr:        cr[0], // its expected that cr will always have len==1
			resources: resources,
		})
	}

	return modules, nil
}

func downloadUnstructuredList(url string) ([]unstructured.Unstructured, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code '%d'", resp.StatusCode)
	}

	return resources.DecodeYaml(resp.Body)
}

func applySpecifiedModules(ctx context.Context, client rootlessdynamic.Interface, desiredModules []moduleDetails, customConfig []unstructured.Unstructured) clierror.Error {
	for _, module := range desiredModules {
		fmt.Printf("Applying %s module\n", module.name)
		err := client.ApplyMany(ctx, module.resources)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to apply module resources"))
		}

		fmt.Println("Applying CR")
		cr := chooseModuleCR(module, customConfig)

		err = client.Apply(ctx, &cr)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to apply module cr"))
		}
	}
	return nil
}

func chooseModuleCR(module moduleDetails, customConfig []unstructured.Unstructured) unstructured.Unstructured {
	customCRIndex := slices.IndexFunc(customConfig, func(u unstructured.Unstructured) bool {
		return u.GetKind() == module.cr.GetKind() && u.GetAPIVersion() == module.cr.GetAPIVersion()
	})

	if customCRIndex >= 0 {
		return customConfig[customCRIndex]
	}

	return module.cr
}

func containsModule(have string, want []ModuleInfo) *ModuleInfo {
	for _, rec := range want {
		if have == rec.Name {
			return &rec
		}
	}
	return nil
}

func getDesiredVersion(moduleInfo ModuleInfo, versions []communitymodules.Version) communitymodules.Version {
	if moduleInfo.Version != "" {
		for _, version := range versions {
			if version.Version == moduleInfo.Version {
				// TODO: what if the user passes a version that does not exist?
				// shall we for sure install the latest version?
				fmt.Printf("Version %s found for %s\n", version.Version, moduleInfo.Name)
				return version
			}
		}
	}

	fmt.Printf("Using latest version for %s\n", moduleInfo.Name)
	return communitymodules.GetLatestVersion(versions)
}

func removeSpecifiedModules(ctx context.Context, client rootlessdynamic.Interface, desiredModules []moduleDetails) clierror.Error {
	for _, module := range desiredModules {
		fmt.Printf("Removing CR\n")
		err := client.Remove(ctx, &module.cr)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to remove module cr"))
		}

		cliErr := RetryUntilRemoved(ctx, client, module)
		if cliErr != nil {
			return cliErr
		}

		fmt.Printf("Removing %s module\n", module.name)
		err = client.RemoveMany(ctx, module.resources)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to remove module resources"))
		}
	}
	return nil
}

func RetryUntilRemoved(ctx context.Context, client rootlessdynamic.Interface, module moduleDetails) clierror.Error {
	return retryUntilRemoved(ctx, client, module, 50)
}

func retryUntilRemoved(ctx context.Context, client rootlessdynamic.Interface, module moduleDetails, attempts uint) clierror.Error {
	retryErr := retry.Do(func() error {
		object, err := client.Get(ctx, &module.cr)
		if object != nil {
			return fmt.Errorf("object still exists")
		}
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	},
		retry.Delay(1*time.Second),
		retry.DelayType(retry.FixedDelay),
		retry.Attempts(attempts),
		retry.LastErrorOnly(true),
		retry.Context(ctx),
	)
	if retryErr != nil {
		return clierror.Wrap(retryErr, clierror.New("failed to remove module cr"))
	}
	return nil
}
