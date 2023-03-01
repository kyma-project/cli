package module

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/cli/alpha/module"
	"github.com/kyma-project/lifecycle-manager/api/v1beta1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type command struct {
	cli.Command
	opts *Options
}

// NewCmd creates a new Kyma CLI command
func NewCmd(o *Options) *cobra.Command {
	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:   "module [kyma] [flags]",
		Short: "Lists all modules available for creation in the cluster or in the given Kyma resource",
		Long: `Use this command to list Kyma modules available in the cluster.

### Detailed description

For more information on Kyma modules, see the 'create module' command.

This command lists all available modules in the cluster. 
A module is available when a ModuleTemplate is found for instantiating it with proper defaults.

Optionally, you can manually add a release channel to filter available modules only for the given channel.

Also, you can specify a Kyma to look up only the active modules within that Kyma instance. If this is specified,
the ModuleTemplates will also have a Field called **State** which will reflect the actual state of the module.

Finally, you can restrict and select a custom Namespace for the command.
`,

		Example: `
List all modules
		kyma alpha list module
List all modules in the "regular" channel
		kyma alpha list module --channel regular
List all modules for the kyma "some-kyma" in the namespace "custom" in the "alpha" channel
		kyma alpha list module -k some-kyma -c alpha -n custom
List all modules for the kyma "some-kyma" in the "alpha" channel
		kyma alpha list module -k some-kyma -c alpha
`,
		RunE:    func(cmd *cobra.Command, args []string) error { return c.Run(cmd.Context(), args) },
		Aliases: []string{"mod", "mods", "modules"},
	}

	cmd.Flags().StringVarP(&o.Channel, "channel", "c", "", "Channel to use for the module template.")
	cmd.Flags().DurationVarP(
		&o.Timeout, "timeout", "t", 1*time.Minute, "Maximum time for the list operation to retrieve ModuleTemplates.",
	)
	cmd.Flags().StringVarP(
		&o.KymaName, "kyma-name", "k", "",
		"Kyma resource to use.",
	)
	cmd.Flags().StringVarP(
		&o.Namespace, "namespace", "n", metav1.NamespaceAll,
		"The Namespace to list the modules in. By default uses all namespaces.",
	)
	cmd.Flags().BoolVar(
		&o.NoHeaders, "no-headers", false,
		"When using the default output format, don't print headers. (default print headers)",
	)

	cmd.Flags().StringVarP(
		&o.Output, "output", "o", "tabwriter",
		fmt.Sprintf("Output format. One of: %s. By default uses https://pkg.go.dev/text/tabwriter. It is currently impossible to add your own template file.", ValidOutputs),
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

	//if !cmd.opts.Verbose {
	//	stderr := os.Stderr
	//	os.Stderr = nil
	//	defer func() { os.Stderr = stderr }()
	//}

	if !cmd.opts.NonInteractive {
		cli.AlphaWarn()
	}

	if err := cmd.opts.validateFlags(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, cmd.opts.Timeout)
	defer cancel()

	return cmd.run(ctx)
}

func (cmd *command) run(ctx context.Context) error {
	clusterAccess := cmd.NewStep("Ensuring Cluster Access")
	if _, err := cmd.EnsureClusterAccess(ctx, cmd.opts.Timeout); err != nil {
		clusterAccess.Failuref("Could not ensure cluster Access")
		return err
	}
	clusterAccess.Successf("Successfully connected to cluster")

	writer := &bytes.Buffer{}

	if cmd.opts.KymaName != "" {
		cmd.NewStep("Listing Active Modules for Kyma Resource")
		kyma := &v1beta1.Kyma{}

		if err := cmd.K8s.Ctrl().Get(ctx, ctrl.ObjectKey{
			Namespace: cmd.opts.Namespace,
			Name:      cmd.opts.KymaName,
		}, kyma); err != nil {
			cmd.CurrentStep.Failure()
			return fmt.Errorf("could not get kyma %s/%s: %w", cmd.opts.Namespace, cmd.opts.KymaName, err)
		}
		if err := cmd.printKymaActiveTemplates(ctx, writer, kyma); err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
	} else {
		cmd.NewStep("Listing Available Modules in the Cluster")
		templates := v1beta1.ModuleTemplateList{}
		if err := cmd.K8s.Ctrl().List(ctx, &templates, &ctrl.ListOptions{Namespace: cmd.opts.Namespace}); err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
		if err := cmd.printModuleTemplates(writer, templates.Items, false); err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
	}
	cmd.CurrentStep.Success()

	fmt.Print(writer.String())

	return nil
}

func (cmd *command) printKymaActiveTemplates(ctx context.Context, writer io.Writer, kyma *v1beta1.Kyma) error {
	templates := make([]v1beta1.ModuleTemplate, 0)
	for _, status := range kyma.Status.Modules {
		tmpltStatus := status.Template
		obj := &metav1.PartialObjectMetadata{}
		obj.SetGroupVersionKind(tmpltStatus.GroupVersionKind())
		obj.SetName(tmpltStatus.GetName())
		obj.SetNamespace(tmpltStatus.GetNamespace())
		key := ctrl.ObjectKeyFromObject(obj)
		tmplt := v1beta1.ModuleTemplate{}
		if err := cmd.K8s.Ctrl().Get(ctx, key, &tmplt); err != nil {
			return err
		}
		anns := tmplt.GetAnnotations()
		if anns == nil {
			anns = make(map[string]string)
		}
		anns[module.TemplateStateAnnotation] = string(status.State)
		tmplt.SetAnnotations(anns)
		templates = append(templates, tmplt)
	}
	return cmd.printModuleTemplates(writer, templates, true)
}

func (cmd *command) printModuleTemplates(writer io.Writer, templates []v1beta1.ModuleTemplate, appendState bool) error {
	var err error

	switch cmd.opts.Output {
	case "tabwriter":
		err = module.NewDefaultTemplateTable(cmd.opts.NoHeaders, appendState).Print(writer, templates)
	case "yaml":
		var data []byte
		data, err = yaml.Marshal(&v1beta1.ModuleTemplateList{Items: templates})
		if err != nil {
			break
		}
		_, err = writer.Write(data)
	case "json":
		var data []byte
		data, err = json.MarshalIndent(&v1beta1.ModuleTemplateList{Items: templates}, "", "  ")
		if err != nil {
			break
		}
		_, err = writer.Write(data)
	default:
		return ErrNoValidOutput
	}

	return err
}
