package module

import (
	"context"
	"fmt"
	"github.com/kyma-project/cli/internal/cli/alpha/module"
	"k8s.io/apimachinery/pkg/types"
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type command struct {
	cli.Command
	opts *Options
}

func NewCmd(o *Options) *cobra.Command {
	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:   "module [name] [flags]",
		Short: "Enables a module in the cluster or in the given Kyma resource",
		Long: `Use this command to enable Kyma modules available in the cluster.

### Detailed description

For more information on Kyma modules, see the 'create module' command.

This command enables an available module in the cluster. 
A module is available when a ModuleTemplate is found for instantiating it with proper defaults.
`,

		Example: `Example:
Enable "my-module" from "alpha"" channel in "default-kyma" from "kyma-system" namespace
		kyma alpha enable module my-module -c alpha -n kyma-system -k default-kyma
`,
		RunE:    func(cmd *cobra.Command, args []string) error { return c.Run(cmd.Context(), args) },
		Aliases: []string{"mod", "mods", "modules"},
	}

	cmd.Flags().DurationVarP(
		&o.Timeout, "timeout", "t", 1*time.Minute, "Maximum time for the operation to enable a module.",
	)
	cmd.Flags().StringVarP(
		&o.Channel, "channel", "c", "",
		"The channel of the module to enable.",
	)
	cmd.Flags().StringVarP(
		&o.Namespace, "namespace", "n", cli.KymaNamespaceDefault,
		"The namespace of the Kyma to use. An empty namespace defaults to 'kyma-system'",
	)
	cmd.Flags().StringVarP(
		&o.KymaName, "kyma-name", "k", cli.KymaNameDefault,
		"The name of the Kyma to use. An empty name uses 'default-kyma'",
	)
	cmd.Flags().BoolVarP(&o.Wait, "wait", "w", false,
		"Wait until the given Kyma resource is ready",
	)

	return cmd
}

func (cmd *command) Run(ctx context.Context, args []string) error {
	if !cmd.opts.NonInteractive {
		cli.AlphaWarn()
	}

	if len(args) != 1 {
		return errors.New("you must pass one Kyma module name to enable it")
	}
	moduleName := args[0]

	if err := cmd.opts.validateFlags(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, cmd.opts.Timeout)
	defer cancel()

	return cmd.run(ctx, moduleName)
}

func (cmd *command) run(ctx context.Context, moduleName string) error {
	start := time.Now()
	if err := cmd.EnsureClusterAccess(ctx, cmd.opts.Timeout); err != nil {
		return err
	}

	kyma := types.NamespacedName{Name: cmd.opts.KymaName, Namespace: cmd.opts.Namespace}
	moduleInteractor := module.NewInteractor(cmd.K8s, kyma)
	modules, err := moduleInteractor.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get modules: %w", err)
	}

	desiredModules, err := enableModule(modules, moduleName, cmd.opts.Channel)
	if err != nil {
		return fmt.Errorf("could not find module to enable: %w", err)
	}

	if len(modules) != len(desiredModules) {
		if err = moduleInteractor.Update(ctx, desiredModules); err != nil {
			return err
		}
		if cmd.opts.Wait {
			if err = moduleInteractor.WaitForKymaReadiness(); err != nil {
				return err
			}
		}
	}

	l := cli.NewLogger(cmd.opts.Verbose).Sugar()
	l.Infof("enabling module took %s", time.Since(start))

	return nil
}

func enableModule(modules []interface{}, name, channel string) ([]interface{}, error) {
	for _, m := range modules {
		mod, _ := m.(map[string]interface{})
		moduleName, found := mod["name"]
		if !found {
			return nil, errors.New("invalid item in modules spec: name field missing")
		}
		if moduleName == name {
			moduleChannel, cFound := mod["channel"]
			if cFound && moduleChannel == channel {
				// module already enabled
				return modules, nil
			}
		}
	}

	newModule := make(map[string]interface{})
	newModule["name"] = name
	newModule["channel"] = channel
	modules = append(modules, newModule)

	return modules, nil
}
