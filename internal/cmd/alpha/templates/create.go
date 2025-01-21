package templates

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/parameters"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/types"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type KubeClientGetter interface {
	GetKubeClientWithClierr() (kube.Client, clierror.Error)
}

type CreateOptions struct {
	types.CreateCommand
	ResourceInfo types.ResourceInfo
}

func BuildCreateCommand(clientGetter KubeClientGetter, createOptions *CreateOptions) *cobra.Command {
	return buildCreateCommand(os.Stdout, clientGetter, createOptions)
}

func buildCreateCommand(out io.Writer, clientGetter KubeClientGetter, createOptions *CreateOptions) *cobra.Command {
	extraValues := []parameters.Value{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: createOptions.Description,
		Long:  createOptions.DescriptionLong,
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(createResource(&createArgs{
				out:           out,
				ctx:           cmd.Context(),
				clientGetter:  clientGetter,
				createOptions: createOptions,
				extraValues:   extraValues,
			}))
		},
	}

	flags := append(createOptions.CustomFlags, buildDefaultFlags(createOptions.ResourceInfo.Scope)...)
	for _, flag := range flags {
		value := parameters.NewTyped(flag.Type, flag.Path, flag.DefaultValue)
		cmd.Flags().VarP(value, flag.Name, flag.Shorthand, flag.Description)
		if flag.Required {
			_ = cmd.MarkFlagRequired(flag.Name)
		}
		extraValues = append(extraValues, value)
	}

	return cmd
}

type createArgs struct {
	out           io.Writer
	ctx           context.Context
	clientGetter  KubeClientGetter
	createOptions *CreateOptions
	extraValues   []parameters.Value
}

func createResource(args *createArgs) clierror.Error {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   args.createOptions.ResourceInfo.Group,
		Version: args.createOptions.ResourceInfo.Version,
		Kind:    args.createOptions.ResourceInfo.Kind,
	})

	client, clierr := args.clientGetter.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	for _, extraValue := range args.extraValues {
		value := extraValue.GetValue()
		if value == nil {
			// value is not set and has no default value
			continue
		}

		fields := strings.Split(
			// remove optional dot at the beginning of the path
			strings.TrimPrefix(extraValue.GetPath(), "."),
			".",
		)

		err := unstructured.SetNestedField(u.Object, value, fields...)
		if err != nil {
			return clierror.Wrap(err, clierror.New(
				fmt.Sprintf("failed to set value %v for path %s", value, extraValue.GetPath()),
			))
		}
	}

	err := client.RootlessDynamic().Apply(args.ctx, u)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create resource"))
	}

	fmt.Fprintf(args.out, "resource %s applied\n", getResourceName(args.createOptions.ResourceInfo.Scope, u))
	return nil
}

func buildDefaultFlags(resourceScope types.Scope) []types.CreateCustomFlag {
	params := []types.CreateCustomFlag{
		{
			Name:        "name",
			Type:        types.StringCustomFlagType,
			Description: "name of the resource",
			Path:        ".metadata.name",
			Required:    true,
		},
	}
	if resourceScope == types.NamespaceScope {
		params = append(params, types.CreateCustomFlag{
			Name:         "namespace",
			Type:         types.StringCustomFlagType,
			Description:  "resource namespace",
			Path:         ".metadata.namespace",
			DefaultValue: "default",
		})
	}

	return params
}

func getResourceName(scope types.Scope, u *unstructured.Unstructured) string {
	if scope == types.NamespaceScope {
		return fmt.Sprintf("%s/%s", u.GetNamespace(), u.GetName())
	}

	return u.GetName()
}
