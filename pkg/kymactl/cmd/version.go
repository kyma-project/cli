package cmd

import (
	"fmt"

	"github.com/kyma-incubator/kymactl/internal"
	"github.com/spf13/cobra"
)

var Version string

func newVersionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Version of the kymactl and connected Kyma cluster",
		Long: `Prints the version of kymactl itself and the version of the kyma cluster connected by current KUBECONFIG
`,
		RunE: version,
	}
	return versionCmd
}

func version(cmd *cobra.Command, args []string) error {
	if Version == "" {
		fmt.Println("No kymactl version available")
	} else {
		fmt.Printf("kymactl version %s\n", Version)
	}

	kymaVersion, err := internal.RunKubectlCmd([]string{"get", "installation/kyma-installation", "-o", "jsonpath='{.spec.version}'"})
	if err != nil {
		return err
	}
	if kymaVersion == "" {
		fmt.Println("No Kyma cluster version available")
	} else {
		fmt.Printf("Kyma cluster version: %s\n", kymaVersion)
	}
	return nil
}
