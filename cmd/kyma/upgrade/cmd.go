package upgrade

import (
	"fmt"
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/hosts"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/nice"

	"github.com/kyma-project/cli/pkg/installation"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

const (
	defaultDomain = "kyma.local"
	upgradeAction = "upgrade"
)

type command struct {
	opts *Options
	cli.Command
}

//NewCmd creates a new kyma command
func NewCmd(o *Options) *cobra.Command {
	cmd := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrades Kyma to match cli version.",
		Long:  `Use this command to upgrade Kyma on a cluster in order to make its version matching the cli version`,
		RunE:  func(_ *cobra.Command, _ []string) error { return cmd.Run() },
	}

	cobraCmd.Flags().BoolVarP(&o.NoWait, "noWait", "n", false, "Flag that determines if the command should wait for Kyma upgrade to complete.")
	cobraCmd.Flags().StringVarP(&o.Domain, "domain", "d", defaultDomain, "Domain used for upgrade.")
	cobraCmd.Flags().StringVarP(&o.TLSCert, "tlsCert", "", "", "TLS certificate for the domain used for upgrade. The certificate must be a base64-encoded value.")
	cobraCmd.Flags().StringVarP(&o.TLSKey, "tlsKey", "", "", "TLS key for the domain used for upgrade. The key must be a base64-encoded value.")
	cobraCmd.Flags().DurationVarP(&o.Timeout, "timeout", "", 1*time.Hour, "Time-out after which CLI stops watching the upgrade progress.")
	cobraCmd.Flags().StringVarP(&o.Password, "password", "p", "", "Predefined cluster password.")
	cobraCmd.Flags().StringArrayVarP(&o.OverrideConfigs, "override", "o", nil, "Path to a YAML file with parameters to override.")
	cobraCmd.Flags().StringVarP(&o.ComponentsConfig, "components", "c", "", "Path to a YAML file with component list to override.")

	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}

	var err error
	if cmd.K8s, err = kube.NewFromConfigWithTimeout("", cmd.KubeconfigPath, cmd.opts.Timeout); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	s := cmd.NewStep("Reading cluster info from ConfigMap")
	clusterConfig, err := installation.GetClusterInfoFromConfigMap(cmd.K8s)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Cluster info read")

	installationInstance := cmd.configureInstallation(clusterConfig)
	result, err := installationInstance.InstallKyma(upgradeAction)
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}

	if clusterConfig.IsLocal {
		s = cmd.NewStep("Adding domains to /etc/hosts")
		err = hosts.AddDevDomainsToEtcHosts(s, clusterConfig, cmd.K8s, cmd.opts.Verbose, cmd.opts.Timeout, cmd.opts.Domain)
		if err != nil {
			s.Failure()
			return err
		}
		s.Successf("Domains added")
	}

	err = cmd.printSummary(result)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *command) configureInstallation(clusterConfig installation.ClusterInfo) *installation.Installation {
	return &installation.Installation{
		Options: &installation.Options{
			NoWait:           cmd.opts.NoWait,
			Verbose:          cmd.opts.Verbose,
			CI:               cmd.opts.CI,
			NonInteractive:   cmd.Factory.NonInteractive,
			Timeout:          cmd.opts.Timeout,
			KubeconfigPath:   cmd.opts.KubeconfigPath,
			Domain:           cmd.opts.Domain,
			TLSCert:          cmd.opts.TLSCert,
			TLSKey:           cmd.opts.TLSKey,
			Password:         cmd.opts.Password,
			OverrideConfigs:  cmd.opts.OverrideConfigs,
			ComponentsConfig: cmd.opts.ComponentsConfig,
			IsLocal:          clusterConfig.IsLocal,
			LocalCluster: &installation.LocalCluster{
				IP:       clusterConfig.LocalIP,
				Profile:  clusterConfig.Profile,
				Provider: clusterConfig.Provider,
				VMDriver: clusterConfig.LocalVMDriver,
			},
		},
	}
}

func (cmd *command) printSummary(result *installation.Result) error {
	nicePrint := nice.Nice{}
	if cmd.Factory.NonInteractive {
		nicePrint.NonInteractive = true
	}

	fmt.Println()
	nicePrint.PrintKyma()
	fmt.Print(" is upgraded to version:\t")
	nicePrint.PrintImportant(result.KymaVersion)

	return nil
}
