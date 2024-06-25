package communitymodules

import (
	"encoding/json"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/olekukonko/tablewriter"
	"io"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net/http"
	"os"
	"strings"
)

const URL = "https://raw.githubusercontent.com/kyma-project/community-modules/main/model.json"

type row []string

type moduleMap map[string]row

// ModulesCatalog returns a map of all available modules and their repositories, if the map is nil it will create a new one
func ModulesCatalog(partialMap moduleMap) (moduleMap, clierror.Error) {
	// TODO: what is "model"? think about this name
	modules, err := getCommunityModules()
	if err != nil {
		return nil, err
	}

	catalog := make(moduleMap)
	if partialMap != nil {
		catalog = partialMap
	}

	for _, rec := range modules {
		if partialMap != nil {
			// TODO: IMO this code doesn't work! We don't read partialMap later.
			partialMap[rec.Name] = append(partialMap[rec.Name], rec.Versions[0].Repository)
		} else {
			catalog[rec.Name] = append(catalog[rec.Name], rec.Name)
			catalog[rec.Name] = append(catalog[rec.Name], rec.Versions[0].Repository)
		}
	}
	return catalog, nil
}

// getCommunityModules returns a list of all available modules from the community-modules repository
func getCommunityModules() (Modules, clierror.Error) {
	resp, err := http.Get(URL)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while getting modules list from github"))
	}
	defer resp.Body.Close()

	var modules Modules
	modules, respErr := decodeCommunityModulesResponse(err, resp, modules)
	if respErr != nil {
		return nil, clierror.WrapE(respErr, clierror.New("while handling response"))
	}
	return modules, nil
}

// decodeCommunityModulesResponse reads the response body and unmarshals it into the template
func decodeCommunityModulesResponse(err error, resp *http.Response, modules Modules) (Modules, clierror.Error) {
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while reading http response"))
	}
	err = json.Unmarshal(bodyText, &modules)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while unmarshalling"))
	}
	return modules, nil
}

// ManagedModules returns a map of all managed modules from the cluster
func ManagedModules(partialMap moduleMap, client cmdcommon.KubeClientConfig, cfg cmdcommon.KymaConfig) (moduleMap, clierror.Error) {

	moduleNames, err := getManagedList(client, cfg)
	if err != nil {
		return nil, clierror.WrapE(err, clierror.New("while getting managed modules"))
	}

	managed := make(moduleMap)
	if partialMap != nil {
		managed = partialMap
	}

	for _, name := range moduleNames {
		if partialMap != nil {
			// TODO: IMO this code doesn't work! We don't read partialMap later.
			partialMap[name] = append(partialMap[name], "Managed")
		} else {
			managed[name] = append(managed[name], name)
		}
	}

	return managed, nil
}

// getManagedList gets a list of all managed modules from the Kyma CR
func getManagedList(client cmdcommon.KubeClientConfig, cfg cmdcommon.KymaConfig) ([]string, clierror.Error) {
	GVRKyma := schema.GroupVersionResource{
		Group:    "operator.kyma-project.io",
		Version:  "v1beta2",
		Resource: "kymas",
	}

	resp, err := client.KubeClient.Dynamic().Resource(GVRKyma).Namespace("kyma-system").
		Get(cfg.Ctx, "default", metav1.GetOptions{})
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while getting Kyma CR"))
	}

	moduleNames, err := decodeKymaCRResponse(resp)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while getting module names from CR"))
	}
	return moduleNames, nil
}

// decodeKymaCRResponse interprets the response and returns a list of managed modules names
func decodeKymaCRResponse(unstruct *unstructured.Unstructured) ([]string, error) {
	var moduleNames []string
	managedFields := unstruct.GetManagedFields()
	for _, field := range managedFields {
		var data map[string]interface{}
		err := json.Unmarshal(field.FieldsV1.Raw, &data)
		if err != nil {
			return nil, err
		}

		spec, ok := data["f:spec"].(map[string]interface{})
		if !ok {
			continue
		}

		modules, ok := spec["f:modules"].(map[string]interface{})
		if !ok {
			continue
		}

		moduleNames = append(moduleNames, extractNames(modules)...)
	}
	return moduleNames, nil
}

func extractNames(modules map[string]interface{}) []string {
	var moduleNames []string
	for key := range modules {
		if strings.Contains(key, "name") {
			name := strings.TrimPrefix(key, "k:{\"name\":\"")
			name = strings.Trim(name, "\"}")
			moduleNames = append(moduleNames, name)
		}
	}
	return moduleNames
}

// InstalledModules returns a map of all installed modules from the cluster, regardless whether they are managed or not
func InstalledModules(partialMap moduleMap, client cmdcommon.KubeClientConfig, cfg cmdcommon.KymaConfig) (moduleMap, clierror.Error) {
	modules, err := getCommunityModules()
	if err != nil {
		return nil, clierror.WrapE(err, clierror.New("while getting installed modules"))
	}

	installed := make(moduleMap)
	if partialMap != nil {
		installed = partialMap
	}

	installed, err = getInstalledModules(partialMap, installed, modules, client, cfg)
	if err != nil {
		return nil, err
	}

	return installed, nil
}

func getInstalledModules(partialMap, installed moduleMap, modules Modules, client cmdcommon.KubeClientConfig, cfg cmdcommon.KymaConfig) (moduleMap, clierror.Error) {
	for _, module := range modules {
		managerName := getManagerName(module)
		deployment, err := client.KubeClient.Static().AppsV1().Deployments("kyma-system").
			Get(cfg.Ctx, managerName, metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			msg := "while getting the " + managerName + " deployment"
			return nil, clierror.Wrap(err, clierror.New(msg))
		}
		if !errors.IsNotFound(err) {
			installedVersion := getInstalledVersion(deployment)
			moduleVersion := module.Versions[0].Version
			version := calculateVersion(moduleVersion, installedVersion)
			addVersionColumn(module.Name, version, partialMap, installed)
		}
	}
	return installed, nil
}

func getInstalledVersion(deployment *v1.Deployment) string {
	deploymentImage := strings.Split(deployment.Spec.Template.Spec.Containers[0].Image, "/")
	nameAndTag := strings.Split(deploymentImage[len(deploymentImage)-1], ":")
	return nameAndTag[len(nameAndTag)-1]
}

func getManagerName(module Module) string {
	managerPath := strings.Split(module.Versions[0].ManagerPath, "/")
	return managerPath[len(managerPath)-1]
}

func addVersionColumn(name, version string, partialMap, installed moduleMap) {
	if partialMap == nil {
		installed[name] = append(installed[name], name)
	}
	installed[name] = append(installed[name], version)
}

func calculateVersion(moduleVersion string, installedVersion string) string {
	if moduleVersion == installedVersion {
		return installedVersion
	}
	return "outdated moduleVersion, latest is " + moduleVersion
}

// renderTable renders the table with the provided headers
func RenderTable(raw bool, modulesMap moduleMap, headers row) {
	if raw {
		for _, row := range modulesMap {
			println(strings.Join(row, "\t"))
		}
	} else {

		var table [][]string
		for _, row := range modulesMap {
			table = append(table, row)
		}

		twTable := setTable(table)
		twTable.SetHeader(headers)
		twTable.Render()
	}
}

// setTable sets the table settings for the tablewriter
func setTable(inTable [][]string) *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.AppendBulk(inTable)
	table.SetRowLine(true)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_CENTER, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
	table.SetBorder(false)
	return table
}
