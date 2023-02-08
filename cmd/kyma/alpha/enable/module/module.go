package module

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/clusterinfo"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const defaultKymaName = "default-kyma"

var kymaResource = schema.GroupVersionResource{
	Group:    "operator.kyma-project.io",
	Version:  "v1alpha1",
	Resource: "kymas",
}

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
		&o.Namespace, "namespace", "n", metav1.NamespaceDefault,
		"The namespace of the Kyma to use. An empty namespace uses 'default'",
	)
	cmd.Flags().StringVarP(
		&o.KymaName, "kyma-name", "k", defaultKymaName,
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

	if cmd.K8s == nil {
		var err error
		if cmd.K8s, err = kube.NewFromConfigWithTimeout("", cmd.KubeconfigPath, cmd.opts.Timeout); err != nil {
			return fmt.Errorf("failed to initialize the Kubernetes client from given kubeconfig: %w", err)
		}
	}

	if _, err := clusterinfo.Discover(ctx, cmd.K8s.Static()); err != nil {
		return err
	}

	kyma, err := cmd.K8s.Dynamic().Resource(kymaResource).Namespace(cmd.opts.Namespace).Get(
		ctx, cmd.opts.KymaName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("could not get Kyma %s/%s: %w", cmd.opts.Namespace, cmd.opts.KymaName, err)
	}

	modules, _, err := unstructured.NestedSlice(kyma.UnstructuredContent(), "spec", "modules")
	if err != nil {
		return fmt.Errorf("could not parse modules spec: %w", err)
	}

	desiredModules, err := enableModule(modules, moduleName, cmd.opts.Channel)
	if err != nil {
		return fmt.Errorf("could not find module to enable: %w", err)
	}

	if len(modules) != len(desiredModules) {
		err = unstructured.SetNestedSlice(kyma.Object, desiredModules, "spec", "modules")
		if err != nil {
			return fmt.Errorf("failed to set modules list in Kyma spec: %w", err)
		}
		_, err = cmd.K8s.Dynamic().Resource(kymaResource).Namespace(cmd.opts.Namespace).Update(
			ctx, kyma, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update Kyma %s in %s: %w", cmd.opts.KymaName, cmd.opts.Namespace, err)
		}

		if cmd.opts.Wait {
			time.Sleep(2 * time.Second)
			checkFn := func(u *unstructured.Unstructured) (bool, error) {
				status, exists, err := unstructured.NestedString(u.Object, "status", "state")
				if err != nil {
					return false, errors.Wrap(err, "error waiting for Kyma readiness")
				}
				return exists && status == "Ready", nil
			}
			err = cmd.K8s.WatchResource(kymaResource, cmd.opts.KymaName, cmd.opts.Namespace, checkFn)
			if err != nil {
				return errors.Wrap(err, "failed to watch resource Kyma for state 'Ready'")
			}
		}
	}

	l := cli.NewLogger(cmd.opts.Verbose).Sugar()
	l.Infof("enabling module took %s", time.Since(start))

	return nil
}

func enableModule(modules []interface{}, name, channel string) ([]interface{}, error) {
	for _, m := range modules {
		module, _ := m.(map[string]interface{})
		moduleName, found := module["name"]
		if !found {
			return nil, errors.New("invalid item in modules spec: name field missing")
		}
		if moduleName == name {
			moduleChannel, cFound := module["channel"]
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
