package cmd

import (
	"github.com/kyma-project/cli/pkg/kyma/cmd/install"
	"github.com/kyma-project/cli/pkg/kyma/cmd/provision/minikube"
	"github.com/kyma-project/cli/pkg/kyma/cmd/test"
	del "github.com/kyma-project/cli/pkg/kyma/cmd/test/delete"
	"github.com/kyma-project/cli/pkg/kyma/cmd/test/list"
	"github.com/kyma-project/cli/pkg/kyma/cmd/test/run"
	"github.com/kyma-project/cli/pkg/kyma/cmd/test/status"
	"github.com/kyma-project/cli/pkg/kyma/cmd/uninstall"

	"github.com/kyma-project/cli/pkg/kyma/cmd/provision"
	"github.com/kyma-project/cli/pkg/kyma/core"
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

Find more information at: https://github.com/kyma-project/cli
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

	uninstallCmd := uninstall.NewCmd(uninstall.NewOptions(o))
	cmd.AddCommand(uninstallCmd)

	testCmd := test.NewCmd()
	testRunCmd := run.NewCmd(run.NewOptions(o))
	testStatusCmd := status.NewCmd(status.NewOptions(o))
	testDeleteCmd := del.NewCmd(del.NewOptions(o))
	testListCmd := list.NewCmd(list.NewOptions(o))
	testCmd.AddCommand(testRunCmd, testStatusCmd, testDeleteCmd, testListCmd)
	cmd.AddCommand(testCmd)

	return cmd
}
