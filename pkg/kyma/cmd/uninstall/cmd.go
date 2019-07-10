package uninstall

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/kyma/core"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

var (
	namespacesToDelete = []string{"istio-system", "kyma-integration", "kyma-system", "natss", "knative-eventing", "knative-serving"}
	crdGroupsToDelete  = []string{"kyma-project.io", "istio.io", "dex.coreos.com", "knative.dev"}
)

const (
	sleep           = 10 * time.Second
	deletionTimeout = "120s"
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

	cobraCmd.Flags().DurationVarP(&o.Timeout, "timeout", "", 30*time.Minute, "Timeout after which Kyma CLI stops watching the uninstallation progress")

	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	var err error
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. PLease make sure that you have a valid kubeconfig.")
	}

	s := cmd.NewStep("Uninstalling Kyma")
	if err := cmd.activateInstallerForUninstall(); err != nil {
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

	s = cmd.NewStep("Deleting CRDs")
	// see https://github.com/kyma-project/kyma/issues/1826
	err = cmd.deleteLeftoverResources("crd", crdGroupsToDelete)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("CRDs deleted")

	s = cmd.NewStep("Deleting Namespaces")
	// see https://github.com/kyma-project/kyma/issues/1826
	err = cmd.deleteLeftoverResources("namespace", namespacesToDelete)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Namespaces deleted")

	err = cmd.printUninstallSummary()
	if err != nil {
		return err
	}

	return nil
}

func (cmd *command) activateInstallerForUninstall() error {
	deployed, err := cmd.K8s.IsPodDeployedByLabel("kyma-installer", "name", "kyma-installer")
	if err != nil {
		return err
	}

	if !deployed {
		return nil
	}

	_, err = cmd.Kubectl().RunCmd("label", "installation/kyma-installation", "action=uninstall")
	return err
}

func (cmd *command) deleteInstaller() error {
	_, err := cmd.Kubectl().RunCmd("delete", "CustomResourceDefinition", "installations.installer.kyma-project.io", "--timeout="+deletionTimeout, "--ignore-not-found=true")
	if err != nil {
		return err
	}

	_, err = cmd.Kubectl().RunCmd("delete", "CustomResourceDefinition", "releases.release.kyma-project.io", "--timeout="+deletionTimeout, "--ignore-not-found=true")
	if err != nil {
		return err
	}

	err = cmd.K8s.Static().CoreV1().Namespaces().Delete("kyma-installer", &metav1.DeleteOptions{})
	if err != nil && !strings.Contains(err.Error(), "not found") { // treat not found as positive deletion
		return err
	}

	for {
		ns, err := cmd.K8s.Static().CoreV1().Namespaces().List(metav1.ListOptions{LabelSelector: "kyma-project.io/installation="})
		if err != nil {
			return err
		}
		if len(ns.Items) == 0 {
			break
		}
		time.Sleep(sleep)
	}

	err = cmd.K8s.Static().RbacV1().ClusterRoleBindings().Delete("kyma-installer", &metav1.DeleteOptions{})
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return err
	}

	err = cmd.K8s.Static().RbacV1().ClusterRoles().Delete("kyma-installer-reader", &metav1.DeleteOptions{})
	if err != nil && !strings.Contains(err.Error(), "not found") { // treat not found as positive deletion
		return err
	}

	return nil
}

//cannot use the original yaml file as the version is not known or might be even custom
func (cmd *command) deleteTiller() error {
	_, err := cmd.Kubectl().RunCmd("-n", "kube-system", "delete", "deployment,ServiceAccount,ClusterRoleBinding,Service,Job", "-l", "kyma-project.io/installation=")
	if err != nil {
		return err
	}

	err = cmd.K8s.Static().RbacV1().RoleBindings("kube-system").Delete("tiller-certs", &metav1.DeleteOptions{})
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return err
	}

	err = cmd.K8s.Static().RbacV1().Roles("kube-system").Delete("tiller-certs-installer", &metav1.DeleteOptions{})
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return err
	}

	err = cmd.K8s.Static().CoreV1().ServiceAccounts("kube-system").Delete("tiller-certs-sa", &metav1.DeleteOptions{})
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return err
	}

	err = cmd.K8s.WaitPodsGone("kube-system", "name", "tiller")
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
				_, err := cmd.Kubectl().RunCmd("delete", resourceType, item, "--timeout="+deletionTimeout)
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
					cmd.CurrentStep.LogInfof("There was an error uninstalling Kyma, which might be OK. Will retry later...\n%s", desc)
					cmd.CurrentStep.LogInfof("For more information, run: 'kubectl logs -n kyma-installer -l name=kyma-installer'")
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
