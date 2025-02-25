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
	RootCommand  types.RootCommand
}

func BuildCreateCommand(clientGetter KubeClientGetter, options *CreateOptions) *cobra.Command {
	return buildCreateCommand(os.Stdout, clientGetter, options)
}

func buildCreateCommand(out io.Writer, clientGetter KubeClientGetter, options *CreateOptions) *cobra.Command {
	extraValues := []parameters.Value{}
	cmd := &cobra.Command{
		Use:     "create <resource_name> [flags]",
		Short:   options.Description,
		Long:    options.DescriptionLong,
		Example: buildCreateExample(options),
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(createResource(&createArgs{
				out:           out,
				ctx:           cmd.Context(),
				clientGetter:  clientGetter,
				createOptions: options,
				extraValues:   extraValues,
			}))
		},
	}

	// define resource name as required args[0]
	nameArgValue := parameters.NewTyped(resourceNameFlag.Type, resourceNameFlag.Path, resourceNameFlag.DefaultValue)
	cmd.Args = AssignRequiredNameArg(nameArgValue)
	extraValues = append(extraValues, nameArgValue)

	// define --namespace/-n flag only is resource is namespace scoped
	if options.ResourceInfo.Scope == types.NamespaceScope {
		namespaceFlagValue := parameters.NewTyped(resourceNamespaceFlag.Type, resourceNamespaceFlag.Path, resourceNamespaceFlag.DefaultValue)
		cmd.Flags().VarP(namespaceFlagValue, resourceNamespaceFlag.Name, resourceNamespaceFlag.Shorthand, resourceNamespaceFlag.Description)
		extraValues = append(extraValues, namespaceFlagValue)
	}

	for _, flag := range options.CustomFlags {
		value := parameters.NewTyped(flag.Type, flag.Path, flag.DefaultValue)
		cmd.Flags().VarP(value, flag.Name, flag.Shorthand, flag.Description)
		if flag.Required {
			_ = cmd.MarkFlagRequired(flag.Name)
		}
		extraValues = append(extraValues, value)
	}

	return cmd
}

func buildCreateExample(options *CreateOptions) string {
	requiredFlagsExample := strings.Join(buildRequiredFlagsExamples(options), " ")
	template := "  # create ROOT_COMMAND resource\n" +
		"  kyma alpha ROOT_COMMAND create <resource_name> " + requiredFlagsExample

	if options.ResourceInfo.Scope == types.NamespaceScope {
		template += "\n\n" +
			"  # create ROOT_COMMAND resource in specific namespace\n" +
			"  kyma alpha ROOT_COMMAND create <resource_name> --namespace <resource_namespace> " + requiredFlagsExample
	}

	return strings.ReplaceAll(template, "ROOT_COMMAND", options.RootCommand.Name)
}

func buildRequiredFlagsExamples(options *CreateOptions) []string {
	requiredFlagsExamples := []string{}
	for _, flag := range options.CustomFlags {
		if flag.Required {
			requiredFlagsExamples = append(requiredFlagsExamples,
				fmt.Sprintf("--%s <value>", flag.Name),
			)
		}
	}

	return requiredFlagsExamples
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

	clierr = parameters.Set(u, args.extraValues)
	if clierr != nil {
		return clierr
	}

	err := client.RootlessDynamic().Apply(args.ctx, u)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create resource"))
	}

	fmt.Fprintf(args.out, "resource %s applied\n", getResourceName(args.createOptions.ResourceInfo.Scope, u))
	return nil
}
