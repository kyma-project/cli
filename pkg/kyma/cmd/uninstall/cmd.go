package uninstall

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kyma-project/cli/pkg/kyma/core"
	"github.com/pkg/errors"

	"github.com/kyma-project/cli/internal/kubectl"
	"github.com/spf13/cobra"
)

var (
	namespacesToDelete = []string{"istio-system", "kyma-integration", "kyma-system", "natss", "knative-eventing", "knative-serving"}
	crdGroupsToDelete  = []string{"kyma-project.io", "istio.io", "dex.coreos.com", "knative.dev"}
)

const (
	sleep                  = 10 * time.Second
	timeoutSimpleDeletion  = "5s"
	timeoutComplexDeletion = "30s"
)

type command struct {
	opts *Options
	core.Command
}

//NewCmd creates a new kyma command
func NewCmd(o *Options) *cobra.Command {

	cmd := command{
		Command: core.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstalls Kyma from a running Kubernetes cluster",
		Long: `Uninstalls Kyma from a running Kubernetes cluster.

Make sure that your KUBECONFIG is already pointing to the target cluster.
This command:
- Removes your cluster administrator account
- Removes Tiller
- Removes Kyma Installer
`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"i"},
	}

	cobraCmd.Flags().DurationVarP(&o.Timeout, "timeout", "", 0, "Timeout after which Kyma CLI stops watching the installation progress")

	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	s := cmd.NewStep("Checking requirements")
	err := cmd.checkUninstallRequirements()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Requirements verified")

	s.LogInfof("Uninstalling Kyma")

	s = cmd.NewStep("Requesting kyma-installer to uninstall Kyma")
	err = cmd.activateInstallerForUninstall()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("kyma-installer is uninstalling Kyma")

	err = cmd.waitForInstallerToUninstall()
	if err != nil {
		return err
	}

	s = cmd.NewStep("Deleting kyma-installer")
	err = cmd.deleteInstaller()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("kyma-installer deleted")

	s = cmd.NewStep("Deleting tiller")
	err = cmd.deleteTiller()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Tiller deleted")

	s = cmd.NewStep("Deleting ClusterRoleBinding for admin")
	err = cmd.deleteClusterRoleBinding()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("ClusterRoleBinding for admin deleted")

	s = cmd.NewStep("Deleting Namespaces")
	// see https://github.com/kyma-project/kyma/issues/1826
	err = cmd.deleteLeftoverResources("namespace", namespacesToDelete)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Namespaces deleted")

	s = cmd.NewStep("Deleting CRDs")
	// see https://github.com/kyma-project/kyma/issues/1826
	err = cmd.deleteLeftoverResources("crd", crdGroupsToDelete)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("CRDs deleted")

	err = cmd.printUninstallSummary()
	if err != nil {
		return err
	}

	return nil
}

func (cmd *command) checkUninstallRequirements() error {
	versionWarning, err := kubectl.CheckVersion(cmd.Options.Verbose)
	if err != nil {
		cmd.CurrentStep.Failure()
		return err
	}
	if versionWarning != "" {
		cmd.CurrentStep.LogError(versionWarning)
	}
	return nil
}

func (cmd *command) activateInstallerForUninstall() error {
	check, err := cmd.Kubectl().IsPodDeployed("kyma-installer", "name", "kyma-installer")
	if err != nil {
		return err
	}
	if !check {
		return nil
	}

	_, err = cmd.Kubectl().RunCmd("label", "installation/kyma-installation", "action=uninstall")
	if err != nil {
		return err
	}

	return nil
}

func (cmd *command) deleteInstaller() error {
	_, err := cmd.Kubectl().RunCmd("delete", "CustomResourceDefinition", "installations.installer.kyma-project.io", "--timeout="+timeoutComplexDeletion, "--ignore-not-found=true")
	if err != nil {
		return err
	}

	_, err = cmd.Kubectl().RunCmd("delete", "CustomResourceDefinition", "releases.release.kyma-project.io", "--timeout="+timeoutComplexDeletion, "--ignore-not-found=true")
	if err != nil {
		return err
	}

	_, err = cmd.Kubectl().RunCmd("delete", "namespace", "kyma-installer", "--timeout="+timeoutComplexDeletion, "--ignore-not-found=true")
	if err != nil {
		return err
	}

	for {
		check, err := cmd.Kubectl().IsClusterResourceDeployed("namespace", "kyma-project.io/installation", "")
		if err != nil {
			return err
		}
		if !check {
			break
		}
		time.Sleep(sleep)
	}

	_, err = cmd.Kubectl().RunCmd("delete", "ClusterRoleBinding", "kyma-installer", "--timeout="+timeoutSimpleDeletion, "--ignore-not-found=true")
	if err != nil {
		return err
	}

	_, err = cmd.Kubectl().RunCmd("delete", "ClusterRole", "kyma-installer-reader", "--timeout="+timeoutSimpleDeletion, "--ignore-not-found=true")
	if err != nil {
		return err
	}

	return nil
}

//cannot use the original yaml file as the version is not known or might be even custom
func (cmd *command) deleteTiller() error {
	_, err := cmd.Kubectl().RunCmd("-n", "kube-system", "delete", "all", "-l", "name=tiller")
	if err != nil {
		return err
	}

	err = cmd.Kubectl().WaitForPodGone("kube-system", "name", "tiller")
	if err != nil {
		return err
	}

	_, err = cmd.Kubectl().RunCmd("delete", "ClusterRoleBinding", "tiller-cluster-admin", "--timeout="+timeoutSimpleDeletion, "--ignore-not-found=true")
	if err != nil {
		return nil
	}

	_, err = cmd.Kubectl().RunCmd("-n", "kube-system", "delete", "ServiceAccount", "tiller", "--timeout="+timeoutSimpleDeletion, "--ignore-not-found=true")
	if err != nil {
		return err
	}

	return nil
}

func (cmd *command) deleteClusterRoleBinding() error {
	_, err := cmd.Kubectl().RunCmd("delete", "clusterrolebinding", "cluster-admin-binding", "--timeout="+timeoutSimpleDeletion, "--ignore-not-found=true")
	if err != nil {
		return err
	}
	return nil
}

func (cmd *command) deleteLeftoverResources(resourceType string, resources []string) error {
	items, err := cmd.Kubectl().RunCmd("get", resourceType, "-o", "jsonpath='{.items[*].metadata.name}'")
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(items))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}
		item := scanner.Text()
		for _, v := range resources {
			if strings.HasSuffix(item, v) {
				cmd.CurrentStep.Status(item)
				_, err := cmd.Kubectl().RunCmd("delete", resourceType, item, "--timeout="+timeoutComplexDeletion)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (cmd *command) printUninstallSummary() error {
	fmt.Println()
	fmt.Println("Kyma uninstalled! :(")
	return nil
}

func (cmd *command) waitForInstallerToUninstall() error {
	currentDesc := ""
	_ = cmd.NewStep("Waiting for uninstallation to start")

	status, err := cmd.Kubectl().RunCmd("get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'")
	if err != nil {
		return err
	}
	if status == "Uninstalled" {
		return nil
	}

	var timeout <-chan time.Time
	var errorOccured bool
	if cmd.opts.Timeout > 0 {
		timeout = time.After(cmd.opts.Timeout)
	}

	for {
		select {
		case <-timeout:
			cmd.CurrentStep.Failure()
			_ = cmd.printUninstallationErrorLog()
			return errors.New("Timeout reached while waiting for Kyma to uninstall")
		default:
			status, desc, err := cmd.getUninstallationStatus()
			if err != nil {
				return err
			}

			switch status {
			case "Uninstalled":
				cmd.CurrentStep.Success()
				return nil

			case "Error":
				if !errorOccured {
					errorOccured = true
					cmd.CurrentStep.Failuref("Failed to uninstall Kyma: %s", desc)
					cmd.CurrentStep.LogInfof("To fetch the logs from the Installer, run: 'kubectl logs -n kyma-installer -l name=kyma-installer'")
				}

			case "InProgress":
				errorOccured = false
				// only do something if the description has changed
				if desc != currentDesc {
					cmd.CurrentStep.Success()
					cmd.CurrentStep = cmd.opts.NewStep(fmt.Sprintf(desc))
					currentDesc = desc
				}

			default:
				cmd.CurrentStep.Failure()
				fmt.Printf("Unexpected status: %s\n", status)
				os.Exit(1)
			}
			time.Sleep(sleep)
		}
	}
}

func (cmd *command) getUninstallationStatus() (status string, desc string, err error) {
	status, err = cmd.Kubectl().RunCmd("get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'")
	if err != nil {
		return
	}
	desc, err = cmd.Kubectl().RunCmd("get", "installation/kyma-installation", "-o", "jsonpath='{.status.description}'")
	return
}

func (cmd *command) printUninstallationErrorLog() error {
	logs, err := cmd.Kubectl().RunCmd("get", "installation", "kyma-installation", "-o", "go-template", `--template={{- range .status.errorLog -}}
{{.component}}:
{{.log}} [{{.occurrences}}]

{{- end}}
`)
	if err != nil {
		return err
	}
	fmt.Println(logs)
	return nil
}
