package cmd

import (
	"github.com/kyma-project/cli/internal/cmd/completion"
	"github.com/kyma-project/cli/internal/cmd/console"
	"github.com/kyma-project/cli/internal/cmd/install"
	"github.com/kyma-project/cli/internal/cmd/provision/minikube"
	"github.com/kyma-project/cli/internal/cmd/test"
	"github.com/kyma-project/cli/internal/cmd/test/definitions"
	del "github.com/kyma-project/cli/internal/cmd/test/delete"
	"github.com/kyma-project/cli/internal/cmd/test/list"
	"github.com/kyma-project/cli/internal/cmd/test/run"
	"github.com/kyma-project/cli/internal/cmd/test/status"
	"github.com/kyma-project/cli/internal/cmd/uninstall"
	"github.com/kyma-project/cli/internal/cmd/version"

	"github.com/kyma-project/cli/internal/cmd/provision"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

//NewCmd creates a new kyma CLI command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kyma",
		Short: "Controls a Kyma cluster.",
		Long: `Kyma is a flexible and easy way to connect and extend enterprise applications in a cloud-native world.
Kyma CLI allows you to install, test, and manage Kyma.

For more information, see: https://github.com/kyma-project/cli
`,
		// Affects children as well
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	cmd.PersistentFlags().BoolVarP(&o.Verbose, "verbose", "v", false, "Displays details of actions triggered by the command.")
	cmd.PersistentFlags().BoolVar(&o.NonInteractive, "non-interactive", false, "Enables the non-interactive shell mode.")
	// Kubeconfig env var and defualt paths are resolved by the kyma k8s client using the k8s defined resolution strategy.
	cmd.PersistentFlags().StringVar(&o.KubeconfigPath, "kubeconfig", "", "Specifies the path to the kubeconfig file.")
	cmd.Flags().Bool("help", false, "Displays help for the command.")

	provisionCmd := provision.NewCmd()
	provisionCmd.AddCommand(minikube.NewCmd(minikube.NewOptions(o)))

	cmd.AddCommand(
		version.NewCmd(version.NewOptions(o)),
		completion.NewCmd(),
		install.NewCmd(install.NewOptions(o)),
		uninstall.NewCmd(uninstall.NewOptions(o)),
		provisionCmd,
		console.NewCmd(console.NewOptions(o)),
	)

	testCmd := test.NewCmd()
	testRunCmd := run.NewCmd(run.NewOptions(o))
	testStatusCmd := status.NewCmd(status.NewOptions(o))
	testDeleteCmd := del.NewCmd(del.NewOptions(o))
	testListCmd := list.NewCmd(list.NewOptions(o))
	testDefsCmd := definitions.NewCmd(definitions.NewOptions(o))
	testCmd.AddCommand(testRunCmd, testStatusCmd, testDeleteCmd, testListCmd, testDefsCmd)
	cmd.AddCommand(testCmd)

	return cmd
}
