package version

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
	"github.com/kyma-project/cli/cmd/kyma/version"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	installConfig "github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
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
		Short: "Displays the version of Kyma CLI and of the connected Kyma cluster.",
		Long: `Use this command to print the version of Kyma CLI and the version of the Kyma cluster the current kubeconfig points to.
`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"v"},
	}

	cobraCmd.Flags().BoolVarP(&o.ClientOnly, "client", "c", false, "Client version only (no server required)")
	cobraCmd.Flags().BoolVarP(&o.VersionDetails, "details", "d", false, "Detailed information for each Kyma version")
	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	var w io.Writer = os.Stdout

	cmd.printCliVersion(w)

	if cmd.opts.ClientOnly {
		//we are done
		return nil
	}

	//print Kyma Version
	provider, err := cmd.metadataProvider()
	if err != nil {
		return err
	}
	versionSet, err := provider.Versions()
	if err != nil {
		return fmt.Errorf("Unable to get Kyma cluster versions due to error: %v. Check if your cluster is available and has Kyma installed", err)
	}

	cmd.printKymaVersion(w, versionSet)

	if cmd.opts.VersionDetails {
		cmd.printKymaVersionDetails(os.Stdout, versionSet)
	}

	return nil
}

func (cmd *command) printCliVersion(w io.Writer) {
	fmt.Fprintf(w, "Kyma CLI version: %s\n", versionOrDefault(version.Version))
}

func (cmd *command) printKymaVersion(w io.Writer, versionSet *helm.KymaVersionSet) {
	fmt.Fprintf(w, "Kyma cluster versions: %s\n", versionOrDefault(strings.Join(versionSet.Names(), ", ")))
}

func (cmd *command) printKymaVersionDetails(w io.Writer, versionSet *helm.KymaVersionSet) {
	for _, version := range versionSet.Versions {
		fmt.Fprintln(w, "-----------------")
		fmt.Fprintf(w, "Kyma cluster version: %s\n", versionOrDefault(version.Version))
		deployTime := time.Unix(version.CreationTime, 0)
		fmt.Fprintf(w, "Deployed at: %s\n", deployTime.UTC().Format(time.RFC850))
		fmt.Fprintf(w, "Profile: %s\n", profileOrDefault(version.Profile))
		fmt.Fprintf(w, "Components: %s\n", strings.Join(version.ComponentNames(), ", "))
	}
}

func (cmd *command) metadataProvider() (*helm.KymaMetadataProvider, error) {
	var err error
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return nil, errors.Wrap(err, "Cannot initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}
	return helm.NewKymaMetadataProvider(installConfig.KubeconfigSource{
		Path: cmd.KubeconfigPath,
	})
}

func versionOrDefault(version string) string {
	return stringOrDefault(version, "N/A")
}

func profileOrDefault(profile string) string {
	return stringOrDefault(profile, "default")
}

func stringOrDefault(s, def string) string {
	if len(s) == 0 {
		return def
	}
	return s
}
