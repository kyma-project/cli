package version

import (
	"fmt"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
	"os"
)

type command struct {
	opts *Options
	cli.Command
}

//Version contains the cli binary version injected by the build system
var Version string

//NewCmd creates a new kyma command
func NewCmd(o *Options) *cobra.Command {

	cmd := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:   "version",
		Short: "Displays the version of Kyma CLI and of the connected Kyma cluster.",
		Long: `Use this command to print the version of Kyma CLI and the version of the Kyma cluster the current kubeconfig points to.
`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"v"},
	}

	cobraCmd.Flags().BoolVarP(&o.ClientOnly, "client", "c", false, "Client version only (no server required)")
	cobraCmd.Flags().BoolVarP(&o.VersionDetails, "details", "d", false, "Detailed information for each Kyma version")
	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	var w io.Writer = os.Stdout

	fmt.Fprintf(w, "Kyma CLI version: %s\n", getCLIVersion())

	if cmd.opts.ClientOnly {
		return nil
	}

	err := cmd.setKubeClient()
	if err != nil {
		return err
	}

	ver, err := version.GetCurrentKymaVersion(cmd.K8s)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Kyma 2 cluster version: %s\n", ver.String())

	return nil
}

func (cmd *command) setKubeClient() error {
	var err error
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Cannot initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}
	return nil
}

func getCLIVersion() string {
	if len(Version) == 0 {
		return "N/A"
	}
	return Version
}
