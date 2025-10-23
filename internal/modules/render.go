package modules

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
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
		Headers: []interface{}{"NAME", "VERSION", "CR POLICY", "MANAGED", "MODULE STATUS", "INSTALLATION STATUS"},
		RowConverter: func(m Module) []interface{} {
			return []interface{}{
				m.Name,
				convertInstall(m.InstallDetails),
				string(m.InstallDetails.CustomResourcePolicy),
				string(m.InstallDetails.Managed),
				m.InstallDetails.ModuleState,
				m.InstallDetails.InstallationState,
			}
		},
	}

	CatalogTableInfo = TableInfo{
		Headers: []interface{}{"NAME", "AVAILABLE VERSIONS", "ORIGIN"},
		RowConverter: func(m Module) []interface{} {
			return []interface{}{
				m.Name,
				convertVersions(m.Versions),
				m.Origin,
			}
		},
	}
)

// Renders uses standard output to print ModuleList in table view
func Render(modulesList ModulesList, tableInfo TableInfo, format types.Format) error {
	switch format {
	case types.JSONFormat:
		return renderJSON(os.Stdout, modulesList, tableInfo)
	case types.YAMLFormat:
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
			formattedFieldName := toCamelCase(fieldName.(string))
			result[i][formattedFieldName] = row[fieldIter]
		}
	}

	return result
}

func convertModuleListToRows(modulesList ModulesList, rowConverter RowConverter) [][]interface{} {
	sort.Slice(modulesList, func(i, j int) bool {
		// First: Core modules (CommunityModule == false)
		if !modulesList[i].CommunityModule && modulesList[j].CommunityModule {
			return true
		}
		if modulesList[i].CommunityModule && !modulesList[j].CommunityModule {
			return false
		}

		// Both are community modules, sort by origin
		if modulesList[i].CommunityModule && modulesList[j].CommunityModule {
			// Second: Community modules with origin != "community"
			if modulesList[i].Origin != OriginCommunity && modulesList[j].Origin == OriginCommunity {
				return true
			}
			if modulesList[i].Origin == OriginCommunity && modulesList[j].Origin != OriginCommunity {
				return false
			}
		}

		// Within the same category, sort by name
		return modulesList[i].Name < modulesList[j].Name
	})

	var result [][]any
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

func toCamelCase(s string) string {
	words := strings.Fields(strings.ToLower(s))
	if len(words) == 0 {
		return ""
	}
	camel := words[0]
	for _, w := range words[1:] {
		if len(w) > 0 {
			camel += strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return camel
}
