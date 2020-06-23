package installation

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	installationSDK "github.com/kyma-incubator/hydroform/install/installation"
	"github.com/kyma-project/cli/cmd/kyma/version"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	pkgErrors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	releaseResourcePattern = "https://raw.githubusercontent.com/kyma-project/kyma/%s/installation/resources/%s"
	registryImagePattern   = "eu.gcr.io/kyma-project/kyma-installer:%s"
	defaultDomain          = "kyma.local"
	sourceLatest           = "latest"
	sourceLatestPublished  = "latest-published"
	sourceLocal            = "local"

	installerFile       = "installer"
	tillerFile          = "tiller"
	installerCRFile     = "installerCR"
	installerConfigFile = "installerConfig"
)

// ComponentsConfig is used to parse component list from the configuration
type ComponentsConfig struct {
	Components []v1alpha1.KymaComponent `json:"components"`
}

// Installation contains the installation elements and configuration options.
type Installation struct {
	k8s         kube.KymaKube
	service     Service
	currentStep step.Step
	// Factory contains the option to determine the interactivity of a Step.
	// +optional
	Factory step.Factory `json:"factory,omitempty"`
	// Options holds the configuration options for the installation.
	Options *Options `json:"options"`
}

// File represents a Kyma installation yaml file in the form of a key value map
// Type alias for clarity; It is still a map slice and can be used anywhere where []map[string]interface{} is used
type File struct {
	Path          string
	Content       []map[string]interface{}
	StringContent string
}

// Result contains the resulting details related to the installation.
type Result struct {
	// KymaVersion indicates the installed Kyma version.
	KymaVersion string
	// Host indicates the host address where Kyma is installed.
	Host string
	// Console holds the address of Kyma console.
	Console string
	// AdminEmail indicates the Email address of the Admin user which can be used to login Kyma.
	AdminEmail string
	// AdminPassword indicates the password of the Admin user which can be used to login Kyma.
	AdminPassword string
	// Warnings includes a set of any warnings from the installation.
	Warnings []string
}

func (i *Installation) newStep(msg string) step.Step {
	s := i.Factory.NewStep(msg)
	i.currentStep = s
	return s
}

// InstallKyma triggers the installation of a Kyma cluster.
func (i *Installation) InstallKyma() (*Result, error) {
	if i.Options.CI || i.Options.NonInteractive {
		i.Factory.NonInteractive = true
	}

	var err error
	if i.k8s, err = kube.NewFromConfigWithTimeout("", i.Options.KubeconfigPath, i.Options.Timeout); err != nil {
		return nil, pkgErrors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	s := i.newStep("Checking existence of previous installation")
	prevInstallationState, err := i.checkPrevInstallation()
	if err != nil {
		s.Failure()
		return nil, err
	}
	s.Successf("Existence of previous installation checked")

	if prevInstallationState == installationSDK.NoInstallationState {
		s = i.newStep("Validating configurations")
		if err := i.validateConfigurations(); err != nil {
			s.Failure()
			return nil, err
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

		s = i.newStep("Loading installation files")
		files, err := i.prepareFiles()
		if err != nil {
			s.Failure()
			return nil, err
		}
		s.Successf("Installation files loaded")

		s = i.newStep("Requesting Kyma Installer to install Kyma")
		if err := i.installInstaller(files); err != nil {
			s.Failure()
			return nil, err
		}
		s.Successf("Kyma Installer is installing Kyma")
	}

	if prevInstallationState != "Installed" && !i.Options.NoWait {
		if err := i.waitForInstaller(prevInstallationState); err != nil {
			return nil, err
		}
	}

	result, err := i.buildResult()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (i *Installation) checkPrevInstallation() (string, error) {
	var err error
	i.service, err = NewInstallationService(i.k8s.Config(), i.Options.Timeout, "")
	if err != nil {
		return "", fmt.Errorf("Failed to create installation service. Make sure your kubeconfig is valid: %s", err.Error())
	}

	prevInstallationState, err := i.service.CheckInstallationState(i.k8s.Config())
	if err != nil {
		installErr := installationSDK.InstallationError{}
		if errors.As(err, &installErr) {
			prevInstallationState.State = "Error"
		} else {
			return "", fmt.Errorf("Failed to get installation state: %s", err.Error())
		}
	}

	var kymaVersion string
	if prevInstallationState.State != installationSDK.NoInstallationState {
		kymaVersion, err = version.KymaVersion(i.Options.Verbose, i.k8s)
		if err != nil {
			return "", err
		}
	}

	switch prevInstallationState.State {
	case "Installed":
		i.currentStep.LogInfof("Kyma is already installed in version %s", kymaVersion)

	case "InProgress", "Error":
		// when installation is in in "Error" state, it doesn't mean that the installation has failed
		// Installer might sill recover from the error and install Kyma successfully
		i.currentStep.LogInfof("Installation in version %s is already in progress", kymaVersion)

	case "":
		return "", fmt.Errorf("Failed to get the installation status")
	}

	return prevInstallationState.State, nil
}

func (i *Installation) validateConfigurations() error {
	switch {
	//Install from local sources
	case strings.EqualFold(i.Options.Source, sourceLocal):
		i.Options.fromLocalSources = true
		if i.Options.LocalSrcPath == "" {
			goPath := os.Getenv("GOPATH")
			if goPath == "" {
				return fmt.Errorf("no 'src-path' configured and no applicable default found. Check if you exported a GOPATH")
			}
			i.Options.LocalSrcPath = filepath.Join(goPath, "src", "github.com", "kyma-project", "kyma")
		}
		if _, err := os.Stat(i.Options.LocalSrcPath); err != nil {
			return fmt.Errorf("configured 'src-path=%s' does not exist. Check if you configured a valid path", i.Options.LocalSrcPath)
		}
		if _, err := os.Stat(filepath.Join(i.Options.LocalSrcPath, "installation", "resources")); err != nil {
			return fmt.Errorf("configured 'src-path=%s' does not seem to point to a Kyma repository. Check if your repository contains the 'installation/resources' folder", i.Options.LocalSrcPath)
		}

	//Install the latest version (latest master)
	case strings.EqualFold(i.Options.Source, sourceLatest):
		latest, err := i.getMasterHash()
		if err != nil {
			return pkgErrors.Wrap(err, "unable to get latest version of kyma")
		}
		i.Options.releaseVersion = fmt.Sprintf("master-%s", latest)
		i.Options.configVersion = "master"
		i.Options.registryTemplate = registryImagePattern

	case strings.EqualFold(i.Options.Source, sourceLatestPublished):
		latest, err := i.getLatestAvailableMasterHash()
		if err != nil {
			return pkgErrors.Wrap(err, "unable to get latest published version of kyma")
		}
		i.Options.releaseVersion = fmt.Sprintf("master-%s", latest)
		i.Options.configVersion = "master"
		i.Options.registryTemplate = registryImagePattern

	//Install the specific version from release (ex: 1.3.0)
	case isSemVer(i.Options.Source):
		i.Options.releaseVersion = i.Options.Source
		i.Options.configVersion = i.Options.Source
		i.Options.registryTemplate = registryImagePattern

	//Install the specific commit hash (e.g. 34edf09a)
	case isHex(i.Options.Source):
		i.Options.releaseVersion = fmt.Sprintf("master-%s", i.Options.Source[:8])
		i.Options.configVersion = i.Options.Source
		i.Options.registryTemplate = registryImagePattern

	//Install the kyma with the specific installer image (docker image URL)
	case isDockerImage(i.Options.Source):
		i.Options.remoteImage = i.Options.Source
		i.Options.configVersion = "master"
	default:
		return fmt.Errorf("failed to parse the source flag. It can take one of the following: 'local', 'latest', 'latest-published', release version (e.g. 1.4.1), commit hash (e.g. 34edf09a) or installer image")
	}

	// If one of the --domain, --tlsKey, or --tlsCert is specified, the others must be specified as well (XOR logic used below)
	if ((i.Options.Domain != defaultDomain && i.Options.Domain != "") || i.Options.TLSKey != "" || i.Options.TLSCert != "") &&
		!((i.Options.Domain != defaultDomain && i.Options.Domain != "") && i.Options.TLSKey != "" && i.Options.TLSCert != "") {
		return pkgErrors.New("You specified one of the --domain, --tlsKey, or --tlsCert without specifying the others. They must be specified together")
	}

	return nil
}

func (i *Installation) prepareFiles() (map[string]*File, error) {
	files, err := i.loadInstallationFiles()
	if err != nil {
		return nil, err
	}

	//In case of local installation from local sources, build installer image.
	//TODO: add image build & push functionality for remote installation from local sources.
	if i.Options.fromLocalSources && i.Options.IsLocal {
		imageName, err := getInstallerImage(files[installerFile])
		if err != nil {
			return nil, err
		}

		err = i.buildKymaInstaller(imageName)
		if err != nil {
			return nil, err
		}
	} else if !i.Options.fromLocalSources {
		if i.Options.remoteImage != "" {
			err = replaceInstallerImage(files[installerFile], i.Options.remoteImage)
		} else {
			err = replaceInstallerImage(files[installerFile], buildDockerImageString(i.Options.registryTemplate, i.Options.releaseVersion))
		}
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

func (i *Installation) installInstaller(files map[string]*File) error {
	componentList, err := i.loadComponentsConfig()
	if err != nil {
		return fmt.Errorf("Could not load components configuration file. Make sure file is a valid YAML and contains component list: %s", err.Error())
	}

	i.service, err = NewInstallationServiceWithComponents(i.k8s.Config(), i.Options.Timeout, "", componentList)
	if err != nil {
		return fmt.Errorf("Failed to create installation service. Make sure your kubeconfig is valid: %s", err.Error())
	}

	files, err = loadStringContent(files)
	if err != nil {
		return fmt.Errorf("Failed to load installation files: %s", err.Error())
	}

	tillerFileContent := files[tillerFile].StringContent
	mergedInstallerFileContent := files[installerFile].StringContent + "---\n" + files[installerCRFile].StringContent
	configuration, err := i.loadConfigurations(files)
	if err != nil {
		return pkgErrors.Wrap(err, "unable to load the configurations")
	}

	err = i.service.TriggerInstallation(i.k8s.Config(), tillerFileContent, mergedInstallerFileContent, configuration)
	if err != nil {
		return fmt.Errorf("Failed to start installation: %s", err.Error())
	}

	return i.k8s.WaitPodStatusByLabel("kyma-installer", "name", "kyma-installer", corev1.PodRunning)
}

func (i *Installation) waitForInstaller(prevInstallationStatus string) error {
	currentDesc := ""
	if prevInstallationStatus == installationSDK.NoInstallationState {
		i.newStep("Waiting for installation to start")
	} else {
		i.newStep("Re-attaching installation status")
	}

	var errorOccured bool
	var timeout <-chan time.Time
	if i.Options.Timeout > 0 {
		timeout = time.After(i.Options.Timeout)
	}

	for {
		select {
		case <-timeout:
			i.currentStep.Failure()
			if _, err := i.service.CheckInstallationState(i.k8s.Config()); err != nil {
				installationError := installationSDK.InstallationError{}
				if ok := errors.As(err, &installationError); ok {
					i.currentStep.LogErrorf("Installation error occurred while installing Kyma: %s. Details: %s", installationError.Error(), installationError.Details())
				}
			}
			return errors.New("Timeout reached while waiting for installation to complete")
		default:
			installationState, err := i.service.CheckInstallationState(i.k8s.Config())
			if err != nil {
				installErr := installationSDK.InstallationError{}
				if errors.As(err, &installErr) {
					if !errorOccured {
						errorOccured = true
						i.currentStep.LogErrorf("%s, which may be OK. Will retry later...", installErr.Error())
						i.currentStep.LogInfo("To fetch the error logs from the installer, run: kubectl get installation kyma-installation -o go-template --template='{{- range .status.errorLog }}{{printf \"%s:\\n %s\\n\" .component .log}}{{- end}}'")
						i.currentStep.LogInfo("To fetch the application logs from the installer, run: kubectl logs -n kyma-installer -l name=kyma-installer")
					}
					time.Sleep(10 * time.Second)
					continue
				} else {
					return fmt.Errorf("Failed to get installation state: %s", err.Error())
				}
			}

			switch installationState.State {
			case "Installed":
				i.currentStep.Success()
				return nil

			case "InProgress":
				errorOccured = false
				// only do something if the description has changed
				if installationState.Description != currentDesc {
					i.currentStep.Success()
					i.currentStep = i.newStep(installationState.Description)
					currentDesc = installationState.Description
				}

			case "":
				i.currentStep.LogInfo("Failed to get the installation status. Will retry later...")

			default:
				i.currentStep.Failure()
				return fmt.Errorf("unexpected status: %s", installationState.State)
			}
			time.Sleep(10 * time.Second)
		}
	}
}

func (i *Installation) buildResult() (*Result, error) {
	// In case that noWait flag is set, check that Kyma was actually installed before building the Result
	if i.Options.NoWait {
		installationState, err := i.service.CheckInstallationState(i.k8s.Config())
		if err != nil {
			return nil, err
		}
		if installationState.State != "Installed" {
			return nil, nil
		}
	}

	v, err := version.KymaVersion(i.Options.Verbose, i.k8s)
	if err != nil {
		return nil, err
	}

	adm, err := i.k8s.Static().CoreV1().Secrets("kyma-system").Get("admin-user", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var consoleURL string
	vs, err := i.k8s.Istio().NetworkingV1alpha3().VirtualServices("kyma-system").Get("console-web", metav1.GetOptions{})
	switch {
	case apiErrors.IsNotFound(err):
		consoleURL = "not installed"
	case err != nil:
		return nil, err
	case vs != nil && vs.Spec != nil && len(vs.Spec.Hosts) > 0:
		consoleURL = fmt.Sprintf("https://%s", vs.Spec.Hosts[0])
	default:
		return nil, pkgErrors.New("console host could not be obtained")
	}

	var warning string
	if !i.Options.IsLocal && i.Options.Domain != defaultDomain {
		warning = "To access the console, configure DNS for the cluster load balancer: https://kyma-project.io/docs/#installation-install-kyma-with-your-own-domain-configure-dns-for-the-cluster-load-balancer"
	}

	return &Result{
		KymaVersion:   v,
		Host:          i.k8s.Config().Host,
		Console:       consoleURL,
		AdminEmail:    string(adm.Data["email"]),
		AdminPassword: string(adm.Data["password"]),
		Warnings:      []string{warning},
	}, nil
}

func (i *Installation) releaseFile(path string) string {
	return fmt.Sprintf(releaseResourcePattern, i.Options.configVersion, path)
}
