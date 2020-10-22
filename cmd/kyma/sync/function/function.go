package function

import (
	"context"
	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
)

type command struct {
	opts *Options
	cli.Command
}

//NewCmd creates a new init command
func NewCmd(o *Options) *cobra.Command {
	c := command{
		opts:    o,
		Command: cli.Command{Options: o.Options},
	}
	cmd := &cobra.Command{
		Use:   "function",
		Short: "Saves locally resources for requested Function.",
		Long: `Use this command to create the local workspace with the default structure of your Function's code and dependencies. Update this configuration to your references and apply it to a Kyma cluster. 
Use the flags to specify the initial configuration for your Function or to choose the location for your project.`, //todo
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Run()
		},
	}

	cmd.Flags().StringVar(&o.Name, "name", "", `Function name.`)
	cmd.Flags().StringVarP(&o.Namespace, "namespace", "n", "", `Namespace from which you want to sync Function.`)
	cmd.Flags().StringVarP(&o.OutputPath, "output", "o", "", `Full path to the directory where you want to save the project.`)

	return cmd
}

func (c *command) Run() error {
	s := c.NewStep("Generating project structure")

	if err := c.opts.setDefaults(c.K8s.DefaultNamespace()); err != nil {
		s.Failure()
		return err
	}

	if _, err := os.Stat(c.opts.OutputPath); os.IsNotExist(err) {
		err = os.MkdirAll(c.opts.OutputPath, 0700)
		if err != nil {
			return err
		}
	}

	ctx := context.Background()

	var buildClient client.Build = func(namespace string, resource schema.GroupVersionResource) client.Client {
		return c.K8s.Dynamic().Resource(resource).Namespace(namespace)
	}

	cfg := workspace.Cfg{
		Name:      c.opts.Name,
		Namespace: c.opts.Namespace,
	}

	err := workspace.Synchronise(ctx, cfg, c.opts.OutputPath, buildClient)
	if err != nil {
		s.Failure()
		return err
	}

	s.Successf("Function synchronised in %s", c.opts.OutputPath)
	return nil
}
