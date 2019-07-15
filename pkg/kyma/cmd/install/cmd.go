package install

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kyma-project/cli/pkg/kyma/cmd/version"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cli/internal/kube"

	"github.com/kyma-project/cli/internal/trust"

	"github.com/kyma-project/cli/pkg/kyma/core"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/kyma-project/cli/internal/helm"
	"github.com/kyma-project/cli/internal/minikube"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	yaml "gopkg.in/yaml.v2"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
)

type command struct {
	opts *Options
	core.Command
}

const (
	sleep                  = 10 * time.Second
	releaseSrcUrlPattern   = "https://raw.githubusercontent.com/kyma-project/kyma/%s/%s"
	releaseResourcePattern = "https://raw.githubusercontent.com/kyma-project/kyma/%s/installation/resources/%s"
	registryMasterPattern  = "eu.gcr.io/kyma-project/develop/kyma-installer:master-%s"
)

var (
	patchMap = map[string][]string{
		"configmap/application-connector-overrides": []string{
			"application-registry.minikubeIP",
		},
		"configmap/core-overrides": []string{
			"test.acceptance.ui.minikubeIP",
			"apiserver-proxy.minikubeIP",
			"iam-kubeconfig-service.minikubeIP",
			"console-backend-service.minikubeIP",
			"test.acceptance.cbs.minikubeIP",
		},
		"configmap/assetstore-overrides": []string{
			"asset-store-controller-manager.minikubeIP",
			"test.integration.minikubeIP",
		},
	}
)

//NewCmd creates a new kyma command
func NewCmd(o *Options) *cobra.Command {

	cmd := command{
		Command: core.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:   "install",
		Short: "Installs Kyma on a running Kubernetes cluster",
		Long: `Installs Kyma on a running Kubernetes cluster. For more information on the command, see https://github.com/kyma-project/cli/tree/master/pkg/kyma/docs/install.md.


`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"i"},
	}

	cobraCmd.Flags().StringVarP(&o.ReleaseVersion, "release", "r", DefaultKymaVersion, "Kyma release to use")
	cobraCmd.Flags().StringVarP(&o.ReleaseConfig, "config", "c", "", "URL or path to the Installer configuration YAML file")
	cobraCmd.Flags().BoolVarP(&o.NoWait, "noWait", "n", false, "Do not wait for the Installer configuration to complete")
	cobraCmd.Flags().StringVarP(&o.Domain, "domain", "d", "kyma.local", "Domain used for installation")
	cobraCmd.Flags().BoolVarP(&o.Local, "local", "l", false, "Install from sources")
	cobraCmd.Flags().StringVarP(&o.LocalSrcPath, "src-path", "", "", "Path to local sources")
	cobraCmd.Flags().StringVarP(&o.LocalInstallerVersion, "installer-version", "", "", "Version of the Kyma Installer Docker image used for local installation")
	cobraCmd.Flags().StringVarP(&o.LocalInstallerDir, "installer-dir", "", "", "The directory of the Kyma Installer Docker image used for local installation")
	cobraCmd.Flags().DurationVarP(&o.Timeout, "timeout", "", 30*time.Minute, "Timeout after which CLI stops watching the installation progress")
	cobraCmd.Flags().StringVarP(&o.Password, "password", "p", "", "Predefined cluster password")
	cobraCmd.Flags().VarP(&o.OverrideConfigs, "override", "o", "Path to YAML file with parameters to override. Multiple entries of this flag are allowed")

	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	var err error
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Please make sure that you have a valid kubeconfig.")
	}

	if err := cmd.validateFlags(); err != nil {
		return err
	}

	s := cmd.NewStep("Checking installation source")
	if cmd.opts.Local {
		s.LogInfof("Installing Kyma from local path: '%s'", cmd.opts.LocalSrcPath)
	} else {
		s.LogInfof("Installing Kyma in version '%s'", cmd.opts.ReleaseVersion)
	}
	s.Successf("Installation source checked")

	s = cmd.NewStep("Installing Tiller")
	if err := cmd.installTiller(); err != nil {
		s.Failure()
		return err
	}
	s.Successf("Tiller deployed")

	s = cmd.NewStep("Deploying Kyma Installer")
	if err := cmd.installInstaller(); err != nil {
		s.Failure()
		return err
	}
	s.Successf("Kyma Installer deployed")

	s = cmd.NewStep("Configuring Helm")
	if err := cmd.configureHelm(); err != nil {
		s.Failure()
		return err
	}
	s.Successf("Helm configured")

	s = cmd.NewStep("Requesting Kyma Installer to install Kyma")
	if err := cmd.activateInstaller(); err != nil {
		s.Failure()
		return err
	}
	s.Successf("Kyma Installer is installing Kyma")

	if !cmd.opts.NoWait {
		if err := cmd.waitForInstaller(); err != nil {
			return err
		}
	}
	if err := cmd.importCertificate(trust.NewCertifier(cmd.K8s)); err != nil {
		// certificate import errors do not mean installation failed
		cmd.CurrentStep.LogError(err.Error())
	}

	if err := cmd.printSummary(); err != nil {
		return err
	}
	return nil
}

func (cmd *command) validateFlags() error {
	if cmd.opts.Local {
		if cmd.opts.LocalSrcPath == "" {
			goPath := os.Getenv("GOPATH")
			if goPath == "" {
				return fmt.Errorf("No local 'src-path' configured and no applicable default found. Check if you exported a GOPATH.")
			}
			cmd.opts.LocalSrcPath = filepath.Join(goPath, "src", "github.com", "kyma-project", "kyma")
		}
		if _, err := os.Stat(cmd.opts.LocalSrcPath); err != nil {
			return fmt.Errorf("Configured 'src-path=%s' does not exist. Check if you configured a valid path.", cmd.opts.LocalSrcPath)
		}
		if _, err := os.Stat(filepath.Join(cmd.opts.LocalSrcPath, "installation", "resources")); err != nil {
			return fmt.Errorf("Configured 'src-path=%s' does not seem to point to a Kyma repository. Check if your repository contains the 'installation/resources' folder.", cmd.opts.LocalSrcPath)
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
	deployed, err := cmd.K8s.IsPodDeployedByLabel("kube-system", "name", "tiller")
	if err != nil {
		return err

	}
	if !deployed {
		_, err = cmd.Kubectl().RunCmd("apply", "-f", cmd.releaseSrcFile("/installation/resources/tiller.yaml"))
		if err != nil {
			return err
		}
	}
	return cmd.K8s.WaitPodStatusByLabel("kube-system", "name", "tiller", corev1.PodRunning)
}

func (cmd *command) configureHelm() error {
	helmHome, err := helm.Home()
	if err != nil {
		return err
	}

	if helmHome == "" {
		cmd.CurrentStep.LogInfof("Helm not installed")
		return nil
	}

	secret, err := cmd.K8s.Static().CoreV1().Secrets("kyma-installer").Get("helm-secret", metav1.GetOptions{})
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(helmHome, "ca.pem"), secret.Data["global.helm.ca.crt"], 0644)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(helmHome, "cert.pem"), secret.Data["global.helm.tls.crt"], 0644)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(helmHome, "key.pem"), secret.Data["global.helm.tls.key"], 0644)
	if err != nil {
		return err
	}
	return nil
}

func (cmd *command) installInstaller() error {
	deployed, err := cmd.K8s.IsPodDeployedByLabel("kyma-installer", "name", "kyma-installer")
	if err != nil {
		return err
	}

	if !deployed {
		if cmd.opts.Local {
			err = cmd.installInstallerFromLocalSources()
		} else {
			err = cmd.installInstallerFromRelease()
		}
		if err != nil {
			return err
		}
	}
	return cmd.K8s.WaitPodStatusByLabel("kyma-installer", "name", "kyma-installer", corev1.PodRunning)
}

func (cmd *command) downloadFile(path string) (io.ReadCloser, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(path)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (cmd *command) getMasterHash() (string, error) {
	ctx, timeoutF := context.WithTimeout(context.Background(), 1*time.Minute)
	defer timeoutF()
	r, err := git.CloneContext(ctx, memory.NewStorage(), nil,
		&git.CloneOptions{
			Depth: 1,
			URL:   "https://github.com/kyma-project/kyma",
		})
	if err != nil {
		return "", err
	}
	h, err := r.Head()
	if err != nil {
		return "", err
	}

	return h.Hash().String()[:8], nil
}

func (cmd *command) buildDockerImageString(version string) string {
	return fmt.Sprintf(registryMasterPattern, version)
}

func (cmd *command) replaceDockerImageURL(resources []map[string]interface{}, imageURL string) ([]map[string]interface{}, error) {
	for _, config := range resources {
		kind, ok := config["kind"]
		if !ok {
			continue
		}

		if kind != "Deployment" {
			continue
		}

		spec, ok := config["spec"].(map[interface{}]interface{})
		if !ok {
			continue
		}

		template, ok := spec["template"].(map[interface{}]interface{})
		if !ok {
			continue
		}

		spec, ok = template["spec"].(map[interface{}]interface{})
		if !ok {
			continue
		}

		if accName, ok := spec["serviceAccountName"]; !ok {
			continue
		} else {
			if accName != "kyma-installer" {
				continue
			}
		}

		containers, ok := spec["containers"].([]interface{})
		if !ok {
			continue
		}
		for _, c := range containers {
			container := c.(map[interface{}]interface{})
			cName, ok := container["name"]
			if !ok {
				continue
			}

			if cName != "kyma-installer-container" {
				continue
			}

			if _, ok := container["image"]; !ok {
				continue
			}
			container["image"] = imageURL
			return resources, nil
		}
	}
	return nil, errors.New("unable to find 'image' field for kyma installer 'Deployment'")
}

func (cmd *command) installInstallerFromRelease() error {
	remoteResources, err := cmd.loadResources(false)
	if err != nil {
		return err
	}

	err = cmd.removeActionLabel(remoteResources)
	if err != nil {
		return err
	}

	if strings.ToLower(cmd.opts.ReleaseVersion) == "master" {
		masterHash, err := cmd.getMasterHash()
		if err != nil {
			return err
		}

		remoteResources, err = cmd.replaceDockerImageURL(remoteResources,
			cmd.buildDockerImageString(masterHash))
		if err != nil {
			return err
		}
	}

	_, err = cmd.Kubectl().RunApplyCmd(remoteResources)
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
	localResources, err := cmd.loadResources(true)
	if err != nil {
		return err
	}

	err = cmd.removeActionLabel(localResources)
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

func (cmd *command) loadResources(isLocal bool) ([]map[string]interface{}, error) {
	resources := make([]map[string]interface{}, 0)
	resources, err := cmd.loadInstallationResourceFile("installer-local.yaml",
		isLocal, resources)
	if err != nil {
		return nil, err
	}

	resources, err = cmd.loadInstallationResourceFile("installer-config-local.yaml.tpl",
		isLocal, resources)
	if err != nil {
		return nil, err
	}

	resources, err = cmd.loadInstallationResourceFile("installer-cr.yaml.tpl",
		isLocal, resources)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func (cmd *command) removeActionLabel(acc []map[string]interface{}) error {
	for _, config := range acc {
		kind, ok := config["kind"]
		if !ok {
			continue
		}

		if kind != "Installation" {
			continue
		}

		meta, ok := config["metadata"].(map[interface{}]interface{})
		if !ok {
			return errors.New("Installation contains no METADATA section")
		}

		labels, ok := meta["labels"].(map[interface{}]interface{})
		if !ok {
			return errors.New("Installation contains no LABELS section")
		}

		_, ok = labels["action"].(string)
		if !ok {
			return nil
		}

		delete(labels, "action")

	}
	return nil
}

func (cmd *command) loadInstallationResourceFile(resourcePath string, local bool,
	acc []map[string]interface{}) ([]map[string]interface{}, error) {

	var yamlReader io.ReadCloser
	var err error

	if !local {
		yamlReader, err = cmd.downloadFile(cmd.releaseFile(resourcePath))
		if err != nil {
			return nil, err
		}
	} else {
		path := filepath.Join(cmd.opts.LocalSrcPath, "installation",
			"resources", resourcePath)
		yamlReader, err = os.Open(path)
		if err != nil {
			return nil, err
		}
	}
	defer yamlReader.Close()

	dec := yaml.NewDecoder(yamlReader)
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
			return fmt.Errorf("Unable to open file: %s. Error: %s",
				file, err.Error())
		}
		rawData, err := ioutil.ReadAll(oFile)
		if err != nil {
			return fmt.Errorf("Unable to read data from file: %s. Error: %s",
				file, err.Error())
		}

		configs := strings.Split(string(rawData), "---")

		for _, c := range configs {
			cfg := make(map[interface{}]interface{})
			err = yaml.Unmarshal([]byte(c), &cfg)
			if err != nil {
				return fmt.Errorf("Unable to parse file data: %s. Error: %s",
					file, err.Error())
			}

			kind, ok := cfg["kind"].(string)
			if !ok {
				return fmt.Errorf("Unable to retrieve the kind of config. File: %s", file)
			}

			meta, ok := cfg["metadata"].(map[interface{}]interface{})
			if !ok {
				return fmt.Errorf("Unable to get metadata from config. File: %s", file)
			}

			namespace, ok := meta["namespace"].(string)
			if !ok {
				return fmt.Errorf("Unable to get Namespace from config. File: %s", file)
			}

			name, ok := meta["name"].(string)
			if !ok {
				return fmt.Errorf("Unable to get name from config. File: %s", file)
			}

			if err := cmd.checkIfResourcePresent(namespace, kind, name); err != nil {
				if strings.Contains(err.Error(), "not found") {
					if err := cmd.applyResourceFile(file); err != nil {
						return fmt.Errorf(
							"Unable to apply file %s. Error: %s", file, err.Error())

					}
					continue
				} else {
					return fmt.Errorf("Unable to check if resource is installed. Error: %s", err.Error())
				}
			}

			_, err := cmd.Kubectl().RunCmd("-n",
				strings.ToLower(namespace),
				"patch",
				kind,
				strings.ToLower(name),
				"--type=merge",
				"-p",
				c)
			if err != nil {
				return fmt.Errorf("Unable to override values. File: %s. Error: %s", file, err.Error())
			}
		}

	}

	return nil
}

func (cmd *command) checkIfResourcePresent(namespace, kind, name string) error {
	_, err := cmd.Kubectl().RunCmd("-n", namespace, "get", kind, name)
	return err
}

func (cmd *command) applyResourceFile(filepath string) error {
	_, err := cmd.Kubectl().RunCmd("apply", "-f", filepath)
	return err
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
	v, err := version.KymaVersion(cmd.opts.Verbose, cmd.K8s)
	if err != nil {
		return err
	}

	adm, err := cmd.K8s.Static().CoreV1().Secrets("kyma-system").Get("admin-user", metav1.GetOptions{})
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
	fmt.Printf("Kyma is installed in version %s\n", v)
	fmt.Printf("Kyma console:\t\thttps://console.%s\n", cmd.opts.Domain)
	fmt.Printf("Kyma admin email:\t%s\n", adm.Data["email"])
	if cmd.opts.Password == "" || cmd.opts.NonInteractive {
		fmt.Printf("Kyma admin password:\t%s\n", adm.Data["password"])
	}
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
			return errors.New("Timeout reached while waiting for installation to complete")
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
					cmd.CurrentStep.LogErrorf("%s failed, which may be OK. Will retry later...", desc)
					cmd.CurrentStep.LogInfo("To fetch the error logs from the installer, run: kubectl get installation  kyma-installation -o go-template --template='{{- range .status.errorLog }}{{printf \"%s:\\n %s\\n\" .component .log}}{{- end}}'")
					cmd.CurrentStep.LogInfo("To fetch the application logs from the installer, run: kubectl logs -n kyma-installer -l name=kyma-installer")
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
	logs, err := cmd.Kubectl().RunCmd("get", "installation", "kyma-installation", "-o", "go-template", "--template={{- range .status.errorLog -}}{{printf \"%s:\n %s [%s]\n\" .component .log .occurrences}}{{- end}}")
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
	return fmt.Sprintf(releaseResourcePattern, cmd.opts.ReleaseVersion, path)
}

func (cmd *command) patchMinikubeIP() error {
	minikubeIP, err := minikube.RunCmd(cmd.opts.Verbose, "ip")
	if err != nil {
		cmd.CurrentStep.LogInfo("unable to perform 'minikube ip' command. Patches won't be applied")
		return nil
	}
	minikubeIP = strings.TrimSpace(minikubeIP)

	for k, v := range patchMap {
		for _, pData := range v {
			if _, err := cmd.Kubectl().RunCmd("-n", "kyma-installer", "get", k); err != nil {
				if strings.Contains(err.Error(), "not found") {
					cmd.CurrentStep.LogInfof("resource '%s' not found, won't be patched", k)
					continue
				} else {
					return err
				}
			}

			_, err = cmd.Kubectl().RunCmd("-n", "kyma-installer", "patch", k, "--type=json",
				"--allow-missing-template-keys=true",
				fmt.Sprintf("--patch=[{'op': 'replace', 'path': '/data/%s', 'value': '%s'}]", pData, minikubeIP))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (cmd *command) importCertificate(ca trust.Certifier) error {
	if !cmd.opts.NoWait {
		// get cert from cluster
		cert, err := ca.Certificate()
		if err != nil {
			return err
		}

		tmpFile, err := ioutil.TempFile(os.TempDir(), "kyma-*.crt")
		if err != nil {
			return errors.Wrap(err, "Cannot create temporary file for Kyma certificate")
		}
		defer os.Remove(tmpFile.Name())

		if _, err = tmpFile.Write(cert); err != nil {
			return errors.Wrap(err, "Failed to write the kyma certificate")
		}
		if err := tmpFile.Close(); err != nil {
			return err
		}

		if err := ca.StoreCertificate(tmpFile.Name(), cmd.CurrentStep); err != nil {
			return err
		}
		cmd.CurrentStep.Successf("Kyma root certificate imported")

	} else {
		cmd.CurrentStep.LogError(ca.Instructions())
	}
	return nil
}
