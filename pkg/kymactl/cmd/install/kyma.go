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

type KymaOptions struct {
	Release      string
	ClusterAdmin string
}

func NewKymaOptions() *KymaOptions {
	return &KymaOptions{}
}

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

	cmd.Flags().StringVarP(&o.Release, "release", "r", "0.5.0", "kyma release to use")
	cmd.Flags().StringVarP(&o.ClusterAdmin, "admin", "a", "default", "cluster admin user to configure")
	return cmd
}

func (o *KymaOptions) Run() error {
	fmt.Printf("Installing kyma in version '%s' with admin user '%s'\n\n‚", o.Release, o.ClusterAdmin)

	var spinner = internal.NewSpinner("Creating ClusterRoleBinding for admin"+o.ClusterAdmin, "ClusterRoleBinding created for admin"+o.ClusterAdmin)
	createClusterRoleBindingCmd := []string{"create", "clusterrolebinding", "cluster-admin-binding", "--clusterrole=cluster-admin", "--user=" + o.ClusterAdmin}
	internal.RunKubeCmd(createClusterRoleBindingCmd)
	internal.StopSpinner(spinner)

	spinner = internal.NewSpinner("Installing tiller", "Tiller installed")
	applyTiller := []string{"apply", "-f", "https://raw.githubusercontent.com/kyma-project/kyma/master/installation/resources/tiller.yaml"}
	internal.RunKubeCmd(applyTiller)
	internal.IsPodReady("kube-system", "name", "tiller")
	internal.StopSpinner(spinner)

	spinner = internal.NewSpinner("Installing kyma-installer", "kyma-installer installed")
	applyInstaller := []string{"apply", "-f", "https://github.com/kyma-project/kyma/releases/download/" + o.Release + "/kyma-config-local.yaml"}
	internal.RunKubeCmd(applyInstaller)
	internal.IsPodReady("kyma-installer", "name", "kyma-installer")
	internal.StopSpinner(spinner)

	spinner = internal.NewSpinner("Requesting kyma-installer to install kyma version "+o.Release, "kyma-installer is installing kyma in version "+o.Release)
	labelInstaller := []string{"label", "installation/kyma-installation", "action=install"}
	internal.RunKubeCmd(labelInstaller)
	internal.StopSpinner(spinner)

	installerStatus(o)

	fmt.Printf("\nInstallation of kyma finished in version '%s' with admin user '%s'\n", o.Release, o.ClusterAdmin)

	return nil
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
