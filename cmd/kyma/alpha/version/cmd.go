package version

import (
	"fmt"
	"io"
	"os"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/metadata"
	"github.com/kyma-project/cli/cmd/kyma/version"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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
		Use:   "version",
		Short: "Displays the version of Kyma CLI and the connected Kyma cluster.",
		Long: `Use this command to print the version of Kyma CLI and the version of the Kyma cluster the current kubeconfig points to.
`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"v"},
	}

	cobraCmd.Flags().BoolVarP(&o.ClientOnly, "client", "c", false, "Client version only (no server required)")
	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	var clusterMetadata *metadata.KymaMetadata

	if !cmd.opts.ClientOnly {
		var err error
		if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
			return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
		}

		provider := metadata.New(cmd.K8s.Static())
		clusterMetadata, err = provider.ReadKymaMetadata()
		if err != nil {
			return fmt.Errorf("Unable to get Kyma cluster version due to error: %v. Check if your cluster is available and has Kyma installed", err)
		}
	}

	printVersion(os.Stdout, cmd.opts.ClientOnly, clusterMetadata)

	return nil
}

func printVersion(w io.Writer, clientOnly bool, clusterMetadata *metadata.KymaMetadata) {
	clientVersion := getVersionOrDefault(version.Version)
	fmt.Fprintf(w, "Kyma CLI version: %s\n", clientVersion)

	if clientOnly {
		return
	}

	serverVersion := getVersionOrDefault(clusterMetadata.Version)
	fmt.Fprintf(w, "Kyma cluster version: %s\n", serverVersion)
}

func getVersionOrDefault(version string) string {
	if len(version) == 0 {
		return "N/A"
	}

	return version
}
