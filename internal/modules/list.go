package modules

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	"github.com/kyma-project/cli.v3/internal/modules/repo"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Module struct {
	Name            string
	Versions        []ModuleVersion
	InstallDetails  ModuleInstallDetails
	CommunityModule bool
	Origin          string
}

type Managed string

const (
	ManagedTrue     Managed = "true"
	ManagedFalse    Managed = "false"
	UnknownValue            = "Unknown"
	NotRunningValue         = "NotRunning"
)

type ModuleInstallDetails struct {
	Version              string
	Channel              string
	Managed              Managed
	CustomResourcePolicy string
	// Possible states: https://github.com/kyma-project/lifecycle-manager/blob/main/api/shared/state.go
	ModuleState       string
	InstallationState string
}

type ModuleVersion struct {
	Repository string
	Version    string
	Channel    string
}

type ModulesList []Module

// ListInstalled returns list of installed module on a cluster
// collects info about modules based on the KymaCR
func ListInstalled(ctx context.Context, client kube.Client, repo repo.ModuleTemplatesRepository, showErrors bool) (ModulesList, error) {
	installedCoreModules, err := listCoreInstalled(ctx, client, repo, showErrors)
	if err != nil {
		return nil, err
	}

	installedCommunityModules, err := listCommunityInstalled(ctx, client, repo, installedCoreModules)
	if err != nil {
		return nil, err
	}

	allInstalled := append(installedCoreModules, installedCommunityModules...)
	return allInstalled, nil
}

func listCoreInstalled(ctx context.Context, client kube.Client, repo repo.ModuleTemplatesRepository, showErrors bool) (ModulesList, error) {
	defaultKyma, err := client.Kyma().GetDefaultKyma(ctx)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, errors.Wrap(err, "failed to get default Kyma CR from the target Kyma environment")
	}
	if err != nil && apierrors.IsNotFound(err) {
		// if cluster is not managed by KLM we won't be looking into Kyma CR spec
		return ModulesList{}, nil
	}

	kymaCRModules := make(chan ModulesList)
	unmanagedModules := make(chan ModulesList)

	go func() {
		kymaCRModules <- collectModulesFromKymaCR(ctx, client, defaultKyma)
	}()
	go func() {
		unmanagedModules <- collectUnmanagedCoreModules(ctx, client, repo, defaultKyma, showErrors)
	}()

	return append(<-kymaCRModules, <-unmanagedModules...), nil
}

func collectModulesFromKymaCR(ctx context.Context, client kube.Client, defaultKyma *kyma.Kyma) ModulesList {
	modulesList := ModulesList{}
	modulesLen := len(defaultKyma.Status.Modules)
	results := make(chan Module, modulesLen)
	var wg sync.WaitGroup

	for _, moduleStatus := range defaultKyma.Status.Modules {
		wg.Add(1)
		go func(ms kyma.ModuleStatus) {
			defer wg.Done()
			moduleSpec := getKymaModuleSpec(defaultKyma, ms.Name)

			installationState, err := getModuleInstallationState(ctx, client, ms, moduleSpec)
			if err != nil {
				fmt.Printf("error occured during %s module installation status check: %v\n", ms.Name, err)
			}

			moduleCRState, err := getModuleCustomResourceStatus(ctx, client, ms, moduleSpec)
			if err != nil {
				fmt.Printf("error occured during %s custom resource status check: %v\n", ms.Name, err)
			}

			results <- Module{
				Name: ms.Name,
				InstallDetails: ModuleInstallDetails{
					Channel:              ms.Channel,
					Managed:              getManaged(moduleSpec),
					CustomResourcePolicy: getCustomResourcePolicy(moduleSpec),
					Version:              ms.Version,
					ModuleState:          moduleCRState,
					InstallationState:    installationState,
				},
				Origin: "kyma",
			}
		}(moduleStatus)
	}

	wg.Wait()
	close(results)

	for m := range results {
		modulesList = append(modulesList, m)
	}

	return modulesList
}

func collectUnmanagedCoreModules(ctx context.Context, client kube.Client, repo repo.ModuleTemplatesRepository, defaultKyma *kyma.Kyma, showErrors bool) ModulesList {
	coreModuleTemplates, err := repo.Core(ctx)
	if err != nil {
		if showErrors {
			fmt.Printf("failed to get unmanaged core modules: %v\n", err)
		}
		return ModulesList{}
	}

	modulesLen := len(coreModuleTemplates)
	results := make(chan Module, modulesLen)
	var wg sync.WaitGroup

	for _, coreModuleTemplate := range coreModuleTemplates {
		wg.Add(1)
		go func(mt kyma.ModuleTemplate) {
			defer wg.Done()
			if mt.Spec.Version == "" || moduleExistsInKymaCR(mt, defaultKyma.Status.Modules) {
				return
			}

			installedManager, err := repo.InstalledManager(ctx, mt)
			if err != nil {
				if showErrors {
					fmt.Printf("failed to get installed manager: %v\n", err)
				}
				return
			}
			if installedManager == nil {
				return
			}

			moduleStatus := getModuleStatus(ctx, client, mt.Spec.Data)
			version, err := getManagerVersion(installedManager)
			if err != nil {
				if showErrors {
					fmt.Printf("failed to get managers version: %v\n", err)
				}
				return
			}

			results <- Module{
				Name: mt.Spec.ModuleName,
				InstallDetails: ModuleInstallDetails{
					Channel:              "",
					Managed:              ManagedFalse,
					CustomResourcePolicy: "N/A",
					Version:              version,
					ModuleState:          moduleStatus,
					InstallationState:    "Unmanaged",
				},
				Origin:          "kyma",
				CommunityModule: true,
			}
		}(coreModuleTemplate)
	}

	wg.Wait()
	close(results)

	modulesList := ModulesList{}
	for m := range results {
		var moduleAlreadyExists bool
		for _, module := range modulesList {
			if module.Name == m.Name && module.InstallDetails.Version == m.InstallDetails.Version {
				moduleAlreadyExists = true
				break
			}
		}

		if !moduleAlreadyExists {
			modulesList = append(modulesList, m)
		}
	}

	return modulesList
}

func moduleExistsInKymaCR(coreModuleTemplate kyma.ModuleTemplate, moduleStatuses []kyma.ModuleStatus) bool {
	for _, moduleStatus := range moduleStatuses {
		if moduleStatus.Name == coreModuleTemplate.Spec.ModuleName {
			return true
		}
	}

	return false
}

// ListCatalog returns list of module catalog on a cluster
// collects info about modules based on ModuleTemplates and ModuleReleaseMetas
func ListCatalog(ctx context.Context, client kube.Client, repo repo.ModuleTemplatesRepository) (ModulesList, error) {
	var coreModulesList ModulesList
	var communityModulesList ModulesList
	var externalModulesList ModulesList
	var err error

	if isClusterManagedByKLM(ctx, client) {
		coreModulesList, err = listCoreModulesCatalog(ctx, client)
		if err != nil {
			return nil, fmt.Errorf("failed to list modules catalog: %v", err)
		}
	}

	communityModulesList, err = listCommunityModulesCatalog(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to list modules catalog: %v", err)
	}

	externalModulesList, err = listExternalModulesCatalog(ctx, repo)
	if err != nil {
		fmt.Printf("failed to list external modules catalog: %v", err)
	}

	allModules := slices.Concat(coreModulesList, communityModulesList, externalModulesList)

	return allModules, nil
}

func listCoreModulesCatalog(ctx context.Context, client kube.Client) (ModulesList, error) {
	moduleTemplates, err := client.Kyma().ListModuleTemplate(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list all ModuleTemplate resources from the target Kyma environment")
	}

	moduleReleaseMetas, err := client.Kyma().ListModuleReleaseMeta(ctx)
	if err != nil {
		moduleList := listOldModulesCatalog(moduleTemplates)
		if len(moduleList) != 0 {
			return moduleList, nil
		}
		return nil, errors.New("failed to list module catalog from the target Kyma environment")
	}

	modulesList := ModulesList{}
	for _, moduleTemplate := range moduleTemplates.Items {
		moduleName := moduleTemplate.Spec.ModuleName
		if moduleName == "" || isCommunityModule(&moduleTemplate) {
			// ignore incompatible/corrupted ModuleTemplates
			continue
		}
		version := ModuleVersion{
			Version:    moduleTemplate.Spec.Version,
			Repository: moduleTemplate.Spec.Info.Repository,
			Channel: getAssignedChannel(
				*moduleReleaseMetas,
				moduleName,
				moduleTemplate.Spec.Version,
			),
		}

		if i := getModuleIndex(modulesList, moduleName, false); i != -1 {
			// append version if module with same name is in the list
			modulesList[i].Versions = append(modulesList[i].Versions, version)
		} else {
			// otherwise create a new record in the list
			modulesList = append(modulesList, Module{
				Name: moduleName,
				Versions: []ModuleVersion{
					version,
				},
				Origin:          "kyma",
				CommunityModule: false,
			})
		}
	}

	return modulesList, nil
}

func listCommunityModulesCatalog(ctx context.Context, repo repo.ModuleTemplatesRepository) (ModulesList, error) {
	communityModules, err := repo.Community(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query community modules: %v", err)
	}

	modulesList := ModulesList{}

	for _, communityModule := range communityModules {
		moduleName := communityModule.Spec.ModuleName
		version := ModuleVersion{
			Version:    communityModule.Spec.Version,
			Repository: communityModule.Spec.Info.Repository,
		}

		if i := getModuleIndex(modulesList, moduleName, true); i != -1 {
			modulesList[i].Versions = append(modulesList[i].Versions, version)
		} else {
			modulesList = append(modulesList, Module{
				Name: moduleName,
				Versions: []ModuleVersion{
					version,
				},
				Origin:          communityModule.Namespace + "/" + communityModule.Name,
				CommunityModule: true,
			})
		}
	}

	return modulesList, nil
}

func listExternalModulesCatalog(ctx context.Context, repo repo.ModuleTemplatesRepository) (ModulesList, error) {
	externalModules, err := repo.ExternalCommunity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query external modules catalog: %v", err)
	}

	modulesList := ModulesList{}

	for _, communityModule := range externalModules {
		moduleName := communityModule.Spec.ModuleName
		version := ModuleVersion{
			Version:    communityModule.Spec.Version,
			Repository: communityModule.Spec.Info.Repository,
		}

		if i := getModuleIndex(modulesList, moduleName, true); i != -1 {
			modulesList[i].Versions = append(modulesList[i].Versions, version)
		} else {
			modulesList = append(modulesList, Module{
				Name: moduleName,
				Versions: []ModuleVersion{
					version,
				},
				Origin:          "community",
				CommunityModule: true,
			})
		}
	}

	return modulesList, nil
}

func isClusterManagedByKLM(ctx context.Context, client kube.Client) bool {
	_, err := client.Kyma().GetDefaultKyma(ctx)
	return err == nil
}

func ListAvailableVersions(ctx context.Context, client kube.Client, repo repo.ModuleTemplatesRepository, moduleName string, isCommunity bool) ([]string, error) {
	catalog, err := ListCatalog(ctx, client, repo)
	if err != nil {
		return nil, err
	}

	var module Module

	for _, m := range catalog {
		if m.CommunityModule == isCommunity && m.Name == moduleName {
			module = m
			break
		}
	}

	var moduleVersions []string

	for _, mv := range module.Versions {
		moduleVersions = append(moduleVersions, mv.Version)
	}

	return moduleVersions, nil
}

func listCommunityInstalled(ctx context.Context, client kube.Client, repo repo.ModuleTemplatesRepository, installedCoreModules ModulesList) (ModulesList, error) {
	communityModuleTemplates, err := repo.Community(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list community module templates: %v", err)
	}

	modulesLen := len(communityModuleTemplates)
	results := make(chan Module, modulesLen)
	var wg sync.WaitGroup

	for _, moduleTemplate := range communityModuleTemplates {
		wg.Add(1)
		go func(mt kyma.ModuleTemplate) {
			defer wg.Done()
			if moduleAlreadyInstalledAsCoreModule(installedCoreModules, mt) {
				return
			}
			installedManager, err := repo.InstalledManager(ctx, mt)
			if err != nil {
				fmt.Printf("failed to get installed manager: %v\n", err)
				return
			}
			if installedManager == nil {
				// skip modules which moduletemplates exist but are not installed
				return
			}

			moduleStatus := getModuleStatus(ctx, client, mt.Spec.Data)
			installationStatus := getManagerStatus(installedManager)
			version, err := getManagerVersion(installedManager)
			if err != nil {
				fmt.Printf("failed to get managers version: %v\n", err)
				return
			}

			results <- Module{
				Name: mt.Spec.ModuleName,
				InstallDetails: ModuleInstallDetails{
					Channel:              "",
					Managed:              ManagedFalse,
					CustomResourcePolicy: "N/A",
					Version:              version,
					ModuleState:          moduleStatus,
					InstallationState:    installationStatus,
				},
				Origin:          mt.Namespace + "/" + mt.Name,
				CommunityModule: true,
			}
		}(moduleTemplate)
	}

	wg.Wait()
	close(results)

	communityModules := ModulesList{}
	for m := range results {
		communityModules = append(communityModules, m)
	}
	return communityModules, nil
}

func moduleAlreadyInstalledAsCoreModule(installedCoreModules ModulesList, moduleTemplate kyma.ModuleTemplate) bool {
	for _, installedModule := range installedCoreModules {
		if installedModule.Name == moduleTemplate.Spec.ModuleName {
			return true
		}
	}

	return false
}

func getManagerStatus(installedManager *unstructured.Unstructured) string {
	status, ok := installedManager.Object["status"].(map[string]any)
	if !ok {
		return UnknownValue
	}

	if conditions, ok := status["conditions"]; ok {
		state := getStateFromConditions(conditions.([]any))
		if state != "" {
			return state
		}
	}

	if readyReplicas, ok := status["readyReplicas"]; ok {
		spec := installedManager.Object["spec"].(map[string]any)
		if wantedReplicas, ok := spec["replicas"]; ok {
			state := resolveStateFromReplicas(readyReplicas.(int64), wantedReplicas.(int64))
			if state != "" {
				return state
			}
		}
	}

	return UnknownValue
}

func getManagerVersion(installedManager *unstructured.Unstructured) (string, error) {
	resMetadata, ok := installedManager.Object["metadata"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("metadata not found in unstructured object")
	}

	version := extractModuleVersion(resMetadata, installedManager)
	return version, nil
}

func extractModuleVersion(metadata map[string]any, installedManager *unstructured.Unstructured) string {
	labels, _ := metadata["labels"].(map[string]any)
	if labels != nil {
		if v, ok := labels["app.kubernetes.io/version"].(string); ok && v != "" {
			return v
		}
	}
	spec, ok := installedManager.Object["spec"].(map[string]any)
	if !ok {
		return UnknownValue
	}
	template, ok := spec["template"].(map[string]any)
	if !ok {
		return UnknownValue
	}
	templateSpec, ok := template["spec"].(map[string]any)
	if !ok {
		return UnknownValue
	}
	containers, ok := templateSpec["containers"].([]any)
	if !ok {
		return UnknownValue
	}
	for _, c := range containers {
		container, _ := c.(map[string]any)
		if cname, ok := container["name"].(string); ok && strings.Contains(cname, "manager") {
			if image, ok := container["image"].(string); ok && image != "" {
				parts := strings.Split(image, ":")
				if len(parts) > 1 {
					return parts[len(parts)-1]
				}
			}
		}
	}
	return UnknownValue
}

func getModuleStatus(ctx context.Context, client kube.Client, data unstructured.Unstructured) string {
	apiVersion, ok := data.Object["apiVersion"].(string)
	if !ok {
		fmt.Println("failed to get apiVersion from data: ", data)
		return UnknownValue
	}

	kind, ok := data.Object["kind"].(string)
	if !ok {
		fmt.Println("failed to get kind from data: ", data)
		return UnknownValue
	}

	resourceList, err := listResourcesByVersionKind(ctx, client, apiVersion, kind)
	if err != nil {
		fmt.Printf("failed to list resources for version: %s and kind %s: %v\n", apiVersion, kind, err)
		return UnknownValue
	}

	return determineModuleStatus(resourceList)
}

func listResourcesByVersionKind(ctx context.Context, client kube.Client, apiVersion, kind string) ([]unstructured.Unstructured, error) {
	resourceList, err := client.RootlessDynamic().List(ctx, &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": apiVersion,
			"kind":       kind,
		},
	}, &rootlessdynamic.ListOptions{
		AllNamespaces: true,
	})
	if err != nil {
		return nil, err
	}

	return resourceList.Items, nil
}

func determineModuleStatus(resources []unstructured.Unstructured) string {
	switch len(resources) {
	case 0:
		return NotRunningValue
	case 1:
		item := resources[0]
		if statusMap, ok := item.Object["status"].(map[string]any); ok {
			if state, ok := statusMap["state"].(string); ok {
				return state
			}
		}
		return UnknownValue
	default:
		return UnknownValue
	}
}

func isCommunityModule(moduleTemplate *kyma.ModuleTemplate) bool {
	managedBy, exist := moduleTemplate.ObjectMeta.Labels["operator.kyma-project.io/managed-by"]
	return !exist || managedBy != "kyma"
}

func getManaged(moduleSpec *kyma.Module) Managed {
	if moduleSpec != nil && moduleSpec.Managed != nil {
		return Managed(strconv.FormatBool(*moduleSpec.Managed))
	}

	// default value
	return "true"
}

func getCustomResourcePolicy(moduleSpec *kyma.Module) string {
	if moduleSpec != nil && moduleSpec.CustomResourcePolicy != "" {
		return moduleSpec.CustomResourcePolicy
	}

	// default value
	return "CreateAndDelete"
}

func getModuleInstallationState(ctx context.Context, client kube.Client, moduleStatus kyma.ModuleStatus, moduleSpec *kyma.Module) (string, error) {
	if moduleSpec == nil {
		// module is under deletion
		return moduleStatus.State, nil
	}

	if moduleSpec.CustomResourcePolicy == "CreateAndDelete" {
		// module CR is managed by klm
		return moduleStatus.State, nil
	}

	if moduleSpec.Managed != nil && !*moduleSpec.Managed {
		// module is unmanaged
		return moduleStatus.State, nil
	}

	// TODO: cover case when policy is set to Ingore and CR is not on the cluster

	moduleTemplate, err := client.Kyma().GetModuleTemplate(ctx, moduleStatus.Template.GetNamespace(), moduleStatus.Template.GetName())
	if err != nil {
		return "", errors.Wrapf(err, "failed to get ModuleTemplate %s/%s", moduleStatus.Template.GetNamespace(), moduleStatus.Template.GetName())
	}

	return getResourceState(ctx, client, moduleTemplate.Spec.Manager)
}

func getModuleCustomResourceStatus(ctx context.Context, client kube.Client, moduleStatus kyma.ModuleStatus, moduleSpec *kyma.Module) (string, error) {
	if moduleSpec == nil {
		// module is under deletion = module cr is under deletion
		return moduleStatus.State, nil
	}

	var moduleTemplate *kyma.ModuleTemplate
	var err error

	if moduleSpec.Managed != nil && !*moduleSpec.Managed {
		moduleTemplate, err = findMatchingModuleTemplate(ctx, client, moduleStatus)
		if err != nil {
			return "", errors.Wrapf(err, "failed to get ModuleTemplate for module %s", moduleStatus.Name)

		}
	} else {
		moduleTemplate, err = client.Kyma().GetModuleTemplate(ctx, moduleStatus.Template.GetNamespace(), moduleStatus.Template.GetName())
		if err != nil {
			if apierrors.IsNotFound(err) {
				return UnknownValue, nil
			}
			return "", errors.Wrapf(err, "failed to get ModuleTemplate %s/%s", moduleStatus.Template.GetNamespace(), moduleStatus.Template.GetName())
		}
	}

	state, err := getStateFromData(ctx, client, moduleTemplate.Spec.Data)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return NotRunningValue, nil
		}
		return "", err
	}

	if state != "" {
		return state, nil
	}

	return UnknownValue, nil
}

func getKymaModuleSpec(kymaCR *kyma.Kyma, moduleName string) *kyma.Module {
	for _, module := range kymaCR.Spec.Modules {
		if module.Name == moduleName {
			return &module
		}
	}

	return nil
}

func getStateFromData(ctx context.Context, client kube.Client, data unstructured.Unstructured) (string, error) {
	if len(data.Object) == 0 {
		return "", nil
	}
	namespace := "kyma-system"
	metadata := data.Object["metadata"].(map[string]interface{})
	if ns, ok := metadata["namespace"]; ok && ns.(string) != "" {
		namespace = metadata["namespace"].(string)
	}

	apiVersion := data.Object["apiVersion"].(string)
	kind := data.Object["kind"].(string)
	name := metadata["name"].(string)

	unstruct := generateUnstruct(apiVersion, kind, name, namespace)
	result, err := client.RootlessDynamic().Get(ctx, &unstruct)
	if err != nil {
		return "", err
	}
	statusRaw, ok := result.Object["status"]
	if !ok || statusRaw == nil {
		return "", nil
	}
	status := statusRaw.(map[string]any)
	if state, ok := status["state"]; ok {
		return state.(string), nil
	}
	return "", nil
}

func getResourceState(ctx context.Context, client kube.Client, manager *kyma.Manager) (string, error) {
	if manager == nil {
		return "", nil
	}
	namespace := "kyma-system"
	if manager.Namespace != "" {
		namespace = manager.Namespace
	}

	apiVersion := fmt.Sprintf("%s/%s", manager.Group, manager.Version)

	unstruct := generateUnstruct(apiVersion, manager.Kind, manager.Name, namespace)
	result, err := client.RootlessDynamic().Get(ctx, &unstruct)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}

	statusRaw, ok := result.Object["status"]
	if !ok || statusRaw == nil {
		return "", nil
	}
	status := statusRaw.(map[string]any)
	if state, ok := status["state"]; ok {
		return state.(string), nil
	}

	if conditions, ok := status["conditions"]; ok {
		state := getStateFromConditions(conditions.([]any))
		if state != "" {
			return state, nil
		}
	}
	//check if readyreplicas and wantedreplicas exist
	if readyReplicas, ok := status["readyReplicas"]; ok {
		spec := result.Object["spec"].(map[string]any)
		if wantedReplicas, ok := spec["replicas"]; ok {
			state := resolveStateFromReplicas(readyReplicas.(int64), wantedReplicas.(int64))
			if state != "" {
				return state, nil
			}
		}
	}

	return "", nil
}

func generateUnstruct(apiVersion, kind, name, namespace string) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
		},
	}
}

func resolveStateFromReplicas(ready, wanted int64) string {
	if ready == wanted {
		return "Ready"
	}
	if ready < wanted {
		return "Processing"
	}
	// ready > wanted
	return "Deleting"
}

func getStateFromConditions(conditions []interface{}) string {
	for _, condition := range conditions {
		conditionUnwrapped := condition.(map[string]interface{})
		if conditionUnwrapped["status"] != "True" {
			continue
		}

		conditionType := conditionUnwrapped["type"].(string)

		switch conditionType {
		case "Available":
			return "Ready"
		case "Processing", "Error", "Warning":
			return conditionType
		}
	}
	return ""
}

// look for channel assigned to version with specified moduleName
func getAssignedChannel(releaseMetas kyma.ModuleReleaseMetaList, moduleName, version string) string {
	for _, releaseMeta := range releaseMetas.Items {
		if releaseMeta.Spec.ModuleName == moduleName {
			return getChannelFromAssignments(releaseMeta.Spec.Channels, version)
		}
	}
	return ""
}

func getChannelFromAssignments(assignments []kyma.ChannelVersionAssignment, version string) string {
	for _, assignment := range assignments {
		if assignment.Version == version {
			return assignment.Channel
		}
	}

	return ""
}

// return index of module with given name. if not exists return -1
func getModuleIndex(list ModulesList, name string, isCommunityModule bool) int {
	for i := range list {
		if list[i].Name == name && list[i].CommunityModule == isCommunityModule {
			return i
		}
	}

	return -1
}

func findMatchingModuleTemplate(ctx context.Context, client kube.Client, moduleStatus kyma.ModuleStatus) (*kyma.ModuleTemplate, error) {
	moduleTemplates, err := client.Kyma().ListModuleTemplate(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list modules available on the target Kyma environment")
	}
	for _, mt := range moduleTemplates.Items {
		if mt.Spec.ModuleName == moduleStatus.Name && mt.Spec.Version == moduleStatus.Version {
			return &mt, nil
		}
	}

	return nil, errors.Errorf("no matching module version found for module: %s, version: %s", moduleStatus.Name, moduleStatus.Version)
}
