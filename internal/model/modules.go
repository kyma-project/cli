package model

import (
	"encoding/json"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/olekukonko/tablewriter"
	"io"
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
func ModulesCatalog(modulesMap moduleMap) (moduleMap, clierror.Error) {

	template, err := getModel()
	if err != nil {
		return nil, err
	}

	catalog := make(moduleMap)
	if modulesMap != nil {
		catalog = modulesMap
	}

	for _, rec := range template {
		if modulesMap != nil {
			modulesMap[rec.Name] = append(modulesMap[rec.Name], rec.Versions[0].Repository)
		} else {
			catalog[rec.Name] = append(catalog[rec.Name], rec.Name)
			catalog[rec.Name] = append(catalog[rec.Name], rec.Versions[0].Repository)
		}
	}
	return catalog, nil
}

// getModel returns a list of all available modules from the community-modules repository
func getModel() (Module, clierror.Error) {
	resp, err := http.Get(URL)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while getting modules list from github"))
	}
	defer resp.Body.Close()

	var template Module
	template, respErr := handleHTTPResponse(err, resp, template)
	if respErr != nil {
		return nil, clierror.WrapE(respErr, clierror.New("while handling response"))
	}
	return template, nil
}

// ManagedModules returns a map of all managed modules from the cluster
func ManagedModules(modulesMap moduleMap, client cmdcommon.KubeClientConfig, cfg cmdcommon.KymaConfig) (moduleMap, clierror.Error) {

	name, err := getManagedList(client, cfg)
	if err != nil {
		return nil, clierror.WrapE(err, clierror.New("while getting managed modules"))
	}

	managed := make(moduleMap)
	if modulesMap != nil {
		managed = modulesMap
	}

	for _, rec := range name {
		if modulesMap != nil {
			modulesMap[rec] = append(modulesMap[rec], "Managed")
		} else {
			managed[rec] = append(managed[rec], rec)
		}
	}

	return managed, nil
}

// getManagedList gets a list of all managed modules from the Kyma CR
func getManagedList(client cmdcommon.KubeClientConfig, cfg cmdcommon.KymaConfig) (row, clierror.Error) {
	GVRKyma := schema.GroupVersionResource{
		Group:    "operator.kyma-project.io",
		Version:  "v1beta2",
		Resource: "kymas",
	}

	unstruct, err := client.KubeClient.Dynamic().Resource(GVRKyma).Namespace("kyma-system").Get(cfg.Ctx, "default", metav1.GetOptions{})
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while getting Kyma CR"))
	}

	name, err := handleClusterResponse(unstruct)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while getting module names from CR"))
	}
	return name, nil
}

// handleClusterResponse interprets the response and returns a list of managed modules names
func handleClusterResponse(unstruct *unstructured.Unstructured) (row, error) {
	var moduleNames row
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

		moduleNames = manageNames(moduleNames, modules)
	}
	return moduleNames, nil
}

func manageNames(moduleNames row, modules map[string]interface{}) row {
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
	template, err := getModel()
	if err != nil {
		return nil, clierror.WrapE(err, clierror.New("while getting installed modules"))
	}

	installed := make(moduleMap)
	if partialMap != nil {
		installed = partialMap
	}

	installed, err = getInstalledModules(partialMap, installed, template, client, cfg)
	if err != nil {
		return nil, err
	}

	return installed, nil
}

func getInstalledModules(moduleMap, installed moduleMap, template Module, client cmdcommon.KubeClientConfig, cfg cmdcommon.KymaConfig) (moduleMap, clierror.Error) {
	for _, rec := range template {
		managerPath := strings.Split(rec.Versions[0].ManagerPath, "/")
		managerName := managerPath[len(managerPath)-1]
		version := rec.Versions[0].Version
		deployment, err := client.KubeClient.Static().AppsV1().Deployments("kyma-system").Get(cfg.Ctx, managerName, metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			msg := "while getting the " + managerName + " deployment"
			return nil, clierror.Wrap(err, clierror.New(msg))
		}
		if !errors.IsNotFound(err) {
			deploymentImage := strings.Split(deployment.Spec.Template.Spec.Containers[0].Image, "/")
			installedVersion := strings.Split(deploymentImage[len(deploymentImage)-1], ":")
			manageVersion(rec.Name, version, installedVersion, moduleMap, installed)
		}
	}
	return installed, nil
}

func manageVersion(name, version string, installedVersion row, moduleMap, installed moduleMap) {
	if version == installedVersion[len(installedVersion)-1] {
		if moduleMap == nil {
			installed[name] = append(installed[name], name)
		}
		installed[name] = append(installed[name], installedVersion[len(installedVersion)-1])

	} else {
		if moduleMap == nil {
			installed[name] = append(installed[name], name)
		}
		installed[name] = append(installed[name], "outdated version, latest is "+version)
	}
}

// handleHTTPResponse reads the response body and unmarshals it into the template
func handleHTTPResponse(err error, resp *http.Response, template Module) (Module, clierror.Error) {
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while reading http response"))
	}
	err = json.Unmarshal(bodyText, &template)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while unmarshalling"))
	}
	return template, nil
}

// renderTable renders the table with the provided headers
func RenderTable(raw bool, modulesMap moduleMap, headers []string) {
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
