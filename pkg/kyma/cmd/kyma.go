package cmd

import (
	"github.com/kyma-incubator/kymactl/pkg/kyma/cmd/install"
	"github.com/kyma-incubator/kymactl/pkg/kyma/cmd/install/cluster"
	"github.com/kyma-incubator/kymactl/pkg/kyma/cmd/uninstall"
	"github.com/spf13/cobra"
)

//KymaOptions defines available options for the command
type KymaOptions struct {
	Verbose bool
}

//NewKymaOptions creates options with default values
func NewKymaOptions() *KymaOptions {
	return &KymaOptions{}
}

//NewKymaCmd creates a new kyma CLI command
func NewKymaCmd(o *KymaOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kyma",
		Short: "Controls a Kyma cluster.",
		Long: `Kyma is a flexible and easy way to connect and extend enterprise applications in a cloud-native world.
kyma CLI controls a Kyma cluster.

Find more information at: https://github.com/kyma-incubator/kymactl
`,
		// Affects children as well
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	cmd.PersistentFlags().BoolVarP(&o.Verbose, "verbose", "v", false, "verbose output")

	versionCmd := NewVersionCmd(NewVersionOptions())
	cmd.AddCommand(versionCmd)

	completionCmd := NewCompletionCmd()
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

	testCmd := NewTestCmd()
	cmd.AddCommand(testCmd)

	return cmd
}
