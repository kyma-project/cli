package modules

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type InstallCommunityModuleData struct {
	ModuleName            string
	Version               string
	IsDefaultCRApplicable bool
	CustomResources       []unstructured.Unstructured
}

// Install takes care of enabling the community module on the cluster.
// 1. resources defined on the module template are applied
// 2. if default custom resource should be applied then it's applied from the installed module template
// 3. if custom resource from file is present, the file is read and resources are applied
func Install(ctx context.Context, client kube.Client, data InstallCommunityModuleData) clierror.Error {
	existingModule, err := findModuleInCatalog(ctx, client, data.ModuleName, data.Version, true)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to retrieve community module from catalog"))
	}

	if err := installModuleResources(ctx, client, existingModule); err != nil {
		return clierror.Wrap(err, clierror.New("failed to install community module"))
	}

	if err := applyDefaultCustomResource(ctx, client, existingModule, data.IsDefaultCRApplicable); err != nil {
		return clierror.Wrap(err, clierror.New("failed to apply default custom resource"))
	}

	if err := applyCustomResourcesFromFile(ctx, client, data.CustomResources); err != nil {
		return clierror.Wrap(err, clierror.New("failed to apply custom resource files"))
	}

	fmt.Printf("%s community module enabled\n", data.ModuleName)
	return nil
}

func findModuleInCatalog(ctx context.Context, client kube.Client, moduleName string, version string, isCommunity bool) (*kyma.ModuleTemplate, error) {
	moduleTemplates, err := client.Kyma().ListModuleTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get module template")
	}

	var foundModule *kyma.ModuleTemplate

	for _, moduleTemplate := range moduleTemplates.Items {
		currModuleName := moduleTemplate.Spec.ModuleName
		currVersion := moduleTemplate.Spec.Version
		currCommunityModule := isCommunityModule(&moduleTemplate)

		moduleNameMatched := currModuleName == moduleName
		versionMatched := currVersion == version
		communityModuleMatched := currCommunityModule == isCommunity

		if moduleNameMatched && versionMatched && communityModuleMatched {
			foundModule = &moduleTemplate
			break
		}
	}

	if foundModule == nil {
		return nil, fmt.Errorf("module not found")
	}

	return foundModule, nil
}

func installModuleResources(ctx context.Context, client kube.Client, existingModule *kyma.ModuleTemplate) error {
	for _, res := range existingModule.Spec.Resources {
		url := res.Link
		if err := applyResourcesFromURL(ctx, client, url); err != nil {
			return errors.Wrap(err, "failed to apply resources from link")
		}
	}

	return nil
}

func applyResourcesFromURL(ctx context.Context, client kube.Client, url string) error {
	resourceYamlStrings, err := getResourceYamlStringsFromURL(url)
	if err != nil {
		return err
	}

	var installedResources []map[string]any

	for _, resourceYamlStr := range resourceYamlStrings {
		var obj map[string]any
		if err := yaml.Unmarshal([]byte(resourceYamlStr), &obj); err != nil {
			rollback(ctx, client, installedResources)
			return fmt.Errorf("failed to parse module resource: %w", err)
		}

		if err := client.RootlessDynamic().Apply(ctx, &unstructured.Unstructured{Object: obj}, false); err != nil {
			rollback(ctx, client, installedResources)
			return fmt.Errorf("failed to apply resource: %w", err)
		}

		installedResources = append(installedResources, obj)
	}
	return nil
}

func getResourceYamlStringsFromURL(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download resource from %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource body: %w", err)
	}

	return strings.Split(string(body), "---"), nil
}

func applyDefaultCustomResource(ctx context.Context, client kube.Client,
	existingModule *kyma.ModuleTemplate, isDefaultCRApplicable bool) error {
	if !isDefaultCRApplicable {
		return nil
	}

	defaultCustomResourceUnstructured := existingModule.Spec.Data

	if err := client.RootlessDynamic().Apply(ctx, &defaultCustomResourceUnstructured, false); err != nil {
		return fmt.Errorf("failed to apply default custom resource: %w", err)
	}

	return nil
}

func applyCustomResourcesFromFile(ctx context.Context, client kube.Client, customResources []unstructured.Unstructured) error {
	if len(customResources) == 0 {
		return nil
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*100)
	defer cancel()

	for _, customResource := range customResources {
		err := client.RootlessDynamic().Apply(timeoutCtx, &customResource, false)
		if err != nil {
			return fmt.Errorf("failed to apply custom resource from path")
		}
	}

	return nil
}

func rollback(ctx context.Context, client kube.Client, resources []map[string]any) error {
	if len(resources) == 0 {
		return nil
	}

	for _, resource := range resources {
		client.RootlessDynamic().Remove(ctx, &unstructured.Unstructured{Object: resource}, false)
	}

	return nil
}
