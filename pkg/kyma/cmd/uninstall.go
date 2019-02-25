package cmd

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/kyma-cli/internal/step"

	"github.com/kyma-incubator/kyma-cli/internal"
	"github.com/kyma-incubator/kyma-cli/pkg/kyma/core"
	"github.com/spf13/cobra"
)

//UninstallOptions defines available options for the command
type UninstallOptions struct {
	*core.Options
}

//NewUninstallOptions creates options with default values
func NewUninstallOptions(o *core.Options) *UninstallOptions {
	return &UninstallOptions{Options: o}
}

//NewUninstallCmd creates a new kyma command
func NewUninstallCmd(o *UninstallOptions) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstalls Kyma from a running kubernetes cluster",
		Long: `Uninstall Kyma on a running kubernetes cluster.

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
func (o *UninstallOptions) Run() error {
	fmt.Printf("Uninstalling Kyma\n")
	fmt.Println()

	s := o.NewStep(fmt.Sprintf("Checking requirements"))
	s.Start()
	err := internal.CheckKubectlVersion()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Requirements are fine")

	s = o.NewStep(fmt.Sprintf("Activate kyma-installer to uninstall kyma"))
	err = activateInstallerForUninstall(o)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("kyma-installer activated to uninstall kyma")

	err = waitForInstallerToUninstall(o)
	if err != nil {
		return err
	}

	/*s = o.NewStep(fmt.Sprintf("Deleting kyma-integration namespace as it is not getting cleaned properly"))
	err = deleteKymaIntegration(o)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("kyma-integration namespace deleted")*/

	s = o.NewStep(fmt.Sprintf("Deleting kyma-installer"))
	err = deleteInstaller(o)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("kyma-installer deleted")

	s = o.NewStep(fmt.Sprintf("Deleting tiller"))
	err = deleteTiller(o)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("tiller deleted")

	s = o.NewStep(fmt.Sprintf("Deleting ClusterRoleBinding for admin"))
	err = deleteClusterRoleBinding(o)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("ClusterRoleBinding for admin deleted")

	err = printUninstallSummary(o)
	if err != nil {
		return err
	}

	return nil
}

func activateInstallerForUninstall(o *UninstallOptions) error {
	check, err := internal.IsPodDeployed("kyma-installer", "name", "kyma-installer")
	if err != nil {
		return err
	}
	if !check {
		return nil
	}

	_, err = internal.RunKubectlCmd([]string{"label", "installation/kyma-installation", "action=uninstall"})
	if err != nil {
		return err
	}

	return nil
}

func deleteInstaller(o *UninstallOptions) error {
	check, err := internal.IsClusterResourceDeployed("namespace", "app", "kyma-cli")
	if err != nil {
		return err
	}
	if !check {
		return nil
	}

	_, err = internal.RunKubectlCmd([]string{"delete", "namespace", "kyma-installer"})
	if err != nil {
		return err
	}

	for {
		check, err := internal.IsClusterResourceDeployed("namespace", "app", "kyma-cli")
		if err != nil {
			return err
		}
		if !check {
			break
		}
		time.Sleep(sleep)
	}

	_, err = internal.RunKubectlCmd([]string{"delete", "ClusterRoleBinding", "kyma-installer"})
	if err != nil {
		return err
	}

	_, err = internal.RunKubectlCmd([]string{"delete", "ClusterRole", "kyma-installer-reader"})
	if err != nil {
		return err
	}

	_, err = internal.RunKubectlCmd([]string{"delete", "CustomResourceDefinition", "installations.installer.kyma-project.io"})
	if err != nil {
		return err
	}

	_, err = internal.RunKubectlCmd([]string{"delete", "CustomResourceDefinition", "releases.release.kyma-project.io"})
	if err != nil {
		return err
	}

	return nil
}

//cannot use the original yaml file as the version is not known or might be even custom
func deleteTiller(o *UninstallOptions) error {
	check, err := internal.IsPodDeployed("kube-system", "name", "tiller")
	if err != nil {
		return err
	}
	if check {
		_, err = internal.RunKubectlCmd([]string{"-n", "kube-system", "delete", "all", "-l", "name=tiller"})
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
	}

	_, err = internal.RunKubectlCmd([]string{"delete", "ClusterRoleBinding", "tiller-cluster-admin"})
	if err != nil {
		return nil
	}

	_, err = internal.RunKubectlCmd([]string{"-n", "kube-system", "delete", "ServiceAccount", "tiller"})
	if err != nil {
		return err
	}

	return nil
}

func deleteClusterRoleBinding(o *UninstallOptions) error {
	check, err := internal.IsClusterResourceDeployed("clusterrolebinding", "app", "kyma-cli")
	if err != nil {
		return err
	}
	if !check {
		return nil
	}

	_, err = internal.RunKubectlCmd([]string{"delete", "clusterrolebinding", "cluster-admin-binding"})
	if err != nil {
		return err
	}
	return nil
}

func printUninstallSummary(o *UninstallOptions) error {
	fmt.Println()
	fmt.Println("Kyma uninstalled! :(")
	return nil
}

func waitForInstallerToUninstall(o *UninstallOptions) error {
	check, err := internal.IsPodDeployed("kyma-installer", "name", "kyma-installer")
	if err != nil {
		return err
	}
	if !check {
		return nil
	}

	cmd := []string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'"}
	status, err := internal.RunKubectlCmd(cmd)
	if err != nil {
		return err
	}
	if status == "Uninstalled" {
		return nil
	}
	var s step.Step
	currentDesc := ""

	for {
		status, err := internal.RunKubectlCmd(cmd)
		if err != nil {
			return err
		}
		desc, err := internal.RunKubectlCmd([]string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.description}'"})
		if err != nil {
			return err
		}

		switch status {
		case "Uninstalled":
			if s != nil {
				s.Success()
			}
			return nil

		case "Error":
			if s != nil {
				s.Failure()
			}
			fmt.Printf("Error installing Kyma: %s\n", desc)
			out, err := internal.RunKubectlCmd([]string{"-n", "kyma-installer", "logs", "-l", "name=kyma-installer"})
			if err != nil {
				return err
			}
			fmt.Print(out)

		case "InProgress":
			// only do something if the description has changed
			if desc != currentDesc {
				if s != nil {
					s.Success()
				} else {
					s = o.NewStep(fmt.Sprintf(desc))
					s.Start()
					currentDesc = desc
				}
			}

		default:
			if s != nil {
				s.Failure()
			}
			return fmt.Errorf("Unexpected status: %s", status)
		}
		time.Sleep(sleep)
	}
}
