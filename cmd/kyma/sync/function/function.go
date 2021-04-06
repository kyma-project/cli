package function

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
		Short: "Synchronizes the local resources for your Function.",
		Long: `Use this command to download the Function's code and dependencies from the cluster to create or update these resources in your local workspace.
Use the flags to specify the name of your Function, the Namespace, or the location for your project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Run(args[0])
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("missing name of the function")
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&o.Namespace, "namespace", "n", "", `Namespace from which you want to sync the Function.`)
	cmd.Flags().StringVarP(&o.Dir, "dir", "d", "", `Full path to the directory where you want to save the project.`)

	return cmd
}

func (c *command) Run(name string) error {
	s := c.NewStep("Generating project structure")
	var err error
	if c.K8s, err = kube.NewFromConfig("", c.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	if err := c.opts.setDefaults(c.K8s.DefaultNamespace()); err != nil {
		s.Failure()
		return err
	}

	if _, err := os.Stat(c.opts.Dir); os.IsNotExist(err) {
		err = os.MkdirAll(c.opts.Dir, 0700)
		if err != nil {
			return err
		}
	}

	ctx := context.Background()

	var buildClient client.Build = func(namespace string, resource schema.GroupVersionResource) client.Client {
		return c.K8s.Dynamic().Resource(resource).Namespace(namespace)
	}

	cfg := workspace.Cfg{
		Name:      name,
		Namespace: c.opts.Namespace,
	}

	err = workspace.Synchronise(ctx, cfg, c.opts.Dir, buildClient, c.kymaHostAddress())
	if err != nil {
		s.Failure()
		return err
	}

	s.Successf("Function synchronised in %s", c.opts.Dir)
	return nil
}

func (c *command) kymaHostAddress() string {
	var apiserverURL string
	vs, err := c.K8s.Istio().NetworkingV1alpha3().VirtualServices("kyma-system").Get(context.Background(), "apiserver-proxy", v1.GetOptions{})
	switch {
	case err != nil:
		fmt.Printf("Unable to read the Kyma host URL due to error: %s. \n%s\n%s\r\n", err.Error(),
			"Check if your cluster is available and has Kyma installed.",
			"If apirules host is provided in the configuration manually, ignore this message.")
	case vs != nil && len(vs.Spec.Hosts) > 0:
		apiserverURL = strings.Trim(vs.Spec.Hosts[0], "apiserver.")
	default:
		fmt.Println("Kyma host URL could not be obtained.")
	}

	return apiserverURL
}
