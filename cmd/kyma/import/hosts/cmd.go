package hosts

import (
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/hosts"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/root"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type command struct {
	opts *cli.Options
	cli.Command
}

//Version contains the cli binary version injected by the build system
var Version string

//NewCmd creates a new kyma command
func NewCmd(o *cli.Options) *cobra.Command {

	cmd := command{
		Command: cli.Command{Options: o},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:   "hosts",
		Short: "Imports the hosts of exposed workloads in the system hosts file.",
		Long: `Use this command to add the hosts of exposed workloads to the hosts file of the local system.
`,
		RunE: func(_ *cobra.Command, _ []string) error { return cmd.Run() },
	}
	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	var err error
	f := step.Factory{
		NonInteractive: true,
	}

	s := f.NewStep("")

	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "failed to initialize the Kubernetes client from given kubeconfig")
	}

	if er := root.IsWithSudo(); er != nil {
		s.LogErrorf("%v", er)
		return nil
	}

	err = hosts.AddDevDomainsToEtcHostsKyma2(s, cmd.K8s)
	if err != nil {
		s.Failure()
		if cmd.opts.Verbose {
			s.LogErrorf("error: %v\n", err)
		}
		return err
	}
	s.Successf("Domain import finished")
	return nil
}
