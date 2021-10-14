package hostfiles

import (
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/hosts"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/root"
	"github.com/kyma-project/cli/pkg/installation"
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
		Short: "Adds domain entries to the system host file.",
		Long: `Use this command to add domain to the host file of the local system.
`,
		RunE: func(_ *cobra.Command, _ []string) error { return cmd.Run() },
	}
	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	var err error
	s := cmd.NewStep("")
	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}

	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid.")
	}

	if !root.IsWithSudo() {
		s.LogError("Could not add entries to host file. Make sure you are using sudo.")
		return nil
	}

	clusterConfig, err := installation.GetClusterInfoFromConfigMap(cmd.K8s)
	if err != nil {
		return errors.Wrap(err, "Failed to get cluster information.")
	}

	err = hosts.AddDevDomainsToEtcHosts2(s, clusterConfig, cmd.K8s, defaultDomain)
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
