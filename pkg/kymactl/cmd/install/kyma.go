package install

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
	ReleaseVersion string
	ReleaseConfig  string
	ClusterAdmin   string
	NoWait         bool
}

//NewKymaOptions creates options with default values
func NewKymaOptions() *KymaOptions {
	return &KymaOptions{}
}

//NewKymaCmd creates a new kyma command
func NewKymaCmd(o *KymaOptions) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "kyma",
		Short: "Installs kyma to a running kubernetes cluster",
		Long: `Install kyma on a running kubernetes cluster.

Assure that your KUBECONFIG is pointing to the target cluster already.
The command will:
- Add your account as the cluster administrator 
- Install tiller
- Install the Kyma installer
- Configures the Kyma installer with the latest minimal configuration
- Triggers the installation
`,
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
		Aliases: []string{"i"},
	}

	cmd.Flags().StringVarP(&o.ReleaseVersion, "release", "r", "0.5.0", "kyma release to use")
	cmd.Flags().StringVar(&o.ReleaseConfig, "config", "c", "URL or path to the installer configuration yaml")
	cmd.Flags().StringVarP(&o.ClusterAdmin, "admin", "a", "default", "cluster admin user to configure")
	cmd.Flags().BoolVarP(&o.NoWait, "noWait", "n", false, "Do not wait for completion of kyma-installer")
	return cmd
}

//Run runs the command
func (o *KymaOptions) Run() error {
	fmt.Printf("Installing kyma in version '%s' with admin user '%s'\n\n‚", o.ReleaseVersion, o.ClusterAdmin)

	var spinner = internal.NewSpinner("Creating ClusterRoleBinding for admin '"+o.ClusterAdmin+"'", "ClusterRoleBinding created for admin "+o.ClusterAdmin)
	err := createlusterRoleBinding(o)
	if err != nil {
		return err
	}
	internal.StopSpinner(spinner)

	spinner = internal.NewSpinner("Installing tiller", "Tiller installed")
	err = installTiller(o)
	if err != nil {
		return err
	}
	internal.StopSpinner(spinner)

	spinner = internal.NewSpinner("Installing kyma-installer", "kyma-installer installed")
	err = installInstaller(o)
	if err != nil {
		return err
	}
	internal.StopSpinner(spinner)

	spinner = internal.NewSpinner("Requesting kyma-installer to install kyma", "kyma-installer is activated to install kyma")
	err = activateInstaller(o)
	if err != nil {
		return err
	}
	internal.StopSpinner(spinner)

	if !o.NoWait {
		err = waitForInstaller(o)
		if err != nil {
			return err
		}
	}

	err = printSummary(o)
	if err != nil {
		return err
	}

	return nil
}

func createlusterRoleBinding(o *KymaOptions) error {
	check, err := internal.IsClusterResourceDeployed("clusterrolebinding", "app", "kymactl")
	if err != nil {
		return err
	}
	if check {
		return nil
	}

	createClusterRoleBindingCmd := []string{"create", "clusterrolebinding", "cluster-admin-binding", "--clusterrole=cluster-admin", "--user=" + o.ClusterAdmin}
	_, err = internal.RunKubectlCmd(createClusterRoleBindingCmd)
	if err != nil {
		return err
	}

	labelClusterRoleBindingCmd := []string{"label", "clusterrolebinding", "cluster-admin-binding", "app=kymactl"}
	_, err = internal.RunKubectlCmd(labelClusterRoleBindingCmd)
	if err != nil {
		return err
	}

	return nil
}

func installTiller(o *KymaOptions) error {
	check, err := internal.IsPodDeployed("kube-system", "name", "tiller")
	if err != nil {
		return err
	}
	if !check {
		applyTiller := []string{"apply", "-f", "https://raw.githubusercontent.com/kyma-project/kyma/" + o.ReleaseVersion + "/installation/resources/tiller.yaml"}
		_, err = internal.RunKubectlCmd(applyTiller)
		if err != nil {
			return err
		}
	}
	err = internal.WaitForPod("kube-system", "name", "tiller")
	if err != nil {
		return err
	}
	return nil
}

func installInstaller(o *KymaOptions) error {
	check, err := internal.IsPodDeployed("kyma-installer", "name", "kyma-installer")
	if err != nil {
		return err
	}
	if !check {
		relaseURL := "https://github.com/kyma-project/kyma/releases/download/" + o.ReleaseVersion + "/kyma-config-local.yaml"
		if o.ReleaseConfig != "" {
			relaseURL = o.ReleaseConfig
		}
		applyInstaller := []string{"apply", "-f", relaseURL}
		_, err = internal.RunKubectlCmd(applyInstaller)
		if err != nil {
			return err
		}
		labelNamespaceCmd := []string{"label", "namespace", "kyma-installer", "app=kymactl"}
		_, err = internal.RunKubectlCmd(labelNamespaceCmd)
		if err != nil {
			return err
		}
	}
	err = internal.WaitForPod("kyma-installer", "name", "kyma-installer")
	if err != nil {
		return err
	}

	return nil
}

func activateInstaller(o *KymaOptions) error {
	installStatusCmd := []string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'"}
	status, err := internal.RunKubectlCmd(installStatusCmd)
	if err != nil {
		return err
	}
	if status == "InProgress" {
		return nil
	}

	labelInstaller := []string{"label", "installation/kyma-installation", "action=install"}
	_, err = internal.RunKubectlCmd(labelInstaller)
	if err != nil {
		return err
	}
	return nil
}

func printSummary(o *KymaOptions) error {
	installVersionCmd := []string{"get", "installation/kyma-installation", "-o", "jsonpath='{.spec.version}'"}
	version, err := internal.RunKubectlCmd(installVersionCmd)
	if err != nil {
		return err
	}
	if version == "" {
		version = "N/A"
	}
	fmt.Printf("Kyma is installed using version %s!\n", version)
	fmt.Println("Happy Kyma-ing!")
	fmt.Println("")
	clusterInfoCmd := []string{"cluster-info"}
	clusterInfo, err := internal.RunKubectlCmd(clusterInfoCmd)
	if err != nil {
		return err
	}
	fmt.Println(clusterInfo)
	return nil
}

func waitForInstaller(o *KymaOptions) error {
	currentDesc := ""
	var spinner chan struct{}
	installStatusCmd := []string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'"}
	installDescriptionCmd := []string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.description}'"}

	status, err := internal.RunKubectlCmd(installStatusCmd)
	if err != nil {
		return err
	}
	if status == "Installed" {
		return nil
	}

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
		case "Installed":
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
			fmt.Println(logs)

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
}
