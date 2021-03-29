package kyma

import (
	"github.com/kyma-project/cli/cmd/kyma/alpha"
	alphaDelete "github.com/kyma-project/cli/cmd/kyma/alpha/delete"
	alphaInstall "github.com/kyma-project/cli/cmd/kyma/alpha/deploy"
	alphaProvision "github.com/kyma-project/cli/cmd/kyma/alpha/provision"
	"github.com/kyma-project/cli/cmd/kyma/alpha/provision/k3s"
	alphaVersion "github.com/kyma-project/cli/cmd/kyma/alpha/version"
	"github.com/kyma-project/cli/cmd/kyma/apply"
	"github.com/kyma-project/cli/cmd/kyma/completion"
	"github.com/kyma-project/cli/cmd/kyma/console"
	"github.com/kyma-project/cli/cmd/kyma/create"
	initial "github.com/kyma-project/cli/cmd/kyma/init"
	"github.com/kyma-project/cli/cmd/kyma/install"
	"github.com/kyma-project/cli/cmd/kyma/provision/aks"
	"github.com/kyma-project/cli/cmd/kyma/provision/gardener"
	"github.com/kyma-project/cli/cmd/kyma/provision/gardener/aws"
	"github.com/kyma-project/cli/cmd/kyma/provision/gardener/az"
	"github.com/kyma-project/cli/cmd/kyma/provision/gardener/gcp"
	"github.com/kyma-project/cli/cmd/kyma/provision/gke"
	"github.com/kyma-project/cli/cmd/kyma/provision/minikube"
	"github.com/kyma-project/cli/cmd/kyma/run"
	"github.com/kyma-project/cli/cmd/kyma/sync"
	"github.com/kyma-project/cli/cmd/kyma/test"
	testdefs "github.com/kyma-project/cli/cmd/kyma/test/definitions"
	testdel "github.com/kyma-project/cli/cmd/kyma/test/delete"
	testlist "github.com/kyma-project/cli/cmd/kyma/test/list"
	testlogs "github.com/kyma-project/cli/cmd/kyma/test/logs"
	testrun "github.com/kyma-project/cli/cmd/kyma/test/run"
	teststatus "github.com/kyma-project/cli/cmd/kyma/test/status"
	"github.com/kyma-project/cli/cmd/kyma/version"

	"github.com/kyma-project/cli/cmd/kyma/provision"
	"github.com/kyma-project/cli/cmd/kyma/upgrade"
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

`,
		// Affects children as well
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	cmd.PersistentFlags().BoolVarP(&o.Verbose, "verbose", "v", false, "Displays details of actions triggered by the command.")
	cmd.PersistentFlags().BoolVar(&o.NonInteractive, "non-interactive", false, "Enables the non-interactive shell mode (no colorized output, no spinner)")
	cmd.PersistentFlags().BoolVar(&o.CI, "ci", false, "Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).")
	// Kubeconfig env var and default paths are resolved by the kyma k8s client using the k8s defined resolution strategy.
	cmd.PersistentFlags().StringVar(&o.KubeconfigPath, "kubeconfig", "", `Path to the kubeconfig file. If undefined, Kyma CLI uses the KUBECONFIG environment variable, or falls back "/$HOME/.kube/config".`)
	cmd.PersistentFlags().BoolP("help", "h", false, "See help for the command")

	//Alpha commands
	alphaCmd := alpha.NewCmd()
	alphaCmd.AddCommand(alphaInstall.NewCmd(alphaInstall.NewOptions(o)))
	alphaCmd.AddCommand(alphaDelete.NewCmd(alphaDelete.NewOptions(o)))
	alphaCmd.AddCommand(alphaVersion.NewCmd(alphaVersion.NewOptions(o)))

	alphaProvisionCmd := alphaProvision.NewCmd()
	alphaProvisionCmd.AddCommand(k3s.NewCmd(k3s.NewOptions(o)))
	alphaCmd.AddCommand(alphaProvisionCmd)

	//Stable commands
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
		alphaCmd,
		version.NewCmd(version.NewOptions(o)),
		completion.NewCmd(),
		install.NewCmd(install.NewOptions(o)),
		provisionCmd,
		console.NewCmd(console.NewOptions(o)),
		upgrade.NewCmd(upgrade.NewOptions(o)),
		create.NewCmd(o),
	)

	testCmd := test.NewCmd()
	testRunCmd := testrun.NewCmd(testrun.NewOptions(o))
	testStatusCmd := teststatus.NewCmd(teststatus.NewOptions(o))
	testDeleteCmd := testdel.NewCmd(testdel.NewOptions(o))
	testListCmd := testlist.NewCmd(testlist.NewOptions(o))
	testDefsCmd := testdefs.NewCmd(testdefs.NewOptions(o))
	testLogsCmd := testlogs.NewCmd(testlogs.NewOptions(o))
	testCmd.AddCommand(testRunCmd, testStatusCmd, testDeleteCmd, testListCmd, testDefsCmd, testLogsCmd)
	cmd.AddCommand(testCmd)

	cmd.AddCommand(
		initial.NewCmd(o),
		apply.NewCmd(o),
		sync.NewCmd(o),
		run.NewCmd(o),
	)

	return cmd
}
