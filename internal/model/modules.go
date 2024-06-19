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

func GetAllModules(moduleMap map[string][]string) (map[string][]string, clierror.Error) {
	if moduleMap != nil {
		template, err := getCatalog()
		if err != nil {
			return nil, clierror.WrapE(err, clierror.New("while trying to get module catalog"))
		}
		for _, rec := range template {
			moduleMap[rec.Name] = append(moduleMap[rec.Name], rec.Versions[0].Repository)
		}
		return moduleMap, nil
	}
	template, err := getCatalog()
	if err != nil {
		return nil, clierror.WrapE(err, clierror.New("while trying to get module catalog"))
	}

	modules := make(map[string][]string)

	for _, rec := range template {
		modules[rec.Name] = append(modules[rec.Name], rec.Name)
		modules[rec.Name] = append(modules[rec.Name], rec.Versions[0].Repository)
	}

	return modules, nil
}

func getCatalog() (Module, clierror.Error) {
	resp, err := http.Get(URL)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while getting modules list from github"))
	}
	defer resp.Body.Close()

	var template Module

	template, respErr := handleResponse(err, resp, template)
	if respErr != nil {
		return nil, clierror.WrapE(respErr, clierror.New("while handling response"))
	}
	return template, nil
}

func GetManagedModules(moduleMap map[string][]string, client cmdcommon.KubeClientConfig, cfg cmdcommon.KymaConfig) (map[string][]string, clierror.Error) {
	if moduleMap != nil {
		name, err := getManaged(client, cfg)
		if err != nil {
			return nil, clierror.WrapE(err, clierror.New("while getting managed modules"))
		}
		for _, rec := range name {
			moduleMap[rec] = append(moduleMap[rec], "Managed")
		}
		return moduleMap, nil
	}
	name, err := getManaged(client, cfg)
	if err != nil {
		return nil, clierror.WrapE(err, clierror.New("while getting managed modules"))
	}

	managed := make(map[string][]string)

	for _, rec := range name {
		managed[rec] = append(managed[rec], rec)
	}

	return managed, nil
}

func getManaged(client cmdcommon.KubeClientConfig, cfg cmdcommon.KymaConfig) ([]string, clierror.Error) {
	GVRKyma := schema.GroupVersionResource{
		Group:    "operator.kyma-project.io",
		Version:  "v1beta2",
		Resource: "kymas",
	}

	unstruct, err := client.KubeClient.Dynamic().Resource(GVRKyma).Namespace("kyma-system").Get(cfg.Ctx, "default", metav1.GetOptions{})
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while getting Kyma CR"))
	}

	name, err := getModuleNames(unstruct)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while getting module names from CR"))
	}
	return name, nil
}

func SetTable(inTable [][]string) *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.AppendBulk(inTable)
	table.SetRowLine(true)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetBorder(false)
	return table

}

func GetInstalledModules(moduleMap map[string][]string, client cmdcommon.KubeClientConfig, cfg cmdcommon.KymaConfig) (map[string][]string, clierror.Error) {
	if moduleMap != nil {
		template, err := getInstalled()
		if err != nil {
			return nil, clierror.WrapE(err, clierror.New("while getting installed modules"))
		}
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
				if version == installedVersion[len(installedVersion)-1] {
					moduleMap[rec.Name] = append(moduleMap[rec.Name], installedVersion[len(installedVersion)-1])
				} else {
					moduleMap[rec.Name] = append(moduleMap[rec.Name], "outdated version,\n latest is "+version)
				}
			}
		}
		return moduleMap, nil
	}

	template, err := getInstalled()
	if err != nil {
		return nil, clierror.WrapE(err, clierror.New("while getting installed modules"))
	}

	installed := make(map[string][]string)

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
			if version == installedVersion[len(installedVersion)-1] {
				installed[rec.Name] = append(installed[rec.Name], rec.Name)
				installed[rec.Name] = append(installed[rec.Name], installedVersion[len(installedVersion)-1])
			} else {
				installed[rec.Name] = append(installed[rec.Name], rec.Name)
				installed[rec.Name] = append(installed[rec.Name], "outdated version, latest is "+version)
			}
		}
	}
	return installed, nil
}

func getInstalled() (Module, clierror.Error) {
	resp, err := http.Get(URL)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while getting modules list from github"))
	}
	defer resp.Body.Close()

	var template Module

	template, respErr := handleResponse(err, resp, template)
	if respErr != nil {
		return nil, clierror.WrapE(respErr, clierror.New("while handling response"))
	}
	return template, nil
}

func handleResponse(err error, resp *http.Response, template Module) (Module, clierror.Error) {
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

func getModuleNames(unstruct *unstructured.Unstructured) ([]string, error) {
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

		for key := range modules {
			if strings.Contains(key, "name") {
				name := strings.TrimPrefix(key, "k:{\"name\":\"")
				name = strings.Trim(name, "\"}")
				moduleNames = append(moduleNames, name)
			}
		}
	}
	return moduleNames, nil
}
