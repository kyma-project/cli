package function

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"github.com/kyma-incubator/hydroform/function/pkg/manager"
	"github.com/kyma-incubator/hydroform/function/pkg/operator"
	resources "github.com/kyma-incubator/hydroform/function/pkg/resources/unstructured"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type command struct {
	opts *Options
	cli.Command
}

//NewCmd creates a new apply command
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
		c.opts.Filename = defaultFilename()
	}

	file, err := os.Open(c.opts.Filename)
	if err != nil {
		return err
	}

	// Load project configuration
	var configuration workspace.Cfg
	if err := yaml.NewDecoder(file).Decode(&configuration); err != nil {
		return errors.Wrap(err, "Could not decode the configuration file")
	}

	if c.K8s, err = kube.NewFromConfig("", c.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}
	client := c.K8s.Dynamic()

	mgr := manager.NewManager()

	function, err := resources.NewFunction(configuration)
	if err != nil {
		return err
	}

	subscriptions, err := resources.NewSubscriptions(configuration)
	if err != nil {
		return err
	}

	apiRules, err := resources.NewApiRule(configuration, c.kymaHostAddress())
	if err != nil {
		return err
	}

	if configuration.Source.Type == workspace.SourceTypeGit {
		gitRepository, err := resources.NewPublicGitRepository(configuration)
		if err != nil {
			return errors.Wrap(err, "Unable to read the Git repository from the provided configuration")
		}
		mgr.AddParent(operator.NewGenericOperator(client.Resource(operator.GVRGitRepository).Namespace(configuration.Namespace), gitRepository), nil)
	}

	mgr.AddParent(
		operator.NewGenericOperator(client.Resource(operator.GVRFunction).Namespace(configuration.Namespace), function),
		[]operator.Operator{
			operator.NewSubscriptionOperator(client.Resource(operator.GVRSubscription).Namespace(configuration.Namespace),
				configuration.Name, configuration.Namespace, subscriptions...),
			operator.NewApiRuleOperator(client.Resource(operator.GVRApiRule).Namespace(configuration.Namespace),
				configuration.Name, apiRules...),
		},
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

	return mgr.Do(ctx, options)
}

func (c *command) kymaHostAddress() string {
	var apiserverURL string
	vs, err := c.K8s.Istio().NetworkingV1alpha3().VirtualServices("kyma-system").Get(context.Background(), "apiserver-proxy", v1.GetOptions{})
	switch {
	case err != nil:
		fmt.Printf("Unable to read the Kyma host URL due to error: %s. \n%s\r\n", err.Error(),
			"Check if your cluster is available and has Kyma installed.")
	case vs != nil && len(vs.Spec.Hosts) > 0:
		apiserverURL = strings.Trim(vs.Spec.Hosts[0], "apiserver.")
	default:
		fmt.Println("Kyma host URL could not be obtained.")
	}

	return apiserverURL
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
