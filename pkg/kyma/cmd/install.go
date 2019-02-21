package cmd

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/kyma-incubator/kyma-cli/internal/kubectl"
	"github.com/kyma-incubator/kyma-cli/internal/minikube"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"

	"github.com/kyma-incubator/kyma-cli/internal/step"

	"github.com/kyma-incubator/kyma-cli/internal"
	"github.com/kyma-incubator/kyma-cli/pkg/kyma/core"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
)

//InstallOptions defines available options for the command
type InstallOptions struct {
	*core.Options
	ReleaseVersion        string
	ReleaseConfig         string
	NoWait                bool
	Domain                string
	Local                 bool
	LocalSrcPath          string
	LocalInstallerVersion string
	LocalInstallerDir     string
}

//NewInstallOptions creates options with default values
func NewInstallOptions(o *core.Options) *InstallOptions {
	return &InstallOptions{Options: o}
}

//NewInstallCmd creates a new kyma command
func NewInstallCmd(o *InstallOptions) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Installs Kyma to a running kubernetes cluster",
		Long: `Install Kyma on a running kubernetes cluster.

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

	cmd.Flags().StringVarP(&o.ReleaseVersion, "release", "r", "0.7.0", "kyma release to use")
	cmd.Flags().StringVarP(&o.ReleaseConfig, "config", "c", "", "URL or path to the installer configuration yaml")
	cmd.Flags().BoolVarP(&o.NoWait, "noWait", "n", false, "Do not wait for completion of kyma-installer")
	cmd.Flags().StringVarP(&o.Domain, "domain", "d", "kyma.local", "domain to use for installation")
	cmd.Flags().BoolVarP(&o.Local, "local", "l", false, "Install from sources")
	cmd.Flags().StringVarP(&o.LocalSrcPath, "src-path", "", "", "Path to local sources to use")
	cmd.Flags().StringVarP(&o.LocalInstallerVersion, "installer-version", "", "", "Version of installer docker image to use while building locally")
	cmd.Flags().StringVarP(&o.LocalInstallerDir, "installer-dir", "", "", "Directory of installer docker image to use while building locally")

	return cmd
}

//Run runs the command
func (o *InstallOptions) Run() error {
	err := validateFlags(o)
	if err != nil {
		return err
	}

	s := o.NewStep(fmt.Sprintf("Checking requirements"))
	err = checkInstallRequirements(o, s)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Requirements are fine")

	if o.Local {
		fmt.Printf("Installing Kyma from local path: '%s'\n", o.LocalSrcPath)
	} else {
		fmt.Printf("Installing Kyma in version '%s'\n", o.ReleaseVersion)
	}
	fmt.Println()

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

func checkInstallRequirements(o *InstallOptions, s step.Step) error {
	versionWarning, err := kubectl.CheckVersion()
	if err != nil {
		s.Failure()
		return err
	}
	if versionWarning != "" {
		s.LogError(versionWarning)
	}
	return nil
}

func validateFlags(o *InstallOptions) error {
	if o.Local {
		if o.LocalSrcPath == "" {
			goPath := os.Getenv("GOPATH")
			if goPath == "" {
				return fmt.Errorf("No local 'src-path' configured and no applicable default found, verify if you have exported a GOPATH?")
			}
			o.LocalSrcPath = filepath.Join(goPath, "src", "github.com", "kyma-project", "kyma")
		}
		if _, err := os.Stat(o.LocalSrcPath); err != nil {
			return fmt.Errorf("Configured 'src-path=%s' does not exist, please check if you configured a valid path", o.LocalSrcPath)
		}
		if _, err := os.Stat(filepath.Join(o.LocalSrcPath, "installation", "resources")); err != nil {
			return fmt.Errorf("Configured 'src-path=%s' seems to not point to a Kyma repository, please verify if your repository contains a folder 'installation/resources'", o.LocalSrcPath)
		}

		// This is to help developer and use appropriate repository if PR image is provided
		if o.LocalInstallerDir == "" && strings.HasPrefix(o.LocalInstallerVersion, "PR-") {
			o.LocalInstallerDir = "eu.gcr.io/kyma-project/pr"
		}
	} else {
		if o.LocalSrcPath != "" {
			return fmt.Errorf("You specified 'src-path=%s' without specifying --local", o.LocalSrcPath)
		}
		if o.LocalInstallerVersion != "" {
			return fmt.Errorf("You specified 'installer-version=%s' without specifying --local", o.LocalInstallerVersion)
		}
		if o.LocalInstallerDir != "" {
			return fmt.Errorf("You specified 'installer-dir=%s' without specifying --local", o.LocalInstallerDir)
		}
	}
	return nil
}

func installTiller(o *InstallOptions) error {
	check, err := kubectl.IsPodDeployed("kube-system", "name", "tiller")
	if err != nil {
		return err
	}
	if !check {
		_, err = kubectl.RunCmd([]string{"apply", "-f", "https://raw.githubusercontent.com/kyma-project/kyma/" + o.ReleaseVersion + "/installation/resources/tiller.yaml"})
		if err != nil {
			return err
		}
	}
	err = kubectl.WaitForPodReady("kube-system", "name", "tiller")
	if err != nil {
		return err
	}
	return nil
}

func installInstaller(o *InstallOptions) error {
	check, err := kubectl.IsPodDeployed("kyma-installer", "name", "kyma-installer")
	if err != nil {
		return err
	}
	if !check {
		if o.Local {
			err = installInstallerFromLocalSources(o)
		} else {
			err = installInstallerFromRelease(o)
			if err != nil {
				return err
			}
			err = configureInstallerFromRelease(o)
		}
		if err != nil {
			return err
		}

	}
	err = kubectl.WaitForPodReady("kyma-installer", "name", "kyma-installer")
	if err != nil {
		return err
	}

	return nil
}

func installInstallerFromRelease(o *InstallOptions) error {
	relaseURL := "https://github.com/kyma-project/kyma/releases/download/" + o.ReleaseVersion + "/kyma-config-local.yaml"
	_, err := kubectl.RunCmd([]string{"apply", "-f", relaseURL})
	if err != nil {
		return err
	}
	return labelInstallerNamespace()
}

func configureInstallerFromRelease(o *InstallOptions) error {
	configURL := "https://github.com/kyma-project/kyma/releases/download/" + o.ReleaseVersion + "/kyma-config-local.yaml"
	if o.ReleaseConfig != "" {
		configURL = o.ReleaseConfig
	}
	_, err := internal.RunKubectlCmd([]string{"apply", "-f", configURL})
	if err != nil {
		return err
	}
	return nil
}

func installInstallerFromLocalSources(o *InstallOptions) error {
	localResources, err := loadLocalResources(o)
	if err != nil {
		return err
	}

	imageName, err := findInstallerImageName(localResources)
	if err != nil {
		return err
	}

	err = buildKymaInstaller(imageName, o)
	if err != nil {
		return err
	}

	err = applyKymaInstaller(localResources, o)

	return labelInstallerNamespace()
}

func findInstallerImageName(resources []map[string]interface{}) (string, error) {
	for _, res := range resources {
		if res["kind"] == "Deployment" {
			var deployment struct {
				Metadata struct {
					Name string
				}
				Spec struct {
					Template struct {
						Spec struct {
							Containers []struct {
								Image string
							}
						}
					}
				}
			}

			err := mapstructure.Decode(res, &deployment)
			if err != nil {
				return "", err
			}

			if deployment.Metadata.Name == "kyma-installer" {
				return deployment.Spec.Template.Spec.Containers[0].Image, nil
			}
		}
	}
	return "", errors.New("'kyma-installer' deployment is missing")
}

func loadLocalResources(o *InstallOptions) ([]map[string]interface{}, error) {
	resources := make([]map[string]interface{}, 0)

	resources, err := loadInstallationResourcesFile("installer-local.yaml", resources, o)
	if err != nil {
		return nil, err
	}

	resources, err = loadInstallationResourcesFile("installer-config-local.yaml.tpl", resources, o)
	if err != nil {
		return nil, err
	}

	resources, err = loadInstallationResourcesFile("installer-cr.yaml.tpl", resources, o)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func loadInstallationResourcesFile(name string, acc []map[string]interface{}, o *InstallOptions) ([]map[string]interface{}, error) {
	path := filepath.Join(o.LocalSrcPath, "installation", "resources", name)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	dec := yaml.NewDecoder(f)
	for {
		m := make(map[string]interface{})
		err := dec.Decode(m)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		acc = append(acc, m)
	}
	return acc, nil
}

func buildKymaInstaller(imageName string, o *InstallOptions) error {
	dc, err := minikube.DockerClient()
	if err != nil {
		return err
	}

	var args []docker.BuildArg
	if o.LocalInstallerDir != "" {
		args = append(args, docker.BuildArg{Name: "INSTALLER_DIR", Value: o.LocalInstallerDir})
	}
	if o.LocalInstallerVersion != "" {
		args = append(args, docker.BuildArg{Name: "INSTALLER_VERSION", Value: o.LocalInstallerVersion})
	}

	return dc.BuildImage(docker.BuildImageOptions{
		Name:         strings.TrimSpace(string(imageName)),
		Dockerfile:   filepath.Join("tools", "kyma-installer", "kyma.Dockerfile"),
		OutputStream: ioutil.Discard,
		ContextDir:   filepath.Join(o.LocalSrcPath),
		BuildArgs:    args,
	})
}

func applyKymaInstaller(resources []map[string]interface{}, o *InstallOptions) error {
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer func() { _ = stdinPipe.Close() }()
	buf := &bytes.Buffer{}
	enc := yaml.NewEncoder(buf)
	for _, y := range resources {
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
	_, err := kubectl.RunCmd([]string{"label", "namespace", "kyma-installer", "app=kyma-cli"})
	return err
}

func activateInstaller(_ *InstallOptions) error {
	status, err := kubectl.RunCmd([]string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'"})
	if err != nil {
		return err
	}
	if status == "InProgress" {
		return nil
	}

	_, err = kubectl.RunCmd([]string{"label", "installation/kyma-installation", "action=install"})
	if err != nil {
		return err
	}
	return nil
}

func printSummary(o *InstallOptions) error {
	version, err := internal.GetKymaVersion()
	if err != nil {
		return err
	}

	pwdEncoded, err := kubectl.RunCmd([]string{"-n", "kyma-system", "get", "secret", "admin-user", "-o", "jsonpath='{.data.password}'"})
	if err != nil {
		return err
	}

	pwdDecoded, err := base64.StdEncoding.DecodeString(pwdEncoded)
	if err != nil {
		return err
	}

	emailEncoded, err := kubectl.RunCmd([]string{"-n", "kyma-system", "get", "secret", "admin-user", "-o", "jsonpath='{.data.email}'"})
	if err != nil {
		return err
	}

	emailDecoded, err := base64.StdEncoding.DecodeString(emailEncoded)
	if err != nil {
		return err
	}

	clusterInfo, err := kubectl.RunCmd([]string{"cluster-info"})
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

func waitForInstaller(o *InstallOptions) error {
	currentDesc := ""
	var s step.Step
	installStatusCmd := []string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'"}

	status, err := kubectl.RunCmd(installStatusCmd)
	if err != nil {
		return err
	}
	if status == "Installed" {
		return nil
	}

	for {
		status, err := kubectl.RunCmd(installStatusCmd)
		if err != nil {
			return err
		}
		desc, err := kubectl.RunCmd([]string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.description}'"})
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
			logs, err := kubectl.RunCmd([]string{"-n", "kyma-installer", "logs", "-l", "name=kyma-installer"})
			if err != nil {
				return err
			}
			fmt.Println(logs)

		case "InProgress":
			// only do something if the description has changed
			if desc != currentDesc {
				if s != nil {
					s.Success()
				}
				s = o.NewStep(fmt.Sprintf(desc))
				currentDesc = desc
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
