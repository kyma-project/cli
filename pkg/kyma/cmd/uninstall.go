package cmd

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-incubator/kyma-cli/internal/kubectl"
	"github.com/kyma-incubator/kyma-cli/internal/step"

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
	s := o.NewStep(fmt.Sprintf("Checking requirements"))
	err := checkUninstallRequirements(o, s)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Requirements are fine")

	fmt.Printf("%s Uninstalling Kyma\n", step.InfoGliph)

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

	s = o.NewStep(fmt.Sprintf("Cleanup CRDs"))
	err = deleteLeftoverCRDs(o)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("CRDs cleaned")

	err = printUninstallSummary(o)
	if err != nil {
		return err
	}

	return nil
}

func checkUninstallRequirements(o *UninstallOptions, s step.Step) error {
	versionWarning, err := kubectl.CheckVersion(o.Verbose)
	if err != nil {
		s.Failure()
		return err
	}
	if versionWarning != "" {
		s.LogError(versionWarning)
	}
	return nil
}

func activateInstallerForUninstall(o *UninstallOptions) error {
	check, err := kubectl.IsPodDeployed("kyma-installer", "name", "kyma-installer", o.Verbose)
	if err != nil {
		return err
	}
	if !check {
		return nil
	}

	_, err = kubectl.RunCmd(o.Verbose, "label", "installation/kyma-installation", "action=uninstall")
	if err != nil {
		return err
	}

	return nil
}

func deleteInstaller(o *UninstallOptions) error {
	_, err := kubectl.RunCmd(o.Verbose, "delete", "CustomResourceDefinition", "installations.installer.kyma-project.io", "--timeout="+timeoutComplexDeletion, "--ignore-not-found=true")
	if err != nil {
		return err
	}

	_, err = kubectl.RunCmd(o.Verbose, "delete", "CustomResourceDefinition", "releases.release.kyma-project.io", "--timeout="+timeoutComplexDeletion, "--ignore-not-found=true")
	if err != nil {
		return err
	}

	_, err = kubectl.RunCmd(o.Verbose, "delete", "namespace", "kyma-installer", "--timeout="+timeoutComplexDeletion, "--ignore-not-found=true")
	if err != nil {
		return err
	}

	for {
		check, err := kubectl.IsClusterResourceDeployed("namespace", "app", "kyma-cli", o.Verbose)
		if err != nil {
			return err
		}
		if !check {
			break
		}
		time.Sleep(sleep)
	}

	_, err = kubectl.RunCmd(o.Verbose, "delete", "ClusterRoleBinding", "kyma-installer", "--timeout="+timeoutSimpleDeletion, "--ignore-not-found=true")
	if err != nil {
		return err
	}

	_, err = kubectl.RunCmd(o.Verbose, "delete", "ClusterRole", "kyma-installer-reader", "--timeout="+timeoutSimpleDeletion, "--ignore-not-found=true")
	if err != nil {
		return err
	}

	return nil
}

//cannot use the original yaml file as the version is not known or might be even custom
func deleteTiller(o *UninstallOptions) error {
	_, err := kubectl.RunCmd(o.Verbose, "-n", "kube-system", "delete", "all", "-l", "name=tiller")
	if err != nil {
		return err
	}

	err = kubectl.WaitForPodGone("kube-system", "name", "tiller", o.Verbose)
	if err != nil {
		return err
	}

	_, err = kubectl.RunCmd(o.Verbose, "delete", "ClusterRoleBinding", "tiller-cluster-admin", "--timeout="+timeoutSimpleDeletion, "--ignore-not-found=true")
	if err != nil {
		return nil
	}

	_, err = kubectl.RunCmd(o.Verbose, "-n", "kube-system", "delete", "ServiceAccount", "tiller", "--timeout="+timeoutSimpleDeletion, "--ignore-not-found=true")
	if err != nil {
		return err
	}

	return nil
}

func deleteClusterRoleBinding(o *UninstallOptions) error {
	_, err := kubectl.RunCmd(o.Verbose, "delete", "clusterrolebinding", "cluster-admin-binding", "--timeout="+timeoutSimpleDeletion, "--ignore-not-found=true")
	if err != nil {
		return err
	}
	return nil
}

func deleteLeftoverCRDs(o *UninstallOptions) error {
	crdNames, err := kubectl.RunCmd(o.Verbose, "get", "crd", "-o", "jsonpath='{.items[*].metadata.name}'")
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(crdNames))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}
		crd := scanner.Text()
		if strings.HasSuffix(crd, "kyma-project.io") || strings.HasSuffix(crd, "istio.io") || strings.HasSuffix(crd, "dex.coreos.com") {
			_, err := kubectl.RunCmd(o.Verbose, "delete", "crd", crd, "--timeout="+timeoutComplexDeletion, "--ignore-not-found=true")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func printUninstallSummary(o *UninstallOptions) error {
	fmt.Println()
	fmt.Println("Kyma uninstalled! :(")
	return nil
}

func waitForInstallerToUninstall(o *UninstallOptions) error {
	check, err := kubectl.IsPodDeployed("kyma-installer", "name", "kyma-installer", o.Verbose)
	if err != nil {
		return err
	}
	if !check {
		return nil
	}

	status, err := kubectl.RunCmd(o.Verbose, "get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'")
	if err != nil {
		return err
	}
	if status == "Uninstalled" {
		return nil
	}
	var s step.Step
	currentDesc := ""

	for {
		status, err := kubectl.RunCmd(o.Verbose, "get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'")
		if err != nil {
			return err
		}
		desc, err := kubectl.RunCmd(o.Verbose, "get", "installation/kyma-installation", "-o", "jsonpath='{.status.description}'")
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
			out, err := kubectl.RunCmd(o.Verbose, "-n", "kyma-installer", "logs", "-l", "name=kyma-installer")
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
