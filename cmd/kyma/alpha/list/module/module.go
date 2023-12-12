package module

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/kyma-project/lifecycle-manager/api/shared"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/yaml"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/clusterinfo"
	"github.com/kyma-project/cli/internal/kube"
)

//go:embed list.tmpl
var moduleTemplates string

var moduleTemplateResource = schema.GroupVersionResource{
	Group:    shared.OperatorGroup,
	Version:  "v1beta2",
	Resource: shared.ModuleTemplateKind.Plural(),
}

var kymaResource = schema.GroupVersionResource{
	Group:    shared.OperatorGroup,
	Version:  "v1beta2",
	Resource: shared.KymaKind.Plural(),
}

type command struct {
	cli.Command
	opts *Options
	meta.RESTMapper
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
		&o.Namespace, "namespace", "n", cli.KymaNamespaceDefault,
		"The Namespace to list the modules in.",
	)
	cmd.Flags().BoolVarP(
		&o.AllNamespaces, "all-namespaces", "A", false,
		"If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace",
	)
	cmd.Flags().BoolVar(
		&o.NoHeaders, "no-headers", false,
		"When using the default output format, don't print headers. (default print headers)",
	)

	cmd.Flags().StringVarP(
		&o.Output, "output", "o", "go-template-file",
		"Output format. One of: (json, yaml). By default uses an in-built template file. It is currently impossible to add your own template file.",
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

	if cmd.K8s == nil {
		var err error
		if cmd.K8s, err = kube.NewFromConfigWithTimeout("", cmd.KubeconfigPath, cmd.opts.Timeout); err != nil {
			return fmt.Errorf("failed to initialize the Kubernetes client from given kubeconfig: %w", err)
		}
	}

	if cmd.RESTMapper == nil {
		disco, err := discovery.NewDiscoveryClientForConfig(cmd.K8s.RestConfig())
		if err != nil {
			cmd.RESTMapper = meta.NewDefaultRESTMapper(scheme.Scheme.PreferredVersionAllGroups())
		} else {
			cmd.RESTMapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(disco))
		}
	}

	if _, err := clusterinfo.Discover(ctx, cmd.K8s.Static()); err != nil {
		return err
	}

	if cmd.opts.KymaName != "" {
		kyma, err := cmd.K8s.Dynamic().Resource(kymaResource).Namespace(cmd.opts.Namespace).Get(
			ctx, cmd.opts.KymaName, metav1.GetOptions{},
		)
		if err != nil {
			return fmt.Errorf("could not get kyma %s/%s: %w", cmd.opts.Namespace, cmd.opts.KymaName, err)
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
	statusItems, _, err := unstructured.NestedSlice(kyma.UnstructuredContent(), "status", "modules")
	if err != nil {
		return fmt.Errorf("could not parse moduleStatus: %w", err)
	}
	templateList := &unstructured.UnstructuredList{Items: make([]unstructured.Unstructured, 0, len(statusItems))}

	for i := range statusItems {
		item, ok := statusItems[i].(map[string]interface{})
		if !ok {
			continue
		}
		tmplt, _, err := unstructured.NestedMap(item, "template")
		if err != nil {
			return fmt.Errorf("could not parse template: %w", err)
		}
		obj := &metav1.PartialObjectMetadata{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(tmplt, obj); err != nil {
			return fmt.Errorf("could not parse template info into obj %s: %w", obj, err)
		}
		mapping, err := cmd.RESTMapping(obj.GroupVersionKind().GroupKind(), obj.GroupVersionKind().Version)
		var resource schema.GroupVersionResource
		if err != nil {
			resource, _ = meta.UnsafeGuessKindToResource(obj.GroupVersionKind())
		} else {
			resource = mapping.Resource
		}
		tpl, err := cmd.K8s.Dynamic().Resource(resource).Namespace(obj.GetNamespace()).Get(
			ctx, obj.GetName(), metav1.GetOptions{},
		)
		if err != nil {
			return err
		}
		annotations := tpl.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		stateValue, ok := item["state"]
		if !ok {
			continue
		}
		stateAnnotationValue, ok := stateValue.(string)
		if !ok {
			continue
		}
		annotations["state.cmd.kyma-project.io"] = stateAnnotationValue
		tpl.SetAnnotations(annotations)
		templateList.Items = append(templateList.Items, *tpl)
		if templateList.GetKind() == "" {
			templateList.SetGroupVersionKind(moduleTemplateResource.GroupVersion().WithKind(tpl.GetKind() + "List"))
		}
	}
	return cmd.printModuleTemplates(templateList)
}

func (cmd *command) printModuleTemplates(templates *unstructured.UnstructuredList) error {
	switch cmd.opts.Output {
	case "go-template-file":
		return cmd.printModuleTemplatesTable(templates)
	case "yaml":
		data, err := yaml.Marshal(templates)
		if err != nil {
			return err
		}
		_, err = fmt.Printf("%s\n", data)
		return err
	default:
		data, err := json.MarshalIndent(templates, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Printf("%s\n", data)
		return err
	}
}

func (cmd *command) printModuleTemplatesTable(templates *unstructured.UnstructuredList) error {
	tmpl, err := template.New("module-template").Parse(moduleTemplates)
	if err != nil {
		return err
	}
	tabWriter := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', 0)
	if !cmd.opts.NoHeaders {
		headers := []string{
			shared.ModuleName,
			"Domain Name (FQDN)",
			"Channel",
			"Version",
			"Template",
			"State",
		}
		if _, err := tabWriter.Write([]byte(strings.Join(headers, "\t") + "\n")); err != nil {
			return err
		}
	}
	if err := tmpl.Execute(tabWriter, templates); err != nil {
		return fmt.Errorf("could not print table: %w", err)
	}
	return tabWriter.Flush()
}
