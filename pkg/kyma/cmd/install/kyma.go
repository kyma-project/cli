package install

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kyma-incubator/kymactl/internal/step"

	"github.com/kyma-incubator/kymactl/internal"
	"github.com/kyma-incubator/kymactl/pkg/kyma/core"
	"github.com/spf13/cobra"
)

const (
	sleep = 5 * time.Second
)

//KymaOptions defines available options for the command
type KymaOptions struct {
	*core.Options
	ReleaseVersion string
	ReleaseConfig  string
	NoWait         bool
	Domain         string
	Local          bool
	LocalPath      string
}

//NewKymaOptions creates options with default values
func NewKymaOptions(o *core.Options) *KymaOptions {
	return &KymaOptions{Options: o}
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

	goPath := os.Getenv("GOPATH")
	if goPath != "" {
		defaultLocalPath := filepath.Join(goPath, "src", "github.com", "kyma-project", "kyma")
		cmd.Flags().BoolVarP(&o.Local, "local", "l", false, "Install from sources")
		cmd.Flags().StringVarP(&o.LocalPath, "local-path", "", defaultLocalPath, "Path to local sources to use")
	}

	return cmd
}

//Run runs the command
func (o *KymaOptions) Run() error {
	if o.Local {
		fmt.Printf("Local installation from: %s\n", o.LocalPath)
	} else {
		fmt.Printf("Installing kyma in version '%s'\n", o.ReleaseVersion)
	}
	fmt.Println()

	s := o.NewStep(fmt.Sprintf("Checking requirements"))
	err := internal.CheckKubectlVersion()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Requirements are fine")

	s = o.NewStep(fmt.Sprintf("Installing tiller"))
	err = installTiller(o)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Tiller installed")

	s = o.NewStep(fmt.Sprintf("Installing kyma-installer"))
	err = installInstaller(o)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("kyma-installer installed")

	s = o.NewStep(fmt.Sprintf("Requesting kyma-installer to install kyma"))
	err = activateInstaller(o)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("kyma-installer is installing kyma")

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
		if o.Local {
			err = installInstallerFromLocalSources(o)
		} else {
			err = installInstallerFromRelease(o)
		}
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

func installInstallerFromRelease(o *KymaOptions) error {
	relaseURL := "https://github.com/kyma-project/kyma/releases/download/" + o.ReleaseVersion + "/kyma-config-local.yaml"
	if o.ReleaseConfig != "" {
		relaseURL = o.ReleaseConfig
	}
	_, err := internal.RunKubectlCmd([]string{"apply", "-f", relaseURL})
	if err != nil {
		return err
	}
	return labelInstallerNamespace()
}

func installInstallerFromLocalSources(o *KymaOptions) error {
	err := buildKymaInstaller(o)
	if err != nil {
		return err
	}

	err = applyKymaInstaller(o)

	return labelInstallerNamespace()
}

func buildKymaInstaller(o *KymaOptions) error {
	dc, err := internal.MinikubeDockerClient()
	if err != nil {
		return err
	}

	imageNameCmd := exec.Command(filepath.Join(o.LocalPath, "installation", "scripts", "extract-kyma-installer-image.sh"))
	imageName, err := imageNameCmd.Output()
	if err != nil {
		return err
	}

	return dc.BuildImage(docker.BuildImageOptions{
		Name:         strings.TrimSpace(string(imageName)),
		Dockerfile:   filepath.Join("tools", "kyma-installer", "kyma.Dockerfile"),
		OutputStream: ioutil.Discard,
		ContextDir:   filepath.Join(o.LocalPath),
	})
}

func applyKymaInstaller(o *KymaOptions) error {
	yamls := make([]map[string]interface{}, 0)
	installerYamlPath := filepath.Join(o.LocalPath, "installation", "resources", "installer-local.yaml")
	installerYamlFile, err := os.Open(installerYamlPath)
	if err != nil {
		return err
	}
	defer func() { _ = installerYamlFile.Close() }()
	dec := yaml.NewDecoder(installerYamlFile)
	for {
		m := make(map[string]interface{})
		err := dec.Decode(m)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		yamls = append(yamls, m)
	}

	installerConfigYamlPath := filepath.Join(o.LocalPath, "installation", "resources", "installer-config-local.yaml.tpl")
	installerConfigYamlFile, err := os.Open(installerConfigYamlPath)
	if err != nil {
		return err
	}
	defer func() { _ = installerConfigYamlFile.Close() }()
	dec = yaml.NewDecoder(installerConfigYamlFile)
	for {
		m := make(map[string]interface{})
		err := dec.Decode(m)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		yamls = append(yamls, m)
	}

	installerCRYamlPath := filepath.Join(o.LocalPath, "installation", "resources", "installer-cr.yaml.tpl")
	installerCRYamlFile, err := os.Open(installerCRYamlPath)
	if err != nil {
		return err
	}
	defer func() { _ = installerCRYamlFile.Close() }()
	dec = yaml.NewDecoder(installerCRYamlFile)
	m := make(map[string]interface{})
	err = dec.Decode(m)
	if err != nil {
		return err
	}
	yamls = append(yamls, m)
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer func() { _ = stdinPipe.Close() }()
	buf := &bytes.Buffer{}
	enc := yaml.NewEncoder(buf)
	for _, y := range yamls {
		err = enc.Encode(y)
		if err != nil {
			return err
		}
	}
	err = enc.Close()
	if err != nil {
		return err
	}
	cmd.Stdin = buf
	return cmd.Run()
}

func labelInstallerNamespace() error {
	_, err := internal.RunKubectlCmd([]string{"label", "namespace", "kyma-installer", "app=kymactl"})
	return err
}

func activateInstaller(_ *KymaOptions) error {
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

func waitForInstaller(_ *KymaOptions) error {
	currentDesc := ""
	var s step.Step
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
			if s != nil {
				s.Success()
			}
			return nil

		case "Error":
			if s != nil {
				s.Failure()
			}
			fmt.Printf("Error installing Kyma: %s\n", desc)
			logs, err := internal.RunKubectlCmd([]string{"-n", "kyma-installer", "logs", "-l", "name=kyma-installer"})
			if err != nil {
				return err
			}
			fmt.Println(logs)

		case "InProgress":
			// only do something if the description has changed
			if desc != currentDesc {
				if s != nil {
					s.Success()
				} else {
					s = o.NewStep(fmt.Sprintf(desc))
					currentDesc = desc
				}
			}

		default:
			if s != nil {
				s.Failure()
			}
			fmt.Printf("Unexpected status: %s\n", status)
			os.Exit(1)
		}
		time.Sleep(sleep)
	}
}
