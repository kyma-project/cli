package install

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/kyma-incubator/kymactl/internal"
	"github.com/spf13/cobra"
)

const (
	sleep = 5 * time.Second
)

//KymaOptions defines available options for the command
type KymaOptions struct {
	ReleaseVersion string
	ReleaseConfig  string
	NoWait         bool
	Domain         string
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
- Install tiller
- Install the Kyma installer
- Configures the Kyma installer with the latest minimal configuration
- Triggers the installation
`,
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
		Aliases: []string{"i"},
	}

	cmd.Flags().StringVarP(&o.ReleaseVersion, "release", "r", "0.6.1", "kyma release to use")
	cmd.Flags().StringVarP(&o.ReleaseConfig, "config", "c", "", "URL or path to the installer configuration yaml")
	cmd.Flags().BoolVarP(&o.NoWait, "noWait", "n", false, "Do not wait for completion of kyma-installer")
	cmd.Flags().StringVarP(&o.Domain, "domain", "d", "kyma.local", "domain to use for installation")
	return cmd
}

//Run runs the command
func (o *KymaOptions) Run() error {
	fmt.Printf("Installing kyma in version '%s'\n", o.ReleaseVersion)
	fmt.Println()

	spinner := internal.NewSpinner("Checking requirements", "Requirements are fine")
	err := internal.CheckKubectlVersion()
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

	spinner = internal.NewSpinner("Requesting kyma-installer to install kyma", "kyma-installer is installing kyma")
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

func installTiller(o *KymaOptions) error {
	check, err := internal.IsPodDeployed("kube-system", "name", "tiller")
	if err != nil {
		return err
	}
	if !check {
		_, err = internal.RunKubectlCmd([]string{"apply", "-f", "https://raw.githubusercontent.com/kyma-project/kyma/" + o.ReleaseVersion + "/installation/resources/tiller.yaml"})
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
		_, err = internal.RunKubectlCmd([]string{"apply", "-f", relaseURL})
		if err != nil {
			return err
		}
		_, err = internal.RunKubectlCmd([]string{"label", "namespace", "kyma-installer", "app=kymactl"})
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
	status, err := internal.RunKubectlCmd([]string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'"})
	if err != nil {
		return err
	}
	if status == "InProgress" {
		return nil
	}

	_, err = internal.RunKubectlCmd([]string{"label", "installation/kyma-installation", "action=install"})
	if err != nil {
		return err
	}
	return nil
}

func printSummary(o *KymaOptions) error {
	version, err := internal.GetKymaVersion()
	if err != nil {
		return err
	}

	pwdEncoded, err := internal.RunKubectlCmd([]string{"-n", "kyma-system", "get", "secret", "admin-user", "-o", "jsonpath='{.data.password}'"})
	if err != nil {
		return err
	}

	pwdDecoded, err := base64.StdEncoding.DecodeString(pwdEncoded)
	if err != nil {
		return err
	}

	emailEncoded, err := internal.RunKubectlCmd([]string{"-n", "kyma-system", "get", "secret", "admin-user", "-o", "jsonpath='{.data.email}'"})
	if err != nil {
		return err
	}

	emailDecoded, err := base64.StdEncoding.DecodeString(emailEncoded)
	if err != nil {
		return err
	}

	clusterInfo, err := internal.RunKubectlCmd([]string{"cluster-info"})
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(clusterInfo)
	fmt.Println()
	fmt.Printf("Kyma is installed in version %s\n", version)
	fmt.Printf("Kyma console:     https://console.%s\n", o.Domain)
	fmt.Printf("Kyma admin email: %s\n", emailDecoded)
	fmt.Printf("Kyma admin pwd:   %s\n", pwdDecoded)
	fmt.Println()
	fmt.Println("Happy Kyma-ing! :)")
	fmt.Println()

	return nil
}

func waitForInstaller(o *KymaOptions) error {
	currentDesc := ""
	var spinner chan struct{}
	installStatusCmd := []string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'"}

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
		desc, err := internal.RunKubectlCmd([]string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.description}'"})
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
			logs, err := internal.RunKubectlCmd([]string{"-n", "kyma-installer", "logs", "-l", "name=kyma-installer"})
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
