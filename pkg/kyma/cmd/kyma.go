package cmd

import (
	"github.com/kyma-incubator/kyma-cli/pkg/kyma/cmd/install"
	"github.com/kyma-incubator/kyma-cli/pkg/kyma/cmd/provision/minikube"
	"github.com/kyma-incubator/kyma-cli/pkg/kyma/cmd/uninstall"
	"github.com/kyma-incubator/kyma-cli/pkg/kyma/cmd/update"

	"github.com/kyma-incubator/kyma-cli/pkg/kyma/cmd/provision"
	"github.com/kyma-incubator/kyma-cli/pkg/kyma/core"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
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

	provisionCmd := provision.NewCmd()
	cmd.AddCommand(provisionCmd)
	provisionMinikubeCmd := minikube.NewCmd(minikube.NewOptions(o))
	provisionCmd.AddCommand(provisionMinikubeCmd)

	installCmd := install.NewCmd(install.NewOptions(o))
	cmd.AddCommand(installCmd)

	updateCmd := update.NewCmd(update.NewOptions(o))
	cmd.AddCommand(updateCmd)

	uninstallCmd := uninstall.NewCmd(uninstall.NewOptions(o))
	cmd.AddCommand(uninstallCmd)

	testCmd := NewTestCmd(NewTestOptions(o))
	cmd.AddCommand(testCmd)

	return cmd
}
