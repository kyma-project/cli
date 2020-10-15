package function

import (
	"context"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

const (
	defaultNamespace = "default"
	functions = "functions"
)

type command struct {
	opts *Options
	cli.Command
}

//NewCmd creates a new init command
func NewCmd(o *Options) *cobra.Command {
	c := command{
		opts:    o,
		Command: cli.Command{Options: o.Options},
	}
	cmd := &cobra.Command{
		Use:   "function",
		Short: "Saves locally resources for requested Function.",
		Long: `Use this command to create the local workspace with the default structure of your Function's code and dependencies. Update this configuration to your references and apply it to a Kyma cluster. 
Use the flags to specify the initial configuration for your Function or to choose the location for your project.`, //todo
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Run()
		},
	}

	cmd.Flags().StringVar(&o.Name, "name", "", `Function name.`)
	cmd.Flags().StringVarP(&o.Namespace, "namespace","n", defaultNamespace, `Namespace from which you want to sync Function.`)
	cmd.Flags().StringVarP(&o.OutputPath, "output", "o", "", `Full path to the directory where you want to save the project.`)
	cmd.Flags().StringVarP(&o.Kubeconfig, "kubeconfig", "k", "", `Full path to the Kubeconfig file.`)

	return cmd
}

func (c *command) Run() error {
	s := c.NewStep("Generating project structure")

	if err := c.opts.setDefaults(); err != nil {
		s.Failure()
		return err
	}

	if _, err := os.Stat(c.opts.OutputPath); os.IsNotExist(err) {
		err = os.MkdirAll(c.opts.OutputPath, 0700)
		if err != nil {
			return err
		}
	}


	crdConfig, err := clientcmd.BuildConfigFromFlags("", c.opts.Kubeconfig)
	if err != nil {
		return err
	}

	crdConfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: v1alpha1.GroupVersion.Group, Version: v1alpha1.GroupVersion.Version}
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)

	restClient, err := rest.UnversionedRESTClientFor(crdConfig)
	if err != nil {
		panic(err.Error())
	}


	function := &v1alpha1.Function{}
	err = restClient.Get().Resource(functions).Namespace(c.opts.Namespace).Name(c.opts.Name).Do(context.Background()).Into(function)
	if err != nil {
		panic(err.Error())
	}

	configuration := workspace.Cfg{
		Name:      c.opts.Name,
		Namespace: c.opts.Namespace,
	}

	err = workspace.Synchronise(configuration,c.opts.OutputPath,*function,restClient)
	if err != nil {
		s.Failure()
		return err
	}

	s.Successf("Function synchronised in %s", c.opts.OutputPath)
	return nil
}
