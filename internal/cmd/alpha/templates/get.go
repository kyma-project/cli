package templates

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/types"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	"github.com/kyma-project/cli.v3/internal/render"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type GetOptions struct {
	types.GetCommand
	ResourceInfo types.ResourceInfo
	RootCommand  types.RootCommand
}

func BuildGetCommand(clientGetter KubeClientGetter, options *GetOptions) *cobra.Command {
	return buildGetCommand(os.Stdout, clientGetter, options)
}

func buildGetCommand(out io.Writer, clientGetter KubeClientGetter, options *GetOptions) *cobra.Command {
	flags := flags{}
	cmd := &cobra.Command{
		Use:     "get [<resource_name>]",
		Example: buildGetExample(options),
		Short:   options.Description,
		Long:    options.DescriptionLong,
		Args:    AssignOptionalNameArg(&flags.name),
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(getResources(&getArgs{
				out:          out,
				ctx:          cmd.Context(),
				getOptions:   options,
				clientGetter: clientGetter,
				flags:        flags,
			}))
		},
	}

	if options.ResourceInfo.Scope == types.NamespaceScope {
		cmd.Flags().StringVarP(&flags.namespace, "namespace", "n", "default", "resource namespace")
		cmd.Flags().BoolVarP(&flags.allNamespaces, "all-namespaces", "A", false, "get from all namespaces")
	}

	return cmd
}

func buildGetExample(options *GetOptions) string {
	template := "  # list all resources\n" +
		"  kyma alpha ROOT_COMMAND get\n" +
		"\n" +
		"  # get resource with a specific name\n" +
		"  kyma alpha ROOT_COMMAND get <resource_name>"

	if options.ResourceInfo.Scope == types.NamespaceScope {
		template += "\n\n" +
			"  # list all resources from the specific namespaces\n" +
			"  kyma alpha ROOT_COMMAND get -n namespace_name\n" +
			"\n" +
			"  # list all resources from all namespaces\n" +
			"  kyma alpha ROOT_COMMAND get -A"
	}

	return strings.ReplaceAll(template, "ROOT_COMMAND", options.RootCommand.Name)
}

type getArgs struct {
	out          io.Writer
	ctx          context.Context
	getOptions   *GetOptions
	clientGetter KubeClientGetter
	flags        flags
}

type flags struct {
	allNamespaces bool
	name          string
	namespace     string
}

func getResources(args *getArgs) clierror.Error {
	u := &unstructured.Unstructured{}
	u.SetNamespace(args.flags.namespace)
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   args.getOptions.ResourceInfo.Group,
		Version: args.getOptions.ResourceInfo.Version,
		Kind:    args.getOptions.ResourceInfo.Kind,
	})

	client, clierr := args.clientGetter.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	nameSelector := ""
	if args.flags.name != "" {
		// set name field selector to get only one resource
		nameSelector = fmt.Sprintf("metadata.name==%s", args.flags.name)
	}

	resources, err := client.RootlessDynamic().List(args.ctx, u, &rootlessdynamic.ListOptions{
		AllNamespaces: args.flags.allNamespaces,
		FieldSelector: nameSelector,
	})
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to get resource"))
	}

	tableInfo := buildTableInfo(args)
	renderTable(args.out, resources.Items, tableInfo)
	return nil
}

func buildTableInfo(opts *getArgs) TableInfo {
	Headers := []interface{}{}
	fieldConverters := []FieldConverter{}

	if opts.flags.allNamespaces {
		Headers = append(Headers, "NAMESPACE")
		fieldConverters = append(fieldConverters, genericFieldConverter(".metadata.namespace"))
	}

	Headers = append(Headers, "NAME")
	fieldConverters = append(fieldConverters, genericFieldConverter(".metadata.name"))

	for _, param := range opts.getOptions.Parameters {
		Headers = append(Headers, strings.ToUpper(param.Name))
		fieldConverters = append(fieldConverters, genericFieldConverter(param.Path))
	}

	return TableInfo{
		Header: Headers,
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
		tableInfo.Header,
		convertResourcesToTable(resources, tableInfo.RowConverter),
	)
}

type FieldConverter func(u unstructured.Unstructured) string

type RowConverter func(unstructured.Unstructured) []interface{}

type TableInfo struct {
	Header       []interface{}
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
