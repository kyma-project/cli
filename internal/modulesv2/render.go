package modulesv2

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/kyma-project/cli.v3/internal/render"
	"gopkg.in/yaml.v3"
)

func RenderCatalog(results []dtos.CatalogResult, format types.Format) error {
	switch format {
	case types.JSONFormat:
		return renderCatalogJSON(results)
	case types.YAMLFormat:
		return renderCatalogYAML(results)
	default:
		return renderCatalogTable(results)
	}
}

func renderCatalogJSON(results []dtos.CatalogResult) error {
	output := convertToOutputFormat(results)
	obj, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	out.Default.Msgln(string(obj))
	return nil
}

func renderCatalogYAML(results []dtos.CatalogResult) error {
	output := convertToOutputFormat(results)
	obj, err := yaml.Marshal(output)
	if err != nil {
		return err
	}

	out.Default.Msgln(string(obj))
	return nil
}

func renderCatalogTable(results []dtos.CatalogResult) error {
	sortCatalogResults(results)

	headers := []interface{}{"NAME", "AVAILABLE VERSIONS", "ORIGIN"}
	rows := convertCatalogToRows(results)

	render.Table(out.Default, headers, rows)
	return nil
}

func convertToOutputFormat(results []dtos.CatalogResult) []map[string]interface{} {
	output := make([]map[string]interface{}, len(results))
	for i, result := range results {
		output[i] = map[string]interface{}{
			"name":              result.Name,
			"availableVersions": result.AvailableVersions,
			"origin":            result.Origin,
		}
	}
	return output
}

func convertCatalogToRows(results []dtos.CatalogResult) [][]interface{} {
	rows := make([][]interface{}, len(results))
	for i, result := range results {
		rows[i] = []interface{}{
			result.Name,
			strings.Join(result.AvailableVersions, ", "),
			result.Origin,
		}
	}
	return rows
}

func sortCatalogResults(results []dtos.CatalogResult) {
	sort.Slice(results, func(i, j int) bool {
		// First: kyma origin modules
		if results[i].Origin == dtos.KYMA_ORIGIN && results[j].Origin != dtos.KYMA_ORIGIN {
			return true
		}
		if results[i].Origin != dtos.KYMA_ORIGIN && results[j].Origin == dtos.KYMA_ORIGIN {
			return false
		}

		// Second: community origin modules
		if results[i].Origin != dtos.COMMUNITY_ORIGIN && results[j].Origin == dtos.COMMUNITY_ORIGIN {
			return false
		}
		if results[i].Origin == dtos.COMMUNITY_ORIGIN && results[j].Origin != dtos.COMMUNITY_ORIGIN {
			return true
		}

		// Within the same category, sort by name
		return results[i].Name < results[j].Name
	})
}
