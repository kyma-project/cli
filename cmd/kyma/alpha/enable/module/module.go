package module

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kyma-project/cli/internal/cli/alpha/module"
	"github.com/kyma-project/lifecycle-manager/api/v1beta1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"

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
		Short: "Enables a module in the cluster or in the given Kyma resource.",
		Long: `Use this command to enable Kyma modules available in the cluster.

### Detailed description

For more information on Kyma modules, see the 'create module' command.

This command enables an available module in the cluster. 
A module is available when it is released with a ModuleTemplate. The ModuleTemplate is used for instantiating the module with proper default configuration.
`,

		Example: `
Enable "my-module" from "alpha" channel in "default-kyma" in "kyma-system" Namespace
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
		"Module's channel to enable.",
	)
	cmd.Flags().StringVarP(
		&o.Namespace, "namespace", "n", cli.KymaNamespaceDefault,
		"Kyma Namespace to use. If empty, the default 'kyma-system' Namespace is used.",
	)
	cmd.Flags().StringVarP(
		&o.KymaName, "kyma-name", "k", cli.KymaNameDefault,
		"Kyma resource to use. If empty, 'default-kyma' is used.",
	)
	cmd.Flags().BoolVarP(
		&o.Force, "force-conflicts", "f", false,
		"Forces the patching of Kyma spec modules in case their managed field was edited by a source other than Kyma CLI.",
	)
	cmd.Flags().BoolVarP(
		&o.Wait, "wait", "w", false,
		"Wait until the given Kyma resource is ready.",
	)

	return cmd
}

func (cmd *command) Run(ctx context.Context, args []string) error {
	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}
	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	l := cli.NewLogger(cmd.opts.Verbose).Sugar()
	undo := zap.RedirectStdLog(l.Desugar())
	defer undo()

	if !cmd.opts.Verbose {
		stderr := os.Stderr
		os.Stderr = nil
		defer func() { os.Stderr = stderr }()
	}

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

	return cmd.run(ctx, l, moduleName)
}

func (cmd *command) run(ctx context.Context, l *zap.SugaredLogger, moduleName string) error {
	clusterAccess := cmd.NewStep("Ensuring Cluster Access")
	if _, err := cmd.EnsureClusterAccess(ctx, cmd.opts.Timeout); err != nil {
		clusterAccess.Failuref("Could not ensure cluster Access")
		return err
	}
	clusterAccess.Successf("Successfully connected to cluster")

	kyma := types.NamespacedName{Name: cmd.opts.KymaName, Namespace: cmd.opts.Namespace}
	moduleInteractor := module.NewInteractor(l, cmd.K8s, kyma, cmd.opts.Timeout, cmd.opts.Force)
	modules, err := moduleInteractor.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get modules: %w", err)
	}

	desiredModules := enableModule(modules, moduleName, cmd.opts.Channel)

	patchStep := cmd.NewStep("Patching modules for Kyma")
	if err = moduleInteractor.Apply(ctx, desiredModules); err != nil {
		patchStep.Failuref("Could not apply modules for Kyma")
		return err
	}
	patchStep.Successf("Modules patched!")

	if cmd.opts.Wait {
		waitStep := cmd.NewStep("Waiting for Kyma to become Ready")
		if err = moduleInteractor.WaitUntilReady(ctx); err != nil {
			waitStep.Failuref("Kyma did not get Ready")
			return err
		}
		waitStep.Successf("Kyma is Ready")
	}

	return nil
}

func enableModule(modules []v1beta1.Module, name, channel string) []v1beta1.Module {
	for _, mod := range modules {
		if mod.Name == name {
			if channel == "" || mod.Channel == channel {
				// module already enabled
				return modules
			}
		}
	}

	newModule := v1beta1.Module{Name: name}
	if channel != "" {
		newModule.Channel = channel
	}

	modules = append(modules, newModule)

	return modules
}
