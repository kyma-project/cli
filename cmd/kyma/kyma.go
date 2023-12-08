package kyma

import (
	"github.com/kyma-project/cli/cmd/kyma/alpha"
	"github.com/kyma-project/cli/cmd/kyma/apply"
	"github.com/kyma-project/cli/cmd/kyma/completion"
	"github.com/kyma-project/cli/cmd/kyma/create"
	"github.com/kyma-project/cli/cmd/kyma/dashboard"
	"github.com/kyma-project/cli/cmd/kyma/get"
	imprt "github.com/kyma-project/cli/cmd/kyma/import"
	"github.com/kyma-project/cli/cmd/kyma/import/certs"
	"github.com/kyma-project/cli/cmd/kyma/import/hosts"
	initial "github.com/kyma-project/cli/cmd/kyma/init"
	"github.com/kyma-project/cli/cmd/kyma/provision"
	"github.com/kyma-project/cli/cmd/kyma/provision/gardener"
	"github.com/kyma-project/cli/cmd/kyma/provision/gardener/aws"
	"github.com/kyma-project/cli/cmd/kyma/provision/gardener/az"
	"github.com/kyma-project/cli/cmd/kyma/provision/gardener/gcp"
	"github.com/kyma-project/cli/cmd/kyma/provision/k3d"
	"github.com/kyma-project/cli/cmd/kyma/run"
	"github.com/kyma-project/cli/cmd/kyma/sync"
	"github.com/kyma-project/cli/cmd/kyma/version"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewCmd creates a new Kyma CLI command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kyma",
		Short: "Controls a Kyma cluster.",
		Long: `Kyma is a flexible and easy way to connect and extend enterprise applications in a cloud-native world.
Kyma CLI allows you to install and manage Kyma.

`,
		// Affects children as well
		SilenceErrors: false,
		SilenceUsage:  true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !o.CI {
				cli.CheckForLatestRelease(version.Version)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				_ = err
			}
		},
	}

	cmd.PersistentFlags().BoolVarP(
		&o.Verbose, "verbose", "v", false, "Displays details of actions triggered by the command.",
	)
	cmd.PersistentFlags().BoolVar(
		&o.NonInteractive, "non-interactive", false,
		"Enables the non-interactive shell mode (no colorized output, no spinner).",
	)
	cmd.PersistentFlags().BoolVar(
		&o.CI, "ci", false,
		"Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).",
	)
	// Kubeconfig env var and default paths are resolved by the kyma k8s client using the k8s defined resolution strategy.
	cmd.PersistentFlags().StringVar(
		&o.KubeconfigPath, "kubeconfig", "",
		`Path to the kubeconfig file. If undefined, Kyma CLI uses the KUBECONFIG environment variable, or falls back "/$HOME/.kube/config".`,
	)
	cmd.PersistentFlags().BoolP("help", "h", false, "Provides command help.")

	//Stable commands
	provisionCmd := provision.NewCmd()
	provisionCmd.AddCommand(k3d.NewCmd(k3d.NewOptions(o)))
	gardenerCmd := gardener.NewCmd()
	gardenerCmd.AddCommand(gcp.NewCmd(gcp.NewOptions(o)))
	gardenerCmd.AddCommand(az.NewCmd(az.NewOptions(o)))
	gardenerCmd.AddCommand(aws.NewCmd(aws.NewOptions(o)))
	provisionCmd.AddCommand(gardenerCmd)

	storeCmd := imprt.NewCmd()
	storeCmd.AddCommand(certs.NewCmd(o))
	storeCmd.AddCommand(hosts.NewCmd(o))

	cmd.AddCommand(
		version.NewCmd(version.NewOptions(o)),
		completion.NewCmd(),
		provisionCmd,
		create.NewCmd(o),
		dashboard.NewCmd(dashboard.NewOptions(o)),
		storeCmd,
	)

	cmd.AddCommand(
		initial.NewCmd(o),
		apply.NewCmd(o),
		sync.NewCmd(o),
		run.NewCmd(o),
		get.NewCmd(o),
	)

	cmd.AddCommand(alpha.NewCmd(o))

	return cmd
}
