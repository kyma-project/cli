package function

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/hydroform/function/pkg/client"
	"github.com/kyma-project/hydroform/function/pkg/manager"
	"github.com/kyma-project/hydroform/function/pkg/operator"
	operator_types "github.com/kyma-project/hydroform/function/pkg/operator/types"
	resources "github.com/kyma-project/hydroform/function/pkg/resources/unstructured"
	"github.com/kyma-project/hydroform/function/pkg/workspace"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type command struct {
	opts *Options
	cli.Command
}

// NewCmd creates a new apply command
func NewCmd(o *Options) *cobra.Command {
	c := command{
		opts:    o,
		Command: cli.Command{Options: o.Options},
	}
	cmd := &cobra.Command{
		Use:   "function",
		Short: "Applies local resources for your Function to the Kyma cluster.",
		Long: `Use this command to apply the local sources of your Function's code and dependencies to the Kyma cluster. 
Use the flags to specify the desired location for the source files or run the command to validate and print the output resources.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Run()
		},
	}

	cmd.Flags().StringVarP(&o.Filename, "filename", "f", "", `Full path to the config file.`)
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", false, `Validated list of objects to be created from sources.`)
	cmd.Flags().DurationVarP(&o.Timeout, "timeout", "t", 0, `Maximum time during which the local resources are being applied, where "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
	cmd.Flags().BoolVarP(&o.Watch, "watch", "w", false, `Flag used to watch resources applied to the cluster to make sure that everything is applied in the correct order.`)
	cmd.Flags().Var(&o.OnError, "onerror", `Flag used to define the Kyma CLI's reaction to an error when applying resources to the cluster. Use one of these options: 
- nothing
- purge`)
	cmd.Flags().VarP(&o.Output, "output", "o", `Flag used to define the command output format. Use one of these options:
- text
- json
- yaml
- none`)

	return cmd
}

func (c *command) Run() error {
	if c.opts.Filename == "" {
		filename, err := defaultFilename()
		if err != nil {
			return err
		}
		c.opts.Filename = filename
	}

	if c.opts.Output.value == "yaml" {
		c.MuteLogger = true
	}
	file, err := os.Open(c.opts.Filename)
	if err != nil {
		return err
	}

	// Load project configuration
	step := c.NewStep("Loading configuration...")
	var configuration workspace.Cfg
	if err := yaml.NewDecoder(file).Decode(&configuration); err != nil {
		step.Failure()
		return errors.Wrap(err, "Could not decode the configuration file")
	}

	if configuration.SchemaVersion == "" {
		configuration.SchemaVersion = workspace.SchemaVersionDefault
	}

	if configuration.Source.SourcePath == "" {
		configuration.Source.SourcePath = filepath.Dir(c.opts.Filename)
	}

	if c.K8s, err = kube.NewFromConfig("", c.KubeconfigPath); err != nil {
		step.Failure()
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}
	client := c.K8s.Dynamic()

	mgr := manager.NewManager()

	function, err := resources.NewFunction(&configuration)
	if err != nil {
		step.Failure()
		return err
	}

	operators := []operator.Operator{}

	if len(configuration.Subscriptions) != 0 {

		subscriptionGVR := operator.SubscriptionGVR(configuration.SchemaVersion)
		err := isDependencyInstalled(client, subscriptionGVR)
		if err != nil {
			step.Failure()
			return err
		}

		subscriptions, err := resources.NewSubscriptions(configuration)
		if err != nil {
			step.Failure()
			return err
		}
		operators = append(operators, operator.NewSubscriptionOperator(client.Resource(subscriptionGVR).Namespace(configuration.Namespace),
			configuration.Name, configuration.Namespace, subscriptions...))
	}

	if len(configuration.APIRules) != 0 {

		err := isDependencyInstalled(client, operator_types.GVRApiRule)
		if err != nil {
			step.Failure()
			return err
		}
		apiRules, err := resources.NewAPIRule(configuration)
		if err != nil {
			step.Failure()
			return err
		}
		operators = append(operators, operator.NewAPIRuleOperator(client.Resource(operator_types.GVRApiRule).Namespace(configuration.Namespace),
			configuration.Name, apiRules...))
	}

	mgr.AddParent(
		operator.NewGenericOperator(client.Resource(operator_types.GVRFunction).Namespace(configuration.Namespace), function),
		operators,
	)

	options := manager.Options{
		Callbacks:          callbacks(c),
		OnError:            chooseOnError(c.opts.OnError),
		DryRun:             c.opts.DryRun,
		WaitForApply:       c.opts.Watch,
		SetOwnerReferences: true,
	}

	ctx, cancel := context.WithCancel(context.Background())
	if c.opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, c.opts.Timeout)
	}
	defer cancel()

	step.Successf("Configuration loaded")

	return mgr.Do(ctx, options)
}

const (
	operatingFormat     = "%s - %s operating... %s"
	createdFormat       = "%s - %s created %s"
	updatedFormat       = "%s - %s updated %s"
	skippedFormat       = "%s - %s unchanged %s"
	deletedFormat       = "%s - %s deleted %s"
	applyFailedFormat   = "%s - %s can't be applied %s"
	deleteFailedFormat  = "%s - %s can't be removed %s"
	unknownStatusFormat = "%s - %s can't resolve status %s"
	dryRunSuffix        = "(dry run)"
	yamlFormat          = "---\n%s\n"
	jsonFormat          = "%s\n"
)

func chooseOnError(onErr value) manager.OnError {
	if onErr.value == NothingOnError {
		return manager.NothingOnError

	}
	return manager.PurgeOnError
}

func isDependencyInstalled(client dynamic.Interface, dependencyCRD schema.GroupVersionResource) error {
	crdGVR := schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}

	_, err := client.Resource(crdGVR).Get(context.Background(), fmt.Sprintf("%s.%s", dependencyCRD.Resource, dependencyCRD.Group), metav1.GetOptions{})

	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return errors.Errorf("failed to apply %s. %s module is missing", dependencyCRD.Resource, dependencyCRD.Group)
		}
		return err
	}
	return nil
}

func (l *logger) chooseFormat(status client.StatusType) string {
	switch status {
	case client.StatusTypeCreated:
		return createdFormat
	case client.StatusTypeUpdated:
		return updatedFormat
	case client.StatusTypeSkipped:
		return skippedFormat
	case client.StatusTypeDeleted:
		return deletedFormat
	case client.StatusTypeApplyFailed:
		return applyFailedFormat
	case client.StatusTypeDeleteFailed:
		return deleteFailedFormat

	}
	return unknownStatusFormat
}

func (l *logger) formatSuffix() string {
	if l.opts.DryRun {
		return dryRunSuffix
	}
	return ""
}

type logger struct {
	*command
}

func callbacks(c *command) operator.Callbacks {
	logger := logger{c}
	return operator.Callbacks{
		Pre: []operator.Callback{
			logger.pre,
		},
		Post: []operator.Callback{
			logger.post,
		},
	}
}

func (l *logger) pre(v interface{}, err error) error {
	if err != nil {
		return err
	}
	entry, ok := v.(*unstructured.Unstructured)
	if !ok {
		return errors.New("can't parse interface{} to the Unstructured")
	}

	switch l.opts.Output.String() {
	case TextOutput:
		info := fmt.Sprintf(operatingFormat, entry.GetKind(), entry.GetName(), l.formatSuffix())
		l.NewStep(info)
		return nil
	case JSONOutput:
		if err != nil {
			return err
		}
		unstructured.RemoveNestedField(entry.Object, "metadata", "ownerReferences")
		unstructured.RemoveNestedField(entry.Object, "metadata", "labels", "ownerID")
		bytes, marshalError := json.MarshalIndent(entry.Object, "", "  ")
		if marshalError != nil {
			return marshalError
		}
		fmt.Printf(jsonFormat, string(bytes))
		return nil
	case YAMLOutput:
		if err != nil {
			return err
		}
		unstructured.RemoveNestedField(entry.Object, "metadata", "ownerReferences")
		unstructured.RemoveNestedField(entry.Object, "metadata", "labels", "ownerID")
		bytes, marshalError := yaml.Marshal(entry.Object)
		if marshalError != nil {
			return marshalError
		}
		fmt.Printf(yamlFormat, string(bytes))
		return nil
	case NoneOutput:
		return err
	}
	return err
}

func (l *logger) post(v interface{}, err error) error {
	entry, ok := v.(client.PostStatusEntry)
	if !ok {
		return errors.New("can't parse interface{} to StatusEntry interface")
	}

	if l.opts.Output.String() != TextOutput {
		return err
	}
	format := l.chooseFormat(entry.StatusType)
	info := fmt.Sprintf(format, entry.GetKind(), entry.GetName(), l.formatSuffix())
	step := l.CurrentStep
	if err != nil {
		step.Failuref(info)
	}
	step.Successf(info)
	return err
}
