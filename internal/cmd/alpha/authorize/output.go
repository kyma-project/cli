package authorize

import (
	"encoding/json"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/out"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func outputResources(outputFormat types.Format, resources []*unstructured.Unstructured) clierror.Error {
	printer := out.Default

	switch outputFormat {
	case types.JSONFormat:
		return outputJSON(printer, resources)
	case types.YAMLFormat, types.DefaultFormat:
		return outputYAML(printer, resources)
	default:
		return outputYAML(printer, resources)
	}
}

func outputJSON(printer *out.Printer, resources []*unstructured.Unstructured) clierror.Error {
	obj, err := json.MarshalIndent(resources, "", "  ")
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to marshal resources to JSON"))
	}

	printer.Msgln(string(obj))
	return nil
}

func outputYAML(printer *out.Printer, resources []*unstructured.Unstructured) clierror.Error {
	resourcesYamls := []string{}

	for _, resource := range resources {
		resourceBytes, err := yaml.Marshal(resource.Object)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to marshal resource to YAML"))
		}

		resourcesYamls = append(resourcesYamls, string(resourceBytes))
	}

	resultYaml := strings.Join(resourcesYamls, "---\n")
	printer.Msgln(resultYaml)

	return nil
}
