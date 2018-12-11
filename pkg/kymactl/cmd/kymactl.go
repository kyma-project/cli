package cmd

import (
	"github.com/kyma-incubator/kymactl/pkg/kymactl/cmd/install"
	"github.com/kyma-incubator/kymactl/pkg/kymactl/cmd/install/cluster"
	"github.com/kyma-incubator/kymactl/pkg/kymactl/cmd/uninstall"
	"github.com/spf13/cobra"
)

type KymactlOptions struct {
	Verbose bool
}

func NewKymactlOptions() *KymactlOptions {
	return &KymactlOptions{}
}

func NewKymactlCmd(o *KymactlOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kymactl",
		Short: "kymactl controls a Kyma cluster.",
		Long: `Kyma is a flexible and easy way to connect and extend enterprise applications in a cloud-native world.
kymactl controls a Kyma cluster.

Find more information at: https://github.com/kyma-incubator/kymactl
`,
		// Affects children as well
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	cmd.PersistentFlags().BoolVarP(&o.Verbose, "verbose", "v", false, "verbose output")

	versionCmd := newVersionCmd()
	cmd.AddCommand(versionCmd)

	completionCmd := newCompletionCmd()
	cmd.AddCommand(completionCmd)

	installCmd := install.NewCmd()
	installClusterCmd := cluster.NewCmd()
	installCmd.AddCommand(installClusterCmd)
	installClusterMinikubeCmd := cluster.NewMinikubeCmd(cluster.NewMinikubeOptions())
	installClusterCmd.AddCommand(installClusterMinikubeCmd)
	installKymaCmd := install.NewKymaCmd(install.NewKymaOptions())
	installCmd.AddCommand(installKymaCmd)
	cmd.AddCommand(installCmd)

	uninstallCmd := uninstall.NewCmd()
	uninstallKymaCmd := uninstall.NewKymaCmd(uninstall.NewKymaOptions())
	uninstallCmd.AddCommand(uninstallKymaCmd)
	cmd.AddCommand(uninstallCmd)

	return cmd
}
