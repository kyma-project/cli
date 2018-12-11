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

//KymaOptions defines available options for the command
type KymaOptions struct {
}

//NewKymaOptions creates options with default values
func NewKymaOptions() *KymaOptions {
	return &KymaOptions{}
}

//NewKymaCmd creates a new kyma command
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

//Run runs the command
func (o *KymaOptions) Run() error {
	fmt.Printf("Uninstalling kyma\n\n")

	var spinner = internal.NewSpinner("Activate kyma-installer to uninstall kyma", "kyma-installer activated to uninstall kyma")
	err := activateInstaller(o)
	if err != nil {
		return err
	}
	internal.StopSpinner(spinner)

	err = waitForInstaller(o)
	if err != nil {
		return err
	}

	spinner = internal.NewSpinner("Deleting namespace kyma-installer", "namespace kyma-installer deleted")
	err = deleteNamespace("kyma-installer")
	if err != nil {
		return err
	}
	internal.StopSpinner(spinner)

	spinner = internal.NewSpinner("Deleting tiller", "tiller deleted")
	err = deleteTiller()
	if err != nil {
		return err
	}
	internal.StopSpinner(spinner)

	spinner = internal.NewSpinner("Deleting ClusterRoleBinding for admin", "ClusterRoleBinding for admin deleted")
	err = deleteClusterRoleBinding()
	if err != nil {
		return err
	}
	internal.StopSpinner(spinner)

	fmt.Printf("\nkyma uninstalled\n")

	return nil
}

func activateInstaller(o *KymaOptions) error {
	check, err := internal.IsPodDeployed("kyma-installer", "name", "kyma-installer")
	if err != nil {
		return err
	}
	if !check {
		return nil
	}

	labelInstaller := []string{"label", "installation/kyma-installation", "action=uninstall"}
	_, err = internal.RunKubectlCmd(labelInstaller)
	if err != nil {
		return err
	}

	return nil
}

func deleteNamespace(namespace string) error {
	check, err := internal.IsClusterResourceDeployed("namespace", "app", "kymactl")
	if err != nil {
		return err
	}
	if !check {
		return nil
	}

	deleteNamespaceCmd := []string{"delete", "namespace", namespace}
	_, err = internal.RunKubectlCmd(deleteNamespaceCmd)
	if err != nil {
		return err
	}

	for {
		check, err := internal.IsClusterResourceDeployed("namespace", "app", "kymactl")
		if err != nil {
			return err
		}
		if !check {
			break
		}
		time.Sleep(sleep)
	}
	return nil
}

func deleteTiller() error {
	check, err := internal.IsPodDeployed("kube-system", "name", "tiller")
	if err != nil {
		return err
	}
	if !check {
		return nil
	}

	deleteTiller := []string{"delete", "-f", "https://raw.githubusercontent.com/kyma-project/kyma/master/installation/resources/tiller.yaml"}
	_, err = internal.RunKubectlCmd(deleteTiller)
	if err != nil {
		return err
	}
	for {
		check, err := internal.IsPodDeployed("kube-system", "name", "tiller")
		if err != nil {
			return err
		}
		if !check {
			break
		}
		time.Sleep(sleep)
	}
	return nil
}

func deleteClusterRoleBinding() error {
	check, err := internal.IsClusterResourceDeployed("clusterrolebinding", "app", "kymactl")
	if err != nil {
		return err
	}
	if !check {
		return nil
	}

	deleteClusterRoleBindingCmd := []string{"delete", "clusterrolebinding", "cluster-admin-binding"}
	_, err = internal.RunKubectlCmd(deleteClusterRoleBindingCmd)
	if err != nil {
		return err
	}
	return nil
}

func waitForInstaller(o *KymaOptions) error {
	check, err := internal.IsPodDeployed("kyma-installer", "name", "kyma-installer")
	if err != nil {
		return err
	}
	if !check {
		return nil
	}

	installStatusCmd := []string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'"}
	installDescriptionCmd := []string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.description}'"}

	status, err := internal.RunKubectlCmd(installStatusCmd)
	if err != nil {
		return err
	}
	if status == "Uninstalled" {
		return nil
	}

	currentDesc := ""
	var spinner chan struct{}

	for {
		status, err := internal.RunKubectlCmd(installStatusCmd)
		if err != nil {
			return err
		}
		desc, err := internal.RunKubectlCmd(installDescriptionCmd)
		if err != nil {
			return err
		}

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
			logs, err := internal.RunKubectlCmd(installLogsCmd)
			if err != nil {
				return err
			}
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
