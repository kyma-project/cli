package modules

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/kyma-project/cli.v3/internal/output"
	"github.com/kyma-project/cli.v3/internal/render"
	"gopkg.in/yaml.v3"
)

type RowConverter func(Module) []interface{}

type TableInfo struct {
	Headers      []interface{}
	RowConverter RowConverter
}

var (
	ModulesTableInfo = TableInfo{
		Headers: []interface{}{"NAME", "VERSION", "CR POLICY", "MANAGED", "STATUS"},
		RowConverter: func(m Module) []interface{} {
			return []interface{}{
				m.Name,
				convertInstall(m.InstallDetails),
				string(m.InstallDetails.CustomResourcePolicy),
				string(m.InstallDetails.Managed),
				m.InstallDetails.State,
			}
		},
	}

	CatalogTableInfo = TableInfo{
		Headers: []interface{}{"NAME", "AVAILABLE VERSIONS", "COMMUNITY"},
		RowConverter: func(m Module) []interface{} {
			return []interface{}{
				m.Name,
				convertVersions(m.Versions),
				m.CommunityModule,
			}
		},
	}
)

// Renders uses standard output to print ModuleList in table view
func Render(modulesList ModulesList, tableInfo TableInfo, format output.Format) error {
	switch format {
	case output.JSONFormat:
		return renderJSON(os.Stdout, modulesList, tableInfo)
	case output.YAMLFormat:
		return renderYAML(os.Stdout, modulesList, tableInfo)
	default:
		return renderTable(os.Stdout, modulesList, tableInfo)
	}
}

func renderJSON(writer io.Writer, modulesList ModulesList, tableInfo TableInfo) error {
	obj, err := json.MarshalIndent(convertToOutputParameters(modulesList, tableInfo), "", "  ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(writer, string(obj))
	return err
}

func renderYAML(writer io.Writer, modulesList ModulesList, tableInfo TableInfo) error {
	obj, err := yaml.Marshal(convertToOutputParameters(modulesList, tableInfo))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(writer, string(obj))
	return err
}

func renderTable(writer io.Writer, modulesList ModulesList, tableInfo TableInfo) error {
	render.Table(
		writer,
		tableInfo.Headers,
		convertModuleListToRows(modulesList, tableInfo.RowConverter),
	)
	return nil
}

func convertToOutputParameters(modulesList ModulesList, tableInfo TableInfo) []map[string]interface{} {
	result := make([]map[string]interface{}, len(modulesList))
	for i, resource := range modulesList {
		result[i] = make(map[string]interface{}, len(tableInfo.Headers))
		row := tableInfo.RowConverter(resource)
		for fieldIter, fieldName := range tableInfo.Headers {
			result[i][fieldName.(string)] = row[fieldIter]
		}
	}

	return result
}

func convertModuleListToRows(modulesList ModulesList, rowConverter RowConverter) [][]interface{} {
	sort.Slice(modulesList, func(i, j int) bool {
		if modulesList[i].CommunityModule == modulesList[j].CommunityModule {
			return modulesList[i].Name < modulesList[j].Name
		}
		return !modulesList[i].CommunityModule && modulesList[j].CommunityModule
	})

	var result [][]interface{}
	for _, module := range modulesList {
		result = append(result, rowConverter(module))
	}
	return result
}

// convert version and channel into field in format 'version (channel)' for core modules and 'version' for community ones
func convertInstall(details ModuleInstallDetails) string {
	if details.Channel != "" {
		return fmt.Sprintf("%s(%s)", details.Version, details.Channel)
	}

	return details.Version
}

// convert versions to string containing values separated with '\n'
// and in format 'VERSION (CHANNEL)' or 'VERSION' if channel is empty
func convertVersions(versions []ModuleVersion) string {
	values := make([]string, len(versions))
	for i, version := range versions {
		value := version.Version
		if version.Channel != "" {
			value += fmt.Sprintf("(%s)", version.Channel)
		}

		values[i] = value
	}

	return strings.Join(values, ", ")
}
