package console

import (
	"fmt"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/pkg/browser"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type command struct {
	opts *Options
	cli.Command
}

//NewCmd creates a new console command
func NewCmd(o *Options) *cobra.Command {

	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:     "console",
		Short:   "Open the Kyma Console in a web browser.",
		Long: `Use this command to open the Kyma Console in a web browser.

### Usage`,

		RunE:    func(_ *cobra.Command, _ []string) error { return c.Run() },
		Aliases: []string{"c"},
	}
    cmd.Flags().Bool("help", false, "Displays help for the command.")
	return cmd
}

//Run runs the command
func (c *command) Run() error {
	var err error
	if c.K8s, err = kube.NewFromConfig("", c.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	fmt.Println("Reading the Kyma console URL from the cluster")

	var consoleURL string
	vs, err := c.K8s.Istio().NetworkingV1alpha3().VirtualServices("kyma-system").Get("core-console", metav1.GetOptions{})
	switch {
	case err != nil:
		fmt.Printf("Unable to read the Kyma console URL due to error: %s. Check if your cluster is available and has Kyma installed\r\n", err.Error())
		return nil
	case vs != nil && vs.Spec != nil && len(vs.Spec.Hosts) > 0:
		consoleURL = fmt.Sprintf("https://%s", vs.Spec.Hosts[0])
	default:
		fmt.Println("Kyma console URL could not be obtained.")
		return nil
	}

	fmt.Println("Opening the Kyma console in the default browser")
	err = browser.OpenURL(consoleURL)
	if err != nil {
		fmt.Println("Failed to open the Kyma console")
		fmt.Printf("Visit %s in your browser to view the console\n", consoleURL)
	}

	return nil
}
