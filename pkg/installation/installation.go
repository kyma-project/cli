package installation

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kyma-project/cli/cmd/kyma/version"
	"github.com/kyma-project/cli/internal/helm"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/kubectl"
	"github.com/kyma-project/cli/internal/nice"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	releaseSrcURLPattern   = "https://raw.githubusercontent.com/kyma-project/kyma/%s/%s"
	releaseResourcePattern = "https://raw.githubusercontent.com/kyma-project/kyma/%s/installation/resources/%s"
	registryImagePattern   = "eu.gcr.io/kyma-project/kyma-installer:%s"
	localDomain            = "kyma.local"
	defaultKymaVersion     = "latest"
	defaultTimeout         = 1 * time.Hour
)

// Installation contains the installation elements and configuration options.
type Installation struct {
	k8s         kube.KymaKube
	kubectl     *kubectl.Wrapper
	currentStep step.Step
	// Factory contains the option to determine the interactivity of a Step.
	// +optional
	Factory step.Factory `json:"factory,omitempty"`
	// Options holds the configuration options for the installation.
	Options *Options `json:"options"`
}

func (i *Installation) getKubectl() *kubectl.Wrapper {
	if i.kubectl == nil {
		i.kubectl = kubectl.NewWrapper(i.Options.Verbose, i.Options.KubeconfigPath)
	}
	return i.kubectl
}

func (i *Installation) newStep(msg string) step.Step {
	s := i.Factory.NewStep(msg)
	i.currentStep = s
	return s
}

// InstallKyma triggers the installation of a Kyma cluster.
func (i *Installation) InstallKyma() error {
	var err error
	if i.k8s, err = kube.NewFromConfig("", i.Options.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	s := i.newStep("Validating configurations")
	if err := i.validateConfigurations(); err != nil {
		s.Failure()
		return err
	}
	s.Successf("Configurations validated")

	s = i.newStep("Checking installation source")
	if i.Options.fromLocalSources {
		s.LogInfof("Installing Kyma from local path: '%s'", i.Options.LocalSrcPath)
	} else {
		if i.Options.releaseVersion != i.Options.configVersion {
			s.LogInfof("Using the installation configuration from '%s'", i.Options.configVersion)
		}
		if i.Options.remoteImage != "" {
			s.LogInfof("Installing Kyma with installer image '%s' ", i.Options.remoteImage)
		} else {
			s.LogInfof("Installing Kyma in version '%s' ", i.Options.releaseVersion)
		}
	}
	s.Successf("Installation source checked")

	s = i.newStep("Installing Tiller")
	if err := i.installTiller(); err != nil {
		s.Failure()
		return err
	}
	s.Successf("Tiller deployed")

	s = i.newStep("Loading installation files")
	resources, err := i.loadAndConfigureInstallationFiles()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Installation files loaded")

	s = i.newStep("Deploying Kyma Installer")
	if err := i.installInstaller(resources); err != nil {
		s.Failure()
		return err
	}
	s.Successf("Kyma Installer deployed")

	if i.Options.IsLocal {
		s = i.newStep("Adding Minikube IP to the overrides")
		err := i.patchMinikubeIP(i.Options.LocalCluster.IP)
		if err != nil {
			s.Failure()
			return err
		}
		s.Successf("Minikube IP added")
	} else {
		if i.Options.Domain != localDomain {
			s = i.newStep("Creating own domain ConfigMap")
			err := i.createOwnDomainConfigMap()
			if err != nil {
				s.Failure()
				return err
			}
			s.Successf("ConfigMap created")
		}
	}

	s = i.newStep("Configuring Helm")
	if err := i.configureHelm(); err != nil {
		s.Failure()
		return err
	}
	s.Successf("Helm configured")

	s = i.newStep("Requesting Kyma Installer to install Kyma")
	if err := i.activateInstaller(); err != nil {
		s.Failure()
		return err
	}
	s.Successf("Kyma Installer is installing Kyma")

	if !i.Options.NoWait {
		if err := i.waitForInstaller(); err != nil {
			return err
		}
	}

	if err := i.printSummary(); err != nil {
		return err
	}
	return nil
}

func (i *Installation) validateConfigurations() error {
	switch {
	//Install from local sources
	case strings.EqualFold(i.Options.Source, "local"):
		i.Options.fromLocalSources = true
		i.Options.releaseVersion = defaultKymaVersion
		i.Options.configVersion = defaultKymaVersion
		if i.Options.LocalSrcPath == "" {
			goPath := os.Getenv("GOPATH")
			if goPath == "" {
				return fmt.Errorf("No 'src-path' configured and no applicable default found. Check if you exported a GOPATH")
			}
			i.Options.LocalSrcPath = filepath.Join(goPath, "src", "github.com", "kyma-project", "kyma")
		}
		if _, err := os.Stat(i.Options.LocalSrcPath); err != nil {
			return fmt.Errorf("Configured 'src-path=%s' does not exist. Check if you configured a valid path", i.Options.LocalSrcPath)
		}
		if _, err := os.Stat(filepath.Join(i.Options.LocalSrcPath, "installation", "resources")); err != nil {
			return fmt.Errorf("Configured 'src-path=%s' does not seem to point to a Kyma repository. Check if your repository contains the 'installation/resources' folder.", i.Options.LocalSrcPath)
		}

	//Install the latest version (latest master)
	case strings.EqualFold(i.Options.Source, "latest"):
		latest, err := i.getMasterHash()
		if err != nil {
			return fmt.Errorf("Unable to get latest version of kyma: %s", err.Error())
		}
		i.Options.releaseVersion = fmt.Sprintf("master-%s", latest)
		i.Options.configVersion = "master"
		i.Options.registryTemplate = registryImagePattern

	//Install the specific version from release (ex: 1.3.0)
	case isSemVer(i.Options.Source):
		i.Options.releaseVersion = i.Options.Source
		i.Options.configVersion = i.Options.Source
		i.Options.registryTemplate = registryImagePattern

	//Install the kyma with the specific installer image (docker image URL)
	case isDockerImage(i.Options.Source):
		i.Options.remoteImage = i.Options.Source
		i.Options.configVersion = defaultKymaVersion
	default:
		return fmt.Errorf("Failed to parse the source flag. It can take one of the following: 'local', 'latest', release version (e.g. 1.4.1), or installer image")
	}

	// If one of the --domain, --tlsKey, or --tlsCert is specified, the others must be specified as well (XOR logic used below)
	if (i.Options.Domain != localDomain || i.Options.TLSKey != "" || i.Options.TLSCert != "") &&
		!(i.Options.Domain != localDomain && i.Options.TLSKey != "" && i.Options.TLSCert != "") {
		return errors.New("You specified one of the --domain, --tlsKey, or --tlsCert without specifying the others. They must be specified together")
	}

	return nil
}

func (i *Installation) installTiller() error {
	deployed, err := i.k8s.IsPodDeployedByLabel("kube-system", "name", "tiller")
	if err != nil {
		return err

	}
	if !deployed {
		_, err = i.getKubectl().RunCmd("apply", "-f", i.releaseSrcFile("/installation/resources/tiller.yaml"))
		if err != nil {
			return err
		}
	}
	return i.k8s.WaitPodStatusByLabel("kube-system", "name", "tiller", corev1.PodRunning)
}

func (i *Installation) loadAndConfigureInstallationFiles() ([]map[string]interface{}, error) {
	var installationFiles []string
	if i.Options.IsLocal {
		installationFiles = []string{"installer-local.yaml", "installer-config-local.yaml.tpl", "installer-cr.yaml.tpl"}
	} else {
		installationFiles = []string{"installer.yaml", "installer-cr-cluster.yaml.tpl"}
	}

	resources, err := i.loadInstallationResourceFiles(installationFiles, i.Options.fromLocalSources)
	if err != nil {
		return nil, err
	}

	err = removeActionLabel(&resources)
	if err != nil {
		return nil, err
	}

	if i.Options.fromLocalSources {
		imageName, err := getInstallerImage(&resources)
		if err != nil {
			return nil, err
		}

		err = i.buildKymaInstaller(imageName)
		if err != nil {
			return nil, err
		}
	} else {
		if i.Options.remoteImage != "" {
			err = replaceInstallerImage(&resources, i.Options.remoteImage)
		} else {
			err = replaceInstallerImage(&resources, buildDockerImageString(i.Options.registryTemplate, i.Options.releaseVersion))
		}
		if err != nil {
			return nil, err
		}
	}

	return resources, nil
}

func (i *Installation) loadInstallationResourceFiles(resourcePaths []string, fromLocalSources bool) ([]map[string]interface{}, error) {

	var err error
	resources := make([]map[string]interface{}, 0)

	for _, resourcePath := range resourcePaths {

		var yamlReader io.ReadCloser

		if !fromLocalSources {
			yamlReader, err = downloadFile(i.releaseFile(resourcePath))
			if err != nil {
				return nil, err
			}
		} else {
			path := filepath.Join(i.Options.LocalSrcPath, "installation",
				"resources", resourcePath)
			yamlReader, err = os.Open(path)
			if err != nil {
				return nil, err
			}
		}

		dec := yaml.NewDecoder(yamlReader)
		for {
			m := make(map[string]interface{})
			err := dec.Decode(m)
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}
			resources = append(resources, m)
		}

		yamlReader.Close()
	}

	return resources, nil
}

func (i *Installation) installInstaller(resources []map[string]interface{}) error {
	deployed, err := i.k8s.IsPodDeployedByLabel("kyma-installer", "name", "kyma-installer")
	if err != nil {
		return err
	}

	if !deployed {
		_, err := i.getKubectl().RunApplyCmd(resources)
		if err != nil {
			return err
		}

		err = i.applyOverrideFiles()
		if err != nil {
			return err
		}

		err = i.setAdminPassword()
		if err != nil {
			return err
		}
	}
	return i.k8s.WaitPodStatusByLabel("kyma-installer", "name", "kyma-installer", corev1.PodRunning)
}

func (i *Installation) applyOverrideFiles() error {
	if len(i.Options.OverrideConfigs) < 1 {
		return nil
	}

	for _, file := range i.Options.OverrideConfigs {
		oFile, err := os.Open(file)
		if err != nil {
			fmt.Printf("unable to open file: %s. error: %s\n",
				file, err.Error())
			continue
		}
		rawData, err := ioutil.ReadAll(oFile)
		if err != nil {
			fmt.Printf("unable to read data from file: %s. error: %s\n",
				file, err.Error())
			continue
		}

		configs := strings.Split(string(rawData), "---")

		for _, c := range configs {
			if c == "" {
				continue
			}

			cfg := make(map[interface{}]interface{})
			err = yaml.Unmarshal([]byte(c), &cfg)
			if err != nil {
				fmt.Printf("unable to parse file data: %s. error: %s\n",
					file, err.Error())
				continue
			}

			kind, ok := cfg["kind"].(string)
			if !ok {
				fmt.Printf("unable to retrieve the kind of config. file: %s\n", file)
				continue
			}

			meta, ok := cfg["metadata"].(map[interface{}]interface{})
			if !ok {
				fmt.Printf("unable to get metadata from config. file: %s\n", file)
				continue
			}

			namespace, ok := meta["namespace"].(string)
			if !ok {
				fmt.Printf("unable to get Namespace from config. file: %s\n", file)
				continue
			}

			name, ok := meta["name"].(string)
			if !ok {
				fmt.Printf("unable to get name from config. file: %s\n", file)
				continue
			}

			if err := i.checkIfResourcePresent(namespace, kind, name); err != nil {
				if strings.Contains(err.Error(), "not found") {
					if err := i.applyResourceFile(file); err != nil {
						fmt.Printf(
							"unable to apply file %s. error: %s\n", file, err.Error())
						continue

					}
					continue
				} else {
					fmt.Printf("unable to check if resource is installed. error: %s\n", err.Error())
					continue
				}
			}

			_, err := i.getKubectl().RunCmd("-n",
				strings.ToLower(namespace),
				"patch",
				kind,
				strings.ToLower(name),
				"--type=merge",
				"-p",
				c)
			if err != nil {
				fmt.Printf("unable to override values. File: %s. Error: %s\n", file, err.Error())
				continue
			}
		}

	}

	return nil
}

func (i *Installation) patchMinikubeIP(minikubeIP string) error {
	if _, err := i.k8s.Static().CoreV1().ConfigMaps("kyma-installer").Get("installation-config-overrides", metav1.GetOptions{}); err != nil {
		if strings.Contains(err.Error(), "not found") {
			i.currentStep.LogInfof("Resource '%s' not found, won't be patched", "configmap/installation-config-overrides")
		} else {
			return err
		}
	}

	if minikubeIP != "" {
		_, err := i.k8s.Static().CoreV1().ConfigMaps("kyma-installer").Patch("installation-config-overrides", types.JSONPatchType,
			[]byte(fmt.Sprintf("[{\"op\": \"replace\", \"path\": \"/data/global.minikubeIP\", \"value\": \"%s\"}]", minikubeIP)))
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Installation) createOwnDomainConfigMap() error {
	cm, err := i.k8s.Static().CoreV1().ConfigMaps("kyma-installer").Get("owndomain-overrides", metav1.GetOptions{})
	if err == nil && cm != nil {
		if cm.Data == nil {
			cm.Data = make(map[string]string)
		}
		cm.Data["global.domainName"] = i.Options.Domain
		cm.Data["global.tlsCrt"] = i.Options.TLSCert
		cm.Data["global.tlsKey"] = i.Options.TLSKey

		_, err = i.k8s.Static().CoreV1().ConfigMaps("kyma-installer").Update(cm)
		if err != nil {
			return err
		}

		return nil
	} else if err != nil && !strings.Contains(err.Error(), "not found") {
		return err
	}

	_, err = i.k8s.Static().CoreV1().ConfigMaps("kyma-installer").Create(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "owndomain-overrides",
			Labels: map[string]string{"installer": "overrides"},
		},
		Data: map[string]string{
			"global.domainName": i.Options.Domain,
			"global.tlsCrt":     i.Options.TLSCert,
			"global.tlsKey":     i.Options.TLSKey,
		},
	})

	return err
}

func (i *Installation) configureHelm() error {
	helmHome, err := helm.Home()
	if err != nil {
		return err
	}

	if helmHome == "" {
		i.currentStep.LogInfo("Helm not installed")
		return nil
	}

	// Wait for the job that generates the helm secret to finish
	for {
		j, err := i.k8s.Static().BatchV1().Jobs("kyma-installer").Get("helm-certs-job", metav1.GetOptions{})
		if err != nil {
			return err
		}
		if j.Status.Succeeded == 1 {
			break
		} else if j.Status.Failed == 1 {
			return errors.New("Could not generate the Helm certificate.")
		}
		time.Sleep(1 * time.Second)
	}

	secret, err := i.k8s.Static().CoreV1().Secrets("kyma-installer").Get("helm-secret", metav1.GetOptions{})
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

func (i *Installation) activateInstaller() error {
	status, err := i.getKubectl().RunCmd("get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'")
	if err != nil {
		return err
	}
	if status == "InProgress" {
		return nil
	}

	_, err = i.getKubectl().RunCmd("label", "installation/kyma-installation", "action=install")
	if err != nil {
		return err
	}
	return nil
}

func (i *Installation) waitForInstaller() error {
	currentDesc := ""
	i.newStep("Waiting for installation to start")

	status, err := i.getKubectl().RunCmd("get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'")
	if err != nil {
		return err
	}
	if status == "Installed" {
		return nil
	}

	var timeout <-chan time.Time
	var errorOccured bool
	if i.Options.Timeout > 0 {
		timeout = time.After(i.Options.Timeout)
	}

	for {
		select {
		case <-timeout:
			i.currentStep.Failure()
			if err := i.printInstallationErrorLog(); err != nil {
				fmt.Println("Error fetching installation error log, please manually check the status of the cluster.")
			}
			return errors.New("Timeout reached while waiting for installation to complete")
		default:
			status, desc, err := i.getInstallationStatus()
			if err != nil {
				// A timeout when asking for the status can happen if the cluster is under high load while installing Kyma.
				// But it should not make the CLI stop waiting immediately.
				if strings.Contains("operation timed out", err.Error()) {
					i.currentStep.LogError("Could not get the status, retrying...")
				} else {
					return err
				}
			}

			switch status {
			case "Installed":
				i.currentStep.Success()
				return nil

			case "Error":
				if !errorOccured {
					errorOccured = true
					i.currentStep.LogErrorf("%s failed, which may be OK. Will retry later...", desc)
					i.currentStep.LogInfo("To fetch the error logs from the installer, run: kubectl get installation kyma-installation -o go-template --template='{{- range .status.errorLog }}{{printf \"%s:\\n %s\\n\" .component .log}}{{- end}}'")
					i.currentStep.LogInfo("To fetch the application logs from the installer, run: kubectl logs -n kyma-installer -l name=kyma-installer")
				}

			case "InProgress":
				errorOccured = false
				// only do something if the description has changed
				if desc != currentDesc {
					i.currentStep.Success()
					i.currentStep = i.newStep(desc)
					currentDesc = desc
				}

			case "":
				i.currentStep.LogInfo("Failed to get the installation status. Will retry later...")

			default:
				i.currentStep.Failure()
				fmt.Printf("Unexpected status: %s\n", status)
				os.Exit(1)
			}
			time.Sleep(10 * time.Second)
		}
	}
}

func (i *Installation) printSummary() error {
	v, err := version.KymaVersion(i.Options.Verbose, i.k8s)
	if err != nil {
		return err
	}

	adm, err := i.k8s.Static().CoreV1().Secrets("kyma-system").Get("admin-user", metav1.GetOptions{})
	if err != nil {
		return err
	}

	var consoleURL string
	vs, err := i.k8s.Istio().NetworkingV1alpha3().VirtualServices("kyma-system").Get("core-console", metav1.GetOptions{})
	switch {
	case apiErrors.IsNotFound(err):
		consoleURL = "not installed"
	case err != nil:
		return err
	case vs != nil && vs.Spec != nil && len(vs.Spec.Hosts) > 0:
		consoleURL = fmt.Sprintf("https://%s", vs.Spec.Hosts[0])
	default:
		return errors.New("Console host could not be obtained.")
	}

	fmt.Println()
	nice.PrintKyma()
	fmt.Print(" is installed in version:\t")
	nice.PrintImportant(v)

	nice.PrintKyma()
	fmt.Print(" is running at:\t\t")
	nice.PrintImportant(i.k8s.Config().Host)

	nice.PrintKyma()
	fmt.Print(" console:\t\t\t")
	nice.PrintImportantf(consoleURL)

	nice.PrintKyma()
	fmt.Print(" admin email:\t\t")
	nice.PrintImportant(string(adm.Data["email"]))

	if i.Options.Password == "" || i.Factory.NonInteractive {
		nice.PrintKyma()
		fmt.Printf(" admin password:\t\t")
		nice.PrintImportant(string(adm.Data["password"]))
	}

	if i.Options.Domain != localDomain {
		fmt.Printf("\nTo access the console, configure DNS for the cluster load balancer: ")
		nice.PrintImportant("https://kyma-project.io/docs/#installation-use-your-own-domain--provider-domain--gke--configure-dns-for-the-cluster-load-balancer")
	}

	fmt.Printf("\nHappy ")
	nice.PrintKyma()
	fmt.Printf("-ing! :)\n\n")

	return nil
}

func (i *Installation) releaseSrcFile(path string) string {
	return fmt.Sprintf(releaseSrcURLPattern, i.Options.configVersion, path)
}

func (i *Installation) releaseFile(path string) string {
	return fmt.Sprintf(releaseResourcePattern, i.Options.configVersion, path)
}
