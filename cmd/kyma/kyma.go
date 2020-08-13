package kyma

import (
	"github.com/kyma-project/cli/cmd/kyma/completion"
	"github.com/kyma-project/cli/cmd/kyma/console"
	"github.com/kyma-project/cli/cmd/kyma/install"
	"github.com/kyma-project/cli/cmd/kyma/provision/aks"
	"github.com/kyma-project/cli/cmd/kyma/provision/gardener"
	"github.com/kyma-project/cli/cmd/kyma/provision/gardener/aws"
	"github.com/kyma-project/cli/cmd/kyma/provision/gardener/az"
	"github.com/kyma-project/cli/cmd/kyma/provision/gardener/gcp"
	"github.com/kyma-project/cli/cmd/kyma/provision/gke"
	"github.com/kyma-project/cli/cmd/kyma/provision/minikube"
	"github.com/kyma-project/cli/cmd/kyma/test"
	"github.com/kyma-project/cli/cmd/kyma/test/definitions"
	del "github.com/kyma-project/cli/cmd/kyma/test/delete"
	"github.com/kyma-project/cli/cmd/kyma/test/list"
	"github.com/kyma-project/cli/cmd/kyma/test/logs"
	"github.com/kyma-project/cli/cmd/kyma/test/run"
	"github.com/kyma-project/cli/cmd/kyma/test/status"
	"github.com/kyma-project/cli/cmd/kyma/version"

	"github.com/kyma-project/cli/cmd/kyma/provision"
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
	cmd.PersistentFlags().BoolVar(&o.CI, "ci", false, "Enables the CI mode to run on CI/CD systems.")
	// Kubeconfig env var and default paths are resolved by the kyma k8s client using the k8s defined resolution strategy.
	cmd.PersistentFlags().StringVar(&o.KubeconfigPath, "kubeconfig", "", `Specifies the path to the kubeconfig file. By default, Kyma CLI uses the KUBECONFIG environment variable or "/$HOME/.kube/config" if the variable is not set.`)
	cmd.PersistentFlags().BoolP("help", "h", false, "Displays help for the command.")

	provisionCmd := provision.NewCmd()
	provisionCmd.AddCommand(minikube.NewCmd(minikube.NewOptions(o)))
	provisionCmd.AddCommand(gke.NewCmd(gke.NewOptions(o)))
	provisionCmd.AddCommand(aks.NewCmd(aks.NewOptions(o)))
	gardenerCmd := gardener.NewCmd()
	gardenerCmd.AddCommand(gcp.NewCmd(gcp.NewOptions(o)))
	gardenerCmd.AddCommand(az.NewCmd(az.NewOptions(o)))
	gardenerCmd.AddCommand(aws.NewCmd(aws.NewOptions(o)))
	provisionCmd.AddCommand(gardenerCmd)

	cmd.AddCommand(
		version.NewCmd(version.NewOptions(o)),
		completion.NewCmd(),
		install.NewCmd(install.NewOptions(o)),
		provisionCmd,
		console.NewCmd(console.NewOptions(o)),
	)

	testCmd := test.NewCmd()
	testRunCmd := run.NewCmd(run.NewOptions(o))
	testStatusCmd := status.NewCmd(status.NewOptions(o))
	testDeleteCmd := del.NewCmd(del.NewOptions(o))
	testListCmd := list.NewCmd(list.NewOptions(o))
	testDefsCmd := definitions.NewCmd(definitions.NewOptions(o))
	testLogsCmd := logs.NewCmd(logs.NewOptions(o))
	testCmd.AddCommand(testRunCmd, testStatusCmd, testDeleteCmd, testListCmd, testDefsCmd, testLogsCmd)
	cmd.AddCommand(testCmd)

	return cmd
}
