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
		Short: "Disables a module in the cluster or in the given Kyma resource.",
		Long: `Use this command to disable active Kyma modules in the cluster.

### Detailed description

For more information on Kyma modules, see the 'create module' command.

This command disables an active module in the cluster.
`,

		Example: `
Disable "my-module" from the "alpha" channel in "default-kyma" in "kyma-system" Namespace
		kyma alpha disable module my-module -c alpha -n kyma-system -k default-kyma
`,
		RunE:    func(cmd *cobra.Command, args []string) error { return c.Run(cmd.Context(), args) },
		Aliases: []string{"mod", "mods", "modules"},
	}

	cmd.Flags().DurationVarP(
		&o.Timeout, "timeout", "t", 1*time.Minute, "Maximum time for the operation to disable a module.",
	)
	cmd.Flags().StringVarP(
		&o.Channel, "channel", "c", "",
		"Module's channel to use.",
	)
	cmd.Flags().StringVarP(
		&o.Namespace, "namespace", "n", cli.KymaNamespaceDefault,
		"Kyma Namespace to use. If empty, the default 'kyma-system' Namespace is used.",
	)
	cmd.Flags().StringVarP(
		&o.KymaName, "kyma-name", "k", cli.KymaNameDefault,
		"Kyma resource to use. If empty, 'default-kyma' is used.",
	)
	cmd.Flags().BoolVarP(&o.Force, "force-conflicts", "f", false,
		"Forces the patching of Kyma spec modules in case their managed field was edited by a source other than Kyma CLI.",
	)
	cmd.Flags().BoolVarP(&o.Wait, "wait", "w", false,
		"Wait until the given Kyma resource is ready.",
	)

	return cmd
}

func (cmd *command) Run(ctx context.Context, args []string) error {
	if !cmd.opts.NonInteractive {
		cli.AlphaWarn()
	}

	if len(args) != 1 {
		return errors.New("you must pass one Kyma module name to disable")
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
	moduleInteractor := module.NewInteractor(cmd.K8s, kyma, cmd.opts.Force)
	modules, err := moduleInteractor.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get modules: %w", err)
	}
	desiredModules, err := disableModule(modules, moduleName, cmd.opts.Channel)
	if err != nil {
		return fmt.Errorf("could not disable module: %w", err)
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
	l.Infof("disabling module took %s", time.Since(start))

	return nil
}

func disableModule(modules []interface{}, name, channel string) ([]interface{}, error) {
	for i := range modules {
		mod, _ := modules[i].(map[string]interface{})
		moduleName, found := mod["name"]
		if !found {
			return nil, errors.New("invalid item in modules spec: name field missing")
		}
		if moduleName != name {
			continue
		}
		if channel != "" {
			moduleChannel, cFound := mod["channel"]
			if cFound && moduleChannel != channel {
				continue
			}
		}
		return append(modules[:i], modules[i+1:]...), nil
	}

	return nil, errors.Errorf("no active module %s %s found to disable", name, channel)
}
