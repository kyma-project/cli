package cmd

import (
	"time"

	"github.com/kyma-incubator/kyma-cli/pkg/kyma/cmd/prepare"
	"github.com/kyma-incubator/kyma-cli/pkg/kyma/core"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	sleep = 10 * time.Second
)

//NewKymaCmd creates a new kyma CLI command
func NewKymaCmd(o *core.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kyma",
		Short: "Controls a Kyma cluster.",
		Long: `Kyma is a flexible and easy way to connect and extend enterprise applications in a cloud-native world.
kyma CLI controls a Kyma cluster.

Find more information at: https://github.com/kyma-incubator/kyma-cli
`,
		// Affects children as well
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	cmd.PersistentFlags().BoolVarP(&o.Verbose, "verbose", "v", false, "verbose output")
	cmd.PersistentFlags().BoolVar(&o.NonInteractive, "non-interactive", false, "Do not use spinners")
	cmd.PersistentFlags().StringVar(&o.KubeconfigPath, "kubeconfig", clientcmd.RecommendedHomeFile, "Path to kubeconfig")

	versionCmd := NewVersionCmd(NewVersionOptions(o))
	cmd.AddCommand(versionCmd)

	completionCmd := NewCompletionCmd()
	cmd.AddCommand(completionCmd)

	prepareCmd := prepare.NewCmd()
	cmd.AddCommand(prepareCmd)
	prepareMinikubeCmd := prepare.NewMinikubeCmd(prepare.NewMinikubeOptions(o))
	prepareCmd.AddCommand(prepareMinikubeCmd)

	installCmd := NewInstallCmd(NewInstallOptions(o))
	cmd.AddCommand(installCmd)

	uninstallCmd := NewUninstallCmd(NewUninstallOptions(o))
	cmd.AddCommand(uninstallCmd)

	testCmd := NewTestCmd(NewTestOptions(o))
	cmd.AddCommand(testCmd)

	return cmd
}
