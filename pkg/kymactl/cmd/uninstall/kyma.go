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
		Use:   "kyma",
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

	var spinner = internal.NewSpinner("Activate kyma-installer to uninstall kyma", "kyma-installer activated to uninstall kyma")
	activateInstaller(o)
	internal.StopSpinner(spinner)

	waitForInstaller(o)

	spinner = internal.NewSpinner("Deleting namespace kyma-installer", "namespace kyma-installer deleted")
	deleteNamespace("kyma-installer")
	internal.StopSpinner(spinner)

	spinner = internal.NewSpinner("Deleting tiller", "tiller deleted")
	deleteTiller()
	internal.StopSpinner(spinner)

	spinner = internal.NewSpinner("Deleting ClusterRoleBinding for admin", "ClusterRoleBinding for admin deleted")
	deleteClusterRoleBinding()
	internal.StopSpinner(spinner)

	fmt.Printf("\nkyma uninstalled\n")

	return nil
}

func activateInstaller(o *KymaOptions) error {
	if !internal.IsPodDeployed("kyma-installer", "name", "kyma-installer") {
		return nil
	}

	labelInstaller := []string{"label", "installation/kyma-installation", "action=uninstall"}
	internal.RunKubeCmd(labelInstaller)

	return nil
}

func deleteNamespace(namespace string) error {
	if !internal.IsClusterResourceDeployed("namespace", "app", "kymactl") {
		return nil
	}

	deleteNamespaceCmd := []string{"delete", "namespace", namespace}
	internal.RunKubeCmd(deleteNamespaceCmd)

	for {
		if !internal.IsClusterResourceDeployed("namespace", "app", "kymactl") {
			break
		}
		time.Sleep(sleep)
	}
	return nil
}

func deleteTiller() error {
	if !internal.IsPodDeployed("kube-system", "name", "tiller") {
		return nil
	}

	deleteTiller := []string{"delete", "-f", "https://raw.githubusercontent.com/kyma-project/kyma/master/installation/resources/tiller.yaml"}
	internal.RunKubeCmd(deleteTiller)
	return nil
}

func deleteClusterRoleBinding() error {
	if !internal.IsClusterResourceDeployed("clusterrolebinding", "app", "kymactl") {
		return nil
	}

	deleteClusterRoleBindingCmd := []string{"delete", "clusterrolebinding", "cluster-admin-binding"}
	internal.RunKubeCmd(deleteClusterRoleBindingCmd)

	return nil
}

func waitForInstaller(o *KymaOptions) error {
	if !internal.IsPodDeployed("kyma-installer", "name", "kyma-installer") {
		return nil
	}

	installStatusCmd := []string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'"}
	installDescriptionCmd := []string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.description}'"}

	status := internal.RunKubeCmd(installStatusCmd)
	if status == "Uninstalled" {
		return nil
	}

	currentDesc := ""
	var spinner chan struct{}

	for {
		status := internal.RunKubeCmd(installStatusCmd)
		desc := internal.RunKubeCmd(installDescriptionCmd)

		switch status {
		case "Uninstalled":
			if spinner != nil {
				internal.StopSpinner(spinner)
				spinner = nil
			}
			return nil

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
	return nil
}
