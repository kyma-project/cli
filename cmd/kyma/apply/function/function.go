package function

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"github.com/kyma-incubator/hydroform/function/pkg/manager"
	"github.com/kyma-incubator/hydroform/function/pkg/operator"
	resources "github.com/kyma-incubator/hydroform/function/pkg/resources/unstructured"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Operators = map[operator.Operator][]operator.Operator

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
	cmd.Flags().StringVar(&o.OnError, "onerror", NothingOnError, `Flag used to define reaction to the error. Use one of:
- nothing
- purge`)
	cmd.Flags().StringVarP(&o.Output, "output", "o", TextOutput, `Flag used to define output of the command. Use one of:
- text
- json
- yaml
- none`)

	return cmd
}

func (c *command) Run() error {
	err := c.defaultFilename()
	if err != nil {
		return errors.Wrap(err, "Could not default all flags")
	}

	err = c.checkRequirements()
	if err != nil {
		return errors.Wrap(err, "Could not parse all flags")
	}

	file, err := os.Open(c.opts.Filename)
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

	operators := Operators{
		operator.GenericOperator(client.Resource(operator.GVKFunction).Namespace(configuration.Namespace), function): {
			operator.NewTriggersOperator(client.Resource(operator.GVKTriggers).Namespace(configuration.Namespace), triggers...),
		},
	}

	if configuration.Source.Type == workspace.SourceTypeGit {
		gitRepository, err := resources.NewPublicGitRepository(configuration)
		if err != nil {
			return errors.Wrap(err, "Unable to read git repository from configuration")
		}
		gitOperator := operator.GenericOperator(client.Resource(operator.GVRGitRepository).Namespace(configuration.Namespace), gitRepository)
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

func (c *command) checkRequirements() error {
	if !isOneOf(c.opts.Output, validOutput) {
		return fmt.Errorf("specified value of Output flag: '%s' is not supported", c.opts.Output)
	}
	if !isOneOf(c.opts.OnError, validOnError) {
		return fmt.Errorf("specified value of OnError flag: '%s' is not supported", c.opts.OnError)
	}

	return nil
}

func (c *command) defaultFilename() error {
	if c.opts.Filename == "" {
		var err error
		c.opts.Filename, err = os.Getwd()
		if err != nil {
			return err
		}
		c.opts.Filename = path.Join(c.opts.Filename, workspace.CfgFilename)
	} else {
		fileInfo, err := os.Stat(c.opts.Filename)
		if err != nil {
			return err
		}

		if fileInfo.Mode().IsDir() {
			c.opts.Filename = path.Join(c.opts.Filename, workspace.CfgFilename)
		}
	}
	return nil
}

func isOneOf(item string, slice []string) bool {
	for _, elem := range slice {
		if elem == item {
			return true
		}
	}
	return false
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

func chooseOnError(onErr OnError) manager.OnError {
	if onErr == NothingOnError {
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

	switch l.opts.Output {
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

	if l.opts.Output != TextOutput {
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
