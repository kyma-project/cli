package hostfiles

import (
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/hosts"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/root"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	defaultDomain = "kyma.local"
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
		Use:   "host-entries",
		Short: "Imports domain entries in the system host file.",
		Long: `Use this command to add domain to the host file of the local system.
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
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid.")
	}

	if !root.IsWithSudo() {
		s.LogError("Elevated permissions are required to make entries to host file. Make sure you are using sudo.")
		return nil
	}

	err = hosts.AddDevDomainsToEtcHostsKyma2(s, cmd.K8s, defaultDomain)
	if err != nil {
		s.Failure()
		if cmd.opts.Verbose {
			s.LogErrorf("error: %v\n", err)
		}
		return err
	}
	s.Successf("Domains added")
	return nil
}
