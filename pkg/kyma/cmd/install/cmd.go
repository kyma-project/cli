package install

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kyma-incubator/kyma-cli/pkg/kyma/core"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/kyma-incubator/kyma-cli/internal/helm"
	"github.com/kyma-incubator/kyma-cli/internal/minikube"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"

	"github.com/kyma-incubator/kyma-cli/internal"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
)

type command struct {
	opts *Options
	core.Command
}

const (
	sleep                = 10 * time.Second
	releaseSrcUrlPattern = "https://raw.githubusercontent.com/kyma-project/kyma/%s/%s"
	releaseUrlPattern    = "https://github.com/kyma-project/kyma/releases/download/%s/%s"
)

//NewCmd creates a new kyma command
func NewCmd(o *Options) *cobra.Command {

	cmd := command{
		Command: core.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
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
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"i"},
	}

	cobraCmd.Flags().StringVarP(&o.ReleaseVersion, "release", "r", "1.0.0", "kyma release to use")
	cobraCmd.Flags().StringVarP(&o.ReleaseConfig, "config", "c", "", "URL or path to the installer configuration yaml")
	cobraCmd.Flags().BoolVarP(&o.NoWait, "noWait", "n", false, "Do not wait for completion of kyma-installer")
	cobraCmd.Flags().StringVarP(&o.Domain, "domain", "d", "kyma.local", "domain to use for installation")
	cobraCmd.Flags().BoolVarP(&o.Local, "local", "l", false, "Install from sources")
	cobraCmd.Flags().StringVarP(&o.LocalSrcPath, "src-path", "", "", "Path to local sources to use")
	cobraCmd.Flags().StringVarP(&o.LocalInstallerVersion, "installer-version", "", "", "Version of installer docker image to use while building locally")
	cobraCmd.Flags().StringVarP(&o.LocalInstallerDir, "installer-dir", "", "", "Directory of installer docker image to use while building locally")
	cobraCmd.Flags().DurationVarP(&o.Timeout, "timeout", "", 0, "Timeout after which CLI should give up watching installation")
	cobraCmd.Flags().StringVarP(&o.Password, "password", "p", "", "Pre-defined cluster password")
	cobraCmd.Flags().VarP(&o.OverrideConfigs, "override", "o", "Path to YAML file with parameters to override. Multiple entries of this flag allowed")

	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	err := cmd.validateFlags()
	if err != nil {
		return err
	}

	s := cmd.NewStep("Checking requirements")
	err = cmd.checkInstallRequirements()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Requirements are fine")

	if cmd.opts.Local {
		s.LogInfof("Installing Kyma from local path: '%s'", cmd.opts.LocalSrcPath)
	} else {
		s.LogInfof("Installing Kyma in version '%s'", cmd.opts.ReleaseVersion)
	}

	s = cmd.NewStep("Installing Tiller")
	err = cmd.installTiller()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Tiller installed")

	s = cmd.NewStep("Installing kyma-installer")
	err = cmd.installInstaller()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("kyma-installer installed")

	s = cmd.NewStep("Configuring Helm")
	err = cmd.configureHelm()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Helm configured")

	s = cmd.NewStep("Requesting kyma-installer to install kyma")
	err = cmd.activateInstaller()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("kyma-installer is installing kyma")

	if !cmd.opts.NoWait {
		err = cmd.waitForInstaller()
		if err != nil {
			return err
		}
	}

	err = cmd.printSummary()
	if err != nil {
		return err
	}

	return nil
}

func (cmd *command) checkInstallRequirements() error {
	versionWarning, err := cmd.Kubectl().CheckVersion()
	if err != nil {
		cmd.CurrentStep.Failure()
		return err
	}
	if versionWarning != "" {
		cmd.CurrentStep.LogError(versionWarning)
	}
	return nil
}

func (cmd *command) validateFlags() error {
	if cmd.opts.Local {
		if cmd.opts.LocalSrcPath == "" {
			goPath := os.Getenv("GOPATH")
			if goPath == "" {
				return fmt.Errorf("No local 'src-path' configured and no applicable default found, verify if you have exported a GOPATH?")
			}
			cmd.opts.LocalSrcPath = filepath.Join(goPath, "src", "github.com", "kyma-project", "kyma")
		}
		if _, err := os.Stat(cmd.opts.LocalSrcPath); err != nil {
			return fmt.Errorf("Configured 'src-path=%s' does not exist, please check if you configured a valid path", cmd.opts.LocalSrcPath)
		}
		if _, err := os.Stat(filepath.Join(cmd.opts.LocalSrcPath, "installation", "resources")); err != nil {
			return fmt.Errorf("Configured 'src-path=%s' seems to not point to a Kyma repository, please verify if your repository contains a folder 'installation/resources'", cmd.opts.LocalSrcPath)
		}

		// This is to help developer and use appropriate repository if PR image is provided
		if cmd.opts.LocalInstallerDir == "" && strings.HasPrefix(cmd.opts.LocalInstallerVersion, "PR-") {
			cmd.opts.LocalInstallerDir = "eu.gcr.io/kyma-project/pr"
		}
	} else {
		if cmd.opts.LocalSrcPath != "" {
			return fmt.Errorf("You specified 'src-path=%s' without specifying --local", cmd.opts.LocalSrcPath)
		}
		if cmd.opts.LocalInstallerVersion != "" {
			return fmt.Errorf("You specified 'installer-version=%s' without specifying --local", cmd.opts.LocalInstallerVersion)
		}
		if cmd.opts.LocalInstallerDir != "" {
			return fmt.Errorf("You specified 'installer-dir=%s' without specifying --local", cmd.opts.LocalInstallerDir)
		}
	}
	return nil
}

func (cmd *command) installTiller() error {
	check, err := cmd.Kubectl().IsPodDeployed("kube-system", "name", "tiller")
	if err != nil {
		return err
	}
	if !check {
		_, err = cmd.Kubectl().RunCmd("apply", "-f", cmd.releaseSrcFile("/installation/resources/tiller.yaml"))
		if err != nil {
			return err
		}
	}
	err = cmd.Kubectl().WaitForPodReady("kube-system", "name", "tiller")
	if err != nil {
		return err
	}

	return nil
}

func (cmd *command) configureHelm() error {
	helmHome, err := helm.GetHelmHome()
	if err != nil {
		return err
	}

	if helmHome == "" {
		cmd.CurrentStep.LogInfof("Helm not installed")
		return nil
	}

	secret, err := cmd.Kubectl().RunCmd("-n", "kyma-installer", "--ignore-not-found=false", "get", "secret", "helm-secret", "-o", "yaml")
	if err != nil {
		return err
	}

	cfg := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(secret), &cfg)
	if err != nil {
		return err
	}

	data, ok := cfg["data"].(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("unable to get data from helm secret")
	}

	err = writeHelmFile(data, "global.helm.ca.crt", helmHome, "ca.pem")
	if err != nil {
		return err
	}

	err = writeHelmFile(data, "global.helm.tls.crt", helmHome, "cert.pem")
	if err != nil {
		return err
	}

	err = writeHelmFile(data, "global.helm.tls.key", helmHome, "key.pem")
	if err != nil {
		return err
	}

	return nil
}

func writeHelmFile(data map[interface{}]interface{}, helmData string, helmHome string, filename string) error {
	value, ok := data[helmData].(string)
	if !ok {
		return fmt.Errorf("unable to get %s from helm secret data", helmData)
	}
	valueDecoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(helmHome, filename), valueDecoded, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (cmd *command) installInstaller() error {
	check, err := cmd.Kubectl().IsPodDeployed("kyma-installer", "name", "kyma-installer")
	if err != nil {
		return err
	}
	if !check {
		if cmd.opts.Local {
			err = cmd.installInstallerFromLocalSources()
		} else {
			err = cmd.installInstallerFromRelease()
			if err != nil {
				return err
			}
			err = cmd.configureInstallerFromRelease()
		}
		if err != nil {
			return err
		}

	}
	err = cmd.Kubectl().WaitForPodReady("kyma-installer", "name", "kyma-installer")
	if err != nil {
		return err
	}

	return nil
}

func (cmd *command) installInstallerFromRelease() error {
	relaseURL := cmd.releaseFile("kyma-installer-local.yaml")
	_, err := cmd.Kubectl().RunCmd("apply", "-f", relaseURL)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *command) configureInstallerFromRelease() error {
	configURL := cmd.releaseFile("/kyma-config-local.yaml")
	if cmd.opts.ReleaseConfig != "" {
		configURL = cmd.opts.ReleaseConfig
	}

	_, err := cmd.Kubectl().RunCmd("apply", "-f", configURL)
	if err != nil {
		return err
	}

	err = cmd.applyOverrideFiles()
	if err != nil {
		return err
	}

	err = cmd.setAdminPassword()
	if err != nil {
		return err
	}

	err = cmd.patchMinikubeIP()
	if err != nil {
		return err
	}

	return nil
}

func (cmd *command) installInstallerFromLocalSources() error {
	localResources, err := cmd.loadLocalResources()
	if err != nil {
		return err
	}

	imageName, err := cmd.findInstallerImageName(localResources)
	if err != nil {
		return err
	}

	err = cmd.buildKymaInstaller(imageName)
	if err != nil {
		return err
	}

	_, err = cmd.Kubectl().RunApplyCmd(localResources)
	if err != nil {
		return err
	}

	err = cmd.applyOverrideFiles()
	if err != nil {
		return err
	}

	err = cmd.setAdminPassword()
	if err != nil {
		return err
	}

	err = cmd.patchMinikubeIP()
	if err != nil {
		return err
	}

	return nil
}

func (cmd *command) findInstallerImageName(resources []map[string]interface{}) (string, error) {
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

func (cmd *command) loadLocalResources() ([]map[string]interface{}, error) {
	resources := make([]map[string]interface{}, 0)

	resources, err := cmd.loadInstallationResourcesFile("installer-local.yaml", resources)
	if err != nil {
		return nil, err
	}

	resources, err = cmd.loadInstallationResourcesFile("installer-config-local.yaml.tpl", resources)
	if err != nil {
		return nil, err
	}

	resources, err = cmd.loadInstallationResourcesFile("installer-cr.yaml.tpl", resources)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func (cmd *command) loadInstallationResourcesFile(name string, acc []map[string]interface{}) ([]map[string]interface{}, error) {
	path := filepath.Join(cmd.opts.LocalSrcPath, "installation", "resources", name)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
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

func (cmd *command) buildKymaInstaller(imageName string) error {
	dc, err := minikube.DockerClient(cmd.opts.Verbose)
	if err != nil {
		return err
	}

	var args []docker.BuildArg
	if cmd.opts.LocalInstallerDir != "" {
		args = append(args, docker.BuildArg{Name: "INSTALLER_DIR", Value: cmd.opts.LocalInstallerDir})
	}
	if cmd.opts.LocalInstallerVersion != "" {
		args = append(args, docker.BuildArg{Name: "INSTALLER_VERSION", Value: cmd.opts.LocalInstallerVersion})
	}

	return dc.BuildImage(docker.BuildImageOptions{
		Name:         strings.TrimSpace(string(imageName)),
		Dockerfile:   filepath.Join("tools", "kyma-installer", "kyma.Dockerfile"),
		OutputStream: ioutil.Discard,
		ContextDir:   filepath.Join(cmd.opts.LocalSrcPath),
		BuildArgs:    args,
	})
}

func (cmd *command) activateInstaller() error {
	status, err := cmd.Kubectl().RunCmd("get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'")
	if err != nil {
		return err
	}
	if status == "InProgress" {
		return nil
	}

	_, err = cmd.Kubectl().RunCmd("label", "installation/kyma-installation", "action=install")
	if err != nil {
		return err
	}
	return nil
}

func (cmd *command) applyOverrideFiles() error {
	oFiles := cmd.opts.OverrideConfigs.Len()
	if oFiles == 0 {
		return nil
	}

	for _, file := range cmd.opts.OverrideConfigs {
		oFile, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("unable to open file: %s. Error: %s",
				file, err.Error())
		}
		rawData, err := ioutil.ReadAll(oFile)
		if err != nil {
			return fmt.Errorf("unable to read data from file: %s. Error: %s",
				file, err.Error())
		}

		configs := strings.Split(string(rawData), "---")

		for _, c := range configs {
			cfg := make(map[interface{}]interface{})
			err = yaml.Unmarshal([]byte(c), &cfg)
			if err != nil {
				return fmt.Errorf("unable to parse file data: %s. Error: %s",
					file, err.Error())
			}

			kind, ok := cfg["kind"].(string)
			if !ok {
				return fmt.Errorf("unable get kind of config. File: %s", file)
			}

			meta, ok := cfg["metadata"].(map[interface{}]interface{})
			if !ok {
				return fmt.Errorf("unable to get metadata from config. File: %s", file)
			}

			namespace, ok := meta["namespace"].(string)
			if !ok {
				return fmt.Errorf("unable to get namespace from config. File: %s", file)
			}

			name, ok := meta["name"].(string)
			if !ok {
				return fmt.Errorf("unable to get name from config. File: %s", file)
			}

			_, err := cmd.Kubectl().RunCmd("-n",
				strings.ToLower(namespace),
				"patch",
				kind,
				strings.ToLower(name),
				"-p",
				c)
			if err != nil {
				return fmt.Errorf("unable to override values. File: %s. Error: %s", file, err.Error())
			}
		}

	}

	return nil
}

func (cmd *command) setAdminPassword() error {
	if cmd.opts.Password == "" {
		return nil
	}
	encPass := base64.StdEncoding.EncodeToString([]byte(cmd.opts.Password))
	_, err := cmd.Kubectl().RunCmd("-n", "kyma-installer", "patch", "configmap", "installation-config-overrides", "--type=json",
		fmt.Sprintf("--patch=[{'op': 'replace', 'path': '/data/global.adminPassword', 'value': '%s'}]", encPass))
	return err
}

func (cmd *command) printSummary() error {
	version, err := internal.GetKymaVersion(cmd.opts.Verbose)
	if err != nil {
		return err
	}

	pwdEncoded, err := cmd.Kubectl().RunCmd("-n", "kyma-system", "get", "secret", "admin-user", "-o", "jsonpath='{.data.password}'")
	if err != nil {
		return err
	}

	pwdDecoded, err := base64.StdEncoding.DecodeString(pwdEncoded)
	if err != nil {
		return err
	}

	emailEncoded, err := cmd.Kubectl().RunCmd("-n", "kyma-system", "get", "secret", "admin-user", "-o", "jsonpath='{.data.email}'")
	if err != nil {
		return err
	}

	emailDecoded, err := base64.StdEncoding.DecodeString(emailEncoded)
	if err != nil {
		return err
	}

	clusterInfo, err := cmd.Kubectl().RunCmd("cluster-info")
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(clusterInfo)
	fmt.Println()
	fmt.Printf("Kyma is installed in version %s\n", version)
	fmt.Printf("Kyma console:     https://console.%s\n", cmd.opts.Domain)
	fmt.Printf("Kyma admin email: %s\n", emailDecoded)
	fmt.Printf("Kyma admin pwd:   %s\n", pwdDecoded)
	fmt.Println()
	fmt.Println("Happy Kyma-ing! :)")
	fmt.Println()

	return nil
}

func (cmd *command) waitForInstaller() error {
	currentDesc := ""
	_ = cmd.NewStep("Waiting for installation to start")

	status, err := cmd.Kubectl().RunCmd("get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'")
	if err != nil {
		return err
	}
	if status == "Installed" {
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
			_ = cmd.printInstallationErrorLog()
			return errors.New("Timeout while awaiting installation to complete")
		default:
			status, desc, err := cmd.getInstallationStatus()
			if err != nil {
				return err
			}

			switch status {
			case "Installed":
				cmd.CurrentStep.Success()
				return nil

			case "Error":
				if !errorOccured {
					errorOccured = true
					cmd.CurrentStep.Failuref("Error installing Kyma: %s", desc)
					cmd.CurrentStep.LogInfof("To fetch the logs from the installer execute: 'kubectl logs -n kyma-installer -l name=kyma-installer'")
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

func (cmd *command) getInstallationStatus() (status string, desc string, err error) {
	status, err = cmd.Kubectl().RunCmd("get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'")
	if err != nil {
		return
	}
	desc, err = cmd.Kubectl().RunCmd("get", "installation/kyma-installation", "-o", "jsonpath='{.status.description}'")
	return
}

func (cmd *command) printInstallationErrorLog() error {
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

func (cmd *command) releaseSrcFile(path string) string {
	return fmt.Sprintf(releaseSrcUrlPattern, cmd.opts.ReleaseVersion, path)
}

func (cmd *command) releaseFile(path string) string {
	return fmt.Sprintf(releaseUrlPattern, cmd.opts.ReleaseVersion, path)
}

func (cmd *command) patchMinikubeIP() error {
	minikubeIP, err := minikube.RunCmd(cmd.opts.Verbose, "ip")
	if err != nil {
		return err
	}
	minikubeIP = strings.TrimSpace(minikubeIP)

	patchMap := map[string][]string{
		"configmap/application-connector-overrides": []string{"application-registry.minikubeIP"},
		"configmap/core-overrides":                  []string{"test.acceptance.ui.minikubeIP", "apiserver-proxy.minikubeIP", "configurations-generator.minikubeIP"},
		"configmap/assetstore-overrides":            []string{"asset-store-controller-manager.minikubeIP", "test.integration.minikubeIP"},
	}
	for k, v := range patchMap {
		for _, pData := range v {
			_, err := cmd.Kubectl().RunCmd("-n", "kyma-installer", "patch", k, "--type=json",
				fmt.Sprintf("--patch=[{'op': 'replace', 'path': '/data/%s', 'value': '%s'}]", pData, minikubeIP))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
