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
}

func BuildGetCommand(clientGetter KubeClientGetter, options *GetOptions) *cobra.Command {
	return buildGetCommand(os.Stdout, clientGetter, options)
}

func buildGetCommand(out io.Writer, clientGetter KubeClientGetter, options *GetOptions) *cobra.Command {
	flags := flags{}
	cmd := &cobra.Command{
		Use:   "get",
		Short: options.Description,
		Long:  options.DescriptionLong,
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

	cmd.Flags().StringVar(&flags.name, "name", "", "resource of the resource")

	if options.ResourceInfo.Scope == types.NamespaceScope {
		cmd.Flags().StringVar(&flags.namespace, "namespace", "default", "resource namespace")
		cmd.Flags().BoolVarP(&flags.allNamespaces, "all-namespaces", "A", false, "get from all namespaces")
	}

	return cmd
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

	tableInfo := buildTableInfo(args.getOptions)
	renderTable(args.out, resources.Items, tableInfo)
	return nil
}

func buildTableInfo(opts *GetOptions) TableInfo {
	Headers := []string{"NAME"}
	fieldConverters := []FieldConverter{
		genericFieldConverter(".metadata.name"),
	}

	if opts.ResourceInfo.Scope == types.NamespaceScope {
		Headers = append(Headers, "NAMESPACE")
		fieldConverters = append(fieldConverters, genericFieldConverter(".metadata.namespace"))
	}

	for _, param := range opts.Parameters {
		Headers = append(Headers, strings.ToUpper(param.Name))
		fieldConverters = append(fieldConverters, genericFieldConverter(param.Path))
	}

	return TableInfo{
		Header: Headers,
		RowConverter: func(u unstructured.Unstructured) []string {
			row := make([]string, len(fieldConverters))
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
		convertResourcesToTable(resources, tableInfo.RowConverter),
		tableInfo.Header,
	)
}

type FieldConverter func(u unstructured.Unstructured) string

type RowConverter func(unstructured.Unstructured) []string

type TableInfo struct {
	Header       []string
	RowConverter RowConverter
}

func convertResourcesToTable(resources []unstructured.Unstructured, rowConverter RowConverter) [][]string {
	slices.SortFunc(resources, func(a, b unstructured.Unstructured) int {
		return cmp.Compare(a.GetNamespace(), b.GetNamespace())
	})

	var result [][]string
	for _, resource := range resources {
		result = append(result, rowConverter(resource))
	}
	return result
}
