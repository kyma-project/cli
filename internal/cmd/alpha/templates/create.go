package templates

import (
	"context"
	"fmt"
	"io"
	"os"

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

func BuildCreateCommand(clientGetter KubeClientGetter, options *CreateOptions) *cobra.Command {
	return buildCreateCommand(os.Stdout, clientGetter, options)
}

func buildCreateCommand(out io.Writer, clientGetter KubeClientGetter, options *CreateOptions) *cobra.Command {
	extraValues := []parameters.Value{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: options.Description,
		Long:  options.DescriptionLong,
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

	flags := append(options.CustomFlags, commonResourceFlags(options.ResourceInfo.Scope)...)
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

	clierr = setExtraValues(u, args.extraValues)
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
