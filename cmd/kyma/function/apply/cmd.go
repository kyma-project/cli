package apply

import (
	"fmt"
	"github.com/kyma-project/cli/internal/cli"
	serverless "github.com/kyma-project/cli/internal/resources/unstructured"
	"github.com/kyma-project/cli/internal/workspace"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
	"os"
	"path"
	"path/filepath"
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
		Use:   "apply",
		Short: "Apply something",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.run()
		},
	}

	cmd.Flags().StringVarP(&o.Dir, "dir", "d", "", `Location of the generated project.`)
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", false, `Only show output.`)

	return cmd
}

func (c *command) run() error {
	if c.opts.Dir == "" {
		var err error
		c.opts.Dir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	file, err := os.Open(path.Join(c.opts.Dir, workspace.CfgFilename))
	if err != nil {
		return err
	}

	// Load project configuration
	var configuration workspace.Cfg
	if err := yaml.NewDecoder(file).Decode(&configuration); err != nil {
		return err
	}
	configuration.SourcePath = c.opts.Dir

	objects := serverless.NewTriggers(configuration)

	obj, err := serverless.NewFunction(configuration)
	if err != nil {
		return err
	}
	objects = append(objects, obj)

	// If --dry-run
	if c.opts.DryRun {
		return printObjects(objects)
	}

	// Update cluster
	return applyProject(c.Command, configuration, objects)
}

var groupResourceVersionFunction = schema.GroupVersionResource{
	Group:    "serverless.kyma-project.io",
	Version:  "v1alpha1",
	Resource: "functions",
}

var groupResourceVersionTrigger = schema.GroupVersionResource{
	Group:    "eventing.knative.dev",
	Version:  "v1alpha1",
	Resource: "triggers",
}

func applyProject(cmd cli.Command, configuration workspace.Cfg, objects []unstructured.Unstructured) error {
	client, err := client(cmd.KubeconfigPath)
	if err != nil {
		return err
	}
	functionClient := client.Resource(groupResourceVersionFunction).Namespace(configuration.Namespace)
	triggerClient := client.Resource(groupResourceVersionTrigger).Namespace(configuration.Namespace)

	for _, obj := range objects {
		switch obj.GroupVersionKind() {
		case groupResourceVersionFunction.GroupVersion().WithKind("Function"):
			err := applyObject(cmd, functionClient, obj)
			if err != nil {
				return err
			}
		case groupResourceVersionTrigger.GroupVersion().WithKind("Trigger"):
			err := applyObject(cmd, triggerClient, obj)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func applyObject(cmd cli.Command, client dynamic.ResourceInterface, obj unstructured.Unstructured) error {
	step := cmd.NewStep(fmt.Sprintf("Applying %s - %s", obj.GetKind(), obj.GetName()))
	// Check if object exists
	response, err := client.Get(obj.GetName(), v1.GetOptions{})
	fnFound := !errors.IsNotFound(err)
	if err != nil && fnFound {
		step.Failure()
		return err
	}

	// If object is up to date return
	var equal bool
	if fnFound {
		equal = equality.Semantic.DeepDerivative(response.Object["spec"], obj.Object["spec"])
	}

	if fnFound && equal {
		step.Successf("%s %s is up to date", obj.GetKind(), obj.GetName())
		return nil
	}

	// If object needs update
	if fnFound && !equal {
		response.Object["spec"] = obj.Object["spec"]
		err = retry.RetryOnConflict(retry.DefaultRetry, func() (err error) {
			_, err = client.Update(response, v1.UpdateOptions{})
			return err
		})

		if err != nil {
			step.Failure()
			return err
		}

		step.Successf("%s %s updated", obj.GetKind(), obj.GetName())
		return nil
	}

	if _, err = client.Create(&obj, v1.CreateOptions{}); err != nil {
		step.Failure()
		return err
	}

	step.Successf("%s %s created", obj.GetKind(), obj.GetName())
	return nil
}

func client(kubeconfig string) (dynamic.Interface, error) {
	home := homedir.HomeDir()

	if kubeconfig == "" && home == "" {
		return nil, fmt.Errorf("unable to find kubeconfig file")
	}

	if kubeconfig == "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(config)
}

const (
	dryRunFormat = "%s\n---\n"
)

func printObjects(objects []unstructured.Unstructured) error {
	for _, object := range objects {
		b, err := yaml.Marshal(object.Object)
		if err != nil {
			return err
		}
		fmt.Printf(dryRunFormat, string(b))
	}
	return nil
}