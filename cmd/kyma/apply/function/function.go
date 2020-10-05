package function

import (
	"context"
	"encoding/json"
	"fmt"
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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
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

	cmd.Flags().VarP(&o.Filename, "filename", "f", `Full path to the config file.`)
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", false, `Validated list of objects to be created from sources.`)
	cmd.Flags().Var(&o.OnError, "onerror", `Flag used to define reaction to the error. Use one of: 
- nothing
- purge`)
	cmd.Flags().VarP(&o.Output, "output", "o", `Flag used to define output of the command. Use one of:
- text
- json
- yaml
- none`)

	return cmd
}

func (c *command) Run() error {
	if c.opts.Filename.String() == "" {
		c.opts.Filename = defaultFilename()
	}

	file, err := os.Open(c.opts.Filename.String())
	if err != nil {
		return err
	}

	// Load project configuration
	var configuration workspace.Cfg
	if err := yaml.NewDecoder(file).Decode(&configuration); err != nil {
		return errors.Wrap(err, "Could not decode configuration file")
	}

	if c.K8s, err = kube.NewFromConfig("", c.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}
	client := c.K8s.Dynamic()

	function, err := resources.NewFunction(configuration)
	if err != nil {
		return err
	}

	triggers, err := resources.NewTriggers(configuration)
	if err != nil {
		return err
	}

	operators := map[operator.Operator][]operator.Operator{
		operator.NewGenericOperator(client.Resource(operator.GVKFunction).Namespace(configuration.Namespace), function): {
			operator.NewTriggersOperator(client.Resource(operator.GVKTriggers).Namespace(configuration.Namespace), triggers...),
		},
	}

	if configuration.Source.Type == workspace.SourceTypeGit {
		gitRepository, err := resources.NewPublicGitRepository(configuration)
		if err != nil {
			return errors.Wrap(err, "Unable to read git repository from configuration")
		}
		gitOperator := operator.NewGenericOperator(client.Resource(operator.GVRGitRepository).Namespace(configuration.Namespace), gitRepository)
		operators[gitOperator] = nil
	}

	mgr := manager.NewManager(operators)
	options := manager.Options{
		Callbacks:          callbacks(c),
		OnError:            chooseOnError(c.opts.OnError),
		DryRun:             c.opts.DryRun,
		SetOwnerReferences: true, // add one more flag to set it to false or add: c.opts.Output != YAMLOutput && c.opts.Output != JSONOutput
	}

	return mgr.Do(context.Background(), options)
}

const (
	operatingFormat     = "%s - %s operating... %s"
	createdFormat       = "%s - %s created %s"
	updatedFormat       = "%s - %s updated %s"
	skippedFormat       = "%s - %s unchanged %s"
	deletedFormat       = "%s - %s deleted %s"
	failedFormat        = "%s - %s can't be removed %s"
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
	case client.StatusTypeFailed:
		return failedFormat
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
