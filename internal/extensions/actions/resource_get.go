package actions

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"slices"

	"github.com/itchyny/gojq"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/extensions/actions/common"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	"github.com/kyma-project/cli.v3/internal/output"
	"github.com/kyma-project/cli.v3/internal/render"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type resourceGetActionConfig struct {
	OutputFormat      output.Format          `yaml:"output"`
	FromAllNamespaces bool                   `yaml:"fromAllNamespaces"`
	Resource          map[string]interface{} `yaml:"resource"`
	OutputParameters  []outputParameter      `yaml:"outputParameters"`
}

type outputParameter struct {
	Name         string `yaml:"name"`
	ResourcePath string `yaml:"resourcePath"`
}

type resourceGetAction struct {
	common.TemplateConfigurator[resourceGetActionConfig]

	kymaConfig *cmdcommon.KymaConfig
}

func NewResourceGet(kymaConfig *cmdcommon.KymaConfig) types.Action {
	return &resourceGetAction{
		kymaConfig: kymaConfig,
	}
}

func (a *resourceGetAction) Run(cmd *cobra.Command, _ []string) clierror.Error {
	u := &unstructured.Unstructured{
		Object: a.Cfg.Resource,
	}

	client, clierr := a.kymaConfig.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	nameSelector := ""
	if u.GetName() != "" {
		// set name field selector to get only one resource
		nameSelector = fmt.Sprintf("metadata.name==%s", u.GetName())
	}

	resources, err := client.RootlessDynamic().List(a.kymaConfig.Ctx, u, &rootlessdynamic.ListOptions{
		AllNamespaces: a.Cfg.FromAllNamespaces,
		FieldSelector: nameSelector,
	})
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to get resource"))
	}

	output, err := a.formatOutput(resources)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to format output"))
	}

	fmt.Fprint(cmd.OutOrStdout(), output)
	return nil
}

func (a *resourceGetAction) formatOutput(resources *unstructured.UnstructuredList) (string, error) {
	tableInfo := buildTableInfo(&a.Cfg)
	outputParameters := convertResourcesToParameters(resources.Items, tableInfo)

	if a.Cfg.OutputFormat == output.JSONFormat {
		obj, err := json.MarshalIndent(outputParameters, "", "  ")
		return string(obj), err
	}

	if a.Cfg.OutputFormat == output.YAMLFormat {
		obj, err := yaml.Marshal(outputParameters)
		return string(obj), err
	}

	buf := bytes.NewBuffer([]byte{})
	renderTable(buf, resources.Items, tableInfo)
	return buf.String(), nil
}

func buildTableInfo(cfg *resourceGetActionConfig) TableInfo {
	Headers := []interface{}{}
	fieldConverters := []FieldConverter{}

	if cfg.FromAllNamespaces {
		Headers = append(Headers, "namespace")
		fieldConverters = append(fieldConverters, genericFieldConverter(".metadata.namespace"))
	}

	Headers = append(Headers, "namespace")
	fieldConverters = append(fieldConverters, genericFieldConverter(".metadata.name"))

	for _, param := range cfg.OutputParameters {
		Headers = append(Headers, param.Name)
		fieldConverters = append(fieldConverters, genericFieldConverter(param.ResourcePath))
	}

	return TableInfo{
		Headers: Headers,
		RowConverter: func(u unstructured.Unstructured) []interface{} {
			row := make([]interface{}, len(fieldConverters))
			for i := range fieldConverters {
				row[i] = fieldConverters[i](u)
			}

			return row
		},
	}
}

func genericFieldConverter(path string) func(u unstructured.Unstructured) string {
	return func(u unstructured.Unstructured) string {
		query, err := gojq.Parse(path)
		if err != nil {
			// ignore result because path is incorrect
			return ""
		}

		value, ok := query.Run(u.Object).Next()
		_, isError := value.(error)
		if !ok || isError {
			// ignore result because of an unexpected error
			return ""
		}

		return fmt.Sprintf("%v", value)
	}
}

func renderTable(writer io.Writer, resources []unstructured.Unstructured, tableInfo TableInfo) {
	render.Table(
		writer,
		tableInfo.Headers,
		convertResourcesToTable(resources, tableInfo.RowConverter),
	)
}

type FieldConverter func(u unstructured.Unstructured) string

type RowConverter func(unstructured.Unstructured) []interface{}

type TableInfo struct {
	Headers      []interface{}
	RowConverter RowConverter
}

func convertResourcesToTable(resources []unstructured.Unstructured, rowConverter RowConverter) [][]interface{} {
	slices.SortFunc(resources, func(a, b unstructured.Unstructured) int {
		return cmp.Compare(a.GetNamespace(), b.GetNamespace())
	})

	var result [][]interface{}
	for _, resource := range resources {
		result = append(result, rowConverter(resource))
	}
	return result
}

func convertResourcesToParameters(resources []unstructured.Unstructured, tableInfo TableInfo) []map[string]interface{} {
	result := make([]map[string]interface{}, len(resources))
	for i, resource := range resources {
		result[i] = make(map[string]interface{}, len(tableInfo.Headers))
		row := tableInfo.RowConverter(resource)
		for fieldIter, fieldName := range tableInfo.Headers {
			result[i][fieldName.(string)] = row[fieldIter]
		}
	}

	return result
}
