package uninstall

import (
	"fmt"
	"os"
	"time"

	"github.com/kyma-incubator/kymactl/internal"
	"github.com/spf13/cobra"
)

const (
	sleep = 10 * time.Second
)

type KymaOptions struct {
}

func NewKymaOptions() *KymaOptions {
	return &KymaOptions{}
}

func NewKymaCmd(o *KymaOptions) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstalls kyma from a running kubernetes cluster",
		Long: `Uninstall kyma on a running kubernetes cluster.

Assure that your KUBECONFIG is pointing to the target cluster already.
The command will:
- Removes your account as the cluster administrator 
- Removes tiller
- Removes the Kyma installer
`,
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
		Aliases: []string{"i"},
	}

	return cmd
}

func (o *KymaOptions) Run() error {
	fmt.Printf("Uninstalling kyma\n\n")

	var spinner = internal.NewSpinner("Activating kyma-installer to uninstall kyma", "kyma-installer activated to uninstall kyma")
	labelInstaller := []string{"label", "installation/kyma-installation", "action=uninstall"}
	internal.RunKubeCmd(labelInstaller)
	internal.StopSpinner(spinner)

	installerStatus(o)

	spinner = internal.NewSpinner("Deleting namespace kyma-installer", "namespace kyma-installer deleted")
	deleteNamespace("kyma-installer")
	internal.StopSpinner(spinner)

	spinner = internal.NewSpinner("Deleting tiller", "tiller deleted")
	deleteTiller := []string{"delete", "-f", "https://raw.githubusercontent.com/kyma-project/kyma/master/installation/resources/tiller.yaml"}
	internal.RunKubeCmd(deleteTiller)
	internal.StopSpinner(spinner)

	spinner = internal.NewSpinner("Deleting ClusterRoleBinding for admin", "ClusterRoleBinding for admin deleted")
	deleteClusterRoleBindingCmd := []string{"delete", "clusterrolebinding", "cluster-admin-binding"}
	internal.RunKubeCmd(deleteClusterRoleBindingCmd)
	internal.StopSpinner(spinner)

	fmt.Printf("\nUninstalling kyma finished\n")

	return nil
}

func deleteNamespace(namespace string) {
	deleteNamespaceCmd := []string{"delete", "namespace", namespace}
	internal.RunKubeCmd(deleteNamespaceCmd)
}

func installerStatus(o *KymaOptions) error {
	currentDesc := ""
	var spinner chan struct{}

	for {
		installStatusCmd := []string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'"}
		installDescriptionCmd := []string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.description}'"}
		status := internal.RunKubeCmd(installStatusCmd)
		desc := internal.RunKubeCmd(installDescriptionCmd)

		switch status {
		case "Installed":
			if spinner != nil {
				internal.StopSpinner(spinner)
				spinner = nil
			}
			installVersionCmd := []string{"get", "installation/kyma-installation", "-o", "jsonpath='{.spec.version}'"}
			version := internal.RunKubeCmd(installVersionCmd)
			if version == "" {
				version = "N/A"
			}
			fmt.Printf("Kyma is installed using version %s!\n", version)
			clusterInfoCmd := []string{"cluster-info"}
			clusterInfo := internal.RunKubeCmd(clusterInfoCmd)
			fmt.Print(clusterInfo)
			break

		case "Error":
			fmt.Printf("Error installing Kyma: %s\n", desc)
			installLogsCmd := []string{"-n", "kyma-installer", "logs", "-l", "name=kyma-installer"}
			logs := internal.RunKubeCmd(installLogsCmd)
			fmt.Print(logs)

		case "InProgress":
			// only do something if the description has changed
			if desc != currentDesc {
				if spinner != nil {
					internal.StopSpinner(spinner)
					spinner = nil
					continue
				} else {
					spinner = internal.NewSpinner(desc, desc)
					currentDesc = desc
				}
			}

		default:
			fmt.Printf("Unexpected status: %s\n", status)
			os.Exit(1)
		}
		time.Sleep(sleep)
	}
}
