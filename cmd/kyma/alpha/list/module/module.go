package module

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/fatih/color"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/clusterinfo"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var moduleTemplateResource = schema.GroupVersionResource{
	Group:    "operator.kyma-project.io",
	Version:  "v1alpha1",
	Resource: "moduletemplates",
}

var kymaResource = schema.GroupVersionResource{
	Group:    "operator.kyma-project.io",
	Version:  "v1alpha1",
	Resource: "kymas",
}

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
		Use:   "module [flags]",
		Short: "list all available modules available for creation in the cluster",
		Long: `Use this command to list Kyma modules availabel in the cluster.

### Detailed description

For more information on Kyma modules, see the "create module" command.

This command lists all available modules in the cluster. 
A module is available when a ModuleTemplate is found for instantiating it with proper defaults.

Optionally, you can manually add a release channel to filter available modules only for the given channel.

Also, you can specify a Kyma to use to lookup only the active modules within that Kyma instance. If this is specified,
the ModuleTemplates will also have a Field called "State" which will reflect the actual state of the module.

Finally, you can restrict and select a custom namespace for the command.
`,

		Example: `Examples:
List all modules
		kyma alpha list module
List all modules in the "regular" channel
		kyma alpha list module --channel regular
List all modules for the kyma "some-kyma" in the namespace "custom" in the "alpha" channel
		kyma alpha list module --kyma some-kyma -c alpha -n "custom"
List all modules for the kyma "some-kyma" in the "alpha" channel
		kyma alpha list module --kyma some-kyma -c alpha
`,
		RunE:    func(cmd *cobra.Command, args []string) error { return c.Run(cmd.Context(), args) },
		Aliases: []string{"mod", "mods", "modules"},
	}

	cmd.Flags().StringVarP(&o.Channel, "channel", "c", "", "Channel to use for the module template.")
	cmd.Flags().DurationVarP(
		&o.Timeout, "timeout", "t", 1*time.Minute, "Maximum time for the list operation to retrieve ModuleTemplates.",
	)
	cmd.Flags().StringVarP(
		&o.KymaName, "kyma", "k", "",
		"The namespaced name of the kyma to use to list active module templates in the form 'namespace/name'.",
	)
	cmd.Flags().StringVarP(
		&o.Namespace, "namespace", "n", metav1.NamespaceAll,
		"The namespace to use. An empty namespace uses 'default'",
	)

	return cmd
}

func (cmd *command) Run(ctx context.Context, _ []string) error {

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
	start := time.Now()

	var err error
	if cmd.K8s, err = kube.NewFromConfigWithTimeout("", cmd.KubeconfigPath, cmd.opts.Timeout); err != nil {
		return fmt.Errorf("failed to initialize the Kubernetes client from given kubeconfig: %w", err)
	}

	if _, err := clusterinfo.Discover(ctx, cmd.K8s.Static()); err != nil {
		return err
	}

	if cmd.opts.KymaName != "" {
		kyma, err := cmd.K8s.Dynamic().Resource(kymaResource).Namespace(cmd.opts.Namespace).Get(
			ctx, cmd.opts.KymaName, metav1.GetOptions{},
		)
		if err != nil {
			return fmt.Errorf("could not get kyma, verify %s: %w", cmd.opts.KymaName, err)
		}
		if err := cmd.printKymaActiveTemplates(ctx, kyma); err != nil {
			return err
		}
	} else {
		templates, err := cmd.K8s.Dynamic().Resource(moduleTemplateResource).Namespace(cmd.opts.Namespace).List(
			ctx, metav1.ListOptions{},
		)
		if err != nil {
			return err
		}
		if err := cmd.printModuleTemplates(templates); err != nil {
			return err
		}
	}

	l := cli.NewLogger(cmd.opts.Verbose).Sugar()
	l.Infof("listing module template(s) took %s", time.Since(start))

	return nil
}

func (cmd *command) printKymaActiveTemplates(ctx context.Context, kyma *unstructured.Unstructured) error {
	statusItems, _, err := unstructured.NestedSlice(kyma.UnstructuredContent(), "status", "moduleStatus")
	if err != nil {
		return fmt.Errorf("could not parse moduleStatus: %w", err)
	}
	var entries []tableEntryWithState
	for i := range statusItems {
		item, _ := statusItems[i].(map[string]interface{})
		templateInfo, _, err := unstructured.NestedMap(item, "templateInfo")
		if err != nil {
			return fmt.Errorf("could not parse templateInfo: %w", err)
		}
		channel, _ := templateInfo["channel"]
		if cmd.opts.Channel != "" && cmd.opts.Channel != channel {
			continue
		}
		template, err := cmd.K8s.Dynamic().Resource(moduleTemplateResource).Namespace(templateInfo["namespace"].(string)).Get(
			ctx, templateInfo["name"].(string), metav1.GetOptions{},
		)
		if err != nil {
			return err
		}
		fqdn, err := getModuleName(*template)
		if err != nil {
			return err
		}
		entries = append(
			entries, tableEntryWithState{
				item["moduleName"].(string),
				fqdn,
				channel.(string),
				templateInfo["version"].(string),
				fmt.Sprintf("%s/%s", item["namespace"].(string), item["name"].(string)),
				item["state"].(string),
			},
		)
	}
	return printTable(entries, cmd.NonInteractive)
}

func (cmd *command) printModuleTemplates(templates *unstructured.UnstructuredList) error {
	var entries []tableEntry
	for _, template := range templates.Items {
		name, ok := template.GetLabels()["operator.kyma-project.io/module-name"]
		if !ok {
			name = template.GetName()
		}
		fqdn, err := getModuleName(template)
		if err != nil {
			return err
		}

		channel, err := getModuleChannel(template)
		if err != nil {
			return err
		}
		if cmd.opts.Channel != "" && cmd.opts.Channel != channel {
			continue
		}

		namespacedName := fmt.Sprintf("%s/%s", template.GetNamespace(), template.GetName())

		version, err := getModuleVersion(template)
		if err != nil {
			return err
		}

		entries = append(entries, tableEntry{name, fqdn, channel, version, namespacedName})
	}

	return printTable(entries, cmd.NonInteractive)
}

type tableEntryWithState struct {
	Module, FQDN, Channel, Version, Source, State string
}
type tableEntry struct {
	Module, FQDN, Channel, Version, Source string
}

func printTable[entry any](entries []entry, nonInteractive bool) error {
	if nonInteractive {
		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", data)
		return nil
	}
	headerFmt := color.New(color.FgWhite, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgCyan).SprintfFunc()
	typeOfTableEntry := reflect.TypeOf(entries).Elem()
	fieldAmount := typeOfTableEntry.NumField()
	columnHeaders := make([]interface{}, 0, fieldAmount)
	for i := 0; i < fieldAmount; i++ {
		columnHeaders = append(columnHeaders, typeOfTableEntry.Field(i).Name)
	}
	tbl := table.New(columnHeaders...)
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for _, entry := range entries {
		value := reflect.ValueOf(entry)
		rowValues := make([]interface{}, 0, fieldAmount)
		for i := 0; i < fieldAmount; i++ {
			rowValues = append(rowValues, value.Field(i).String())
		}
		tbl.AddRow(rowValues...)
	}
	tbl.Print()
	return nil
}

func getModuleName(template unstructured.Unstructured) (string, error) {
	name, _, err := unstructured.NestedString(
		template.UnstructuredContent(), "spec", "descriptor", "component", "name",
	)
	if err != nil {
		return "", fmt.Errorf("could not resolve module version for %s", template)
	}
	return name, nil
}
func getModuleVersion(template unstructured.Unstructured) (string, error) {
	version, _, err := unstructured.NestedString(
		template.UnstructuredContent(), "spec", "descriptor", "component", "version",
	)
	if err != nil {
		return "", fmt.Errorf("could not resolve module version for %s", template)
	}
	return version, nil
}

func getModuleChannel(template unstructured.Unstructured) (string, error) {
	channel, _, err := unstructured.NestedString(template.UnstructuredContent(), "spec", "channel")
	if err != nil {
		return "", err
	}
	return channel, nil
}
