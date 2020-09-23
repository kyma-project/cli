package installation

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	installationSDK "github.com/kyma-incubator/hydroform/install/installation"
	"github.com/kyma-project/cli/cmd/kyma/version"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/docker"
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
	installerCRFile     = "installerCR"
	installerConfigFile = "installerConfig"
)

// ComponentsConfig is used to parse component list from the configuration
type ComponentsConfig struct {
	Components []v1alpha1.KymaComponent `json:"components"`
}

// Installation contains the installation elements and configuration options.
type Installation struct {
	Docker      docker.KymaClient
	K8s         kube.KymaKube
	Service     Service
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
	// Duration indicates the duration of the installation.
	Duration time.Duration
}

func (i *Installation) newStep(msg string) step.Step {
	s := i.Factory.NewStep(msg)
	i.currentStep = s
	return s
}

// InstallKyma triggers the installation of a Kyma cluster.
func (i *Installation) InstallKyma() (*Result, error) {
	// Start timer for the installation
	installationTimer := time.Now()

	if i.Options.CI || i.Options.NonInteractive {
		i.Factory.NonInteractive = true
	}

	s := i.newStep("Preparing installation")
	// Checking existence of previous installation
	prevInstallationState, kymaVersion, err := i.checkPrevInstallation()
	if err != nil {
		s.Failure()
		return nil, err
	}
	logInfo := i.getInstallationLogInfo(prevInstallationState, kymaVersion)

	if prevInstallationState == installationSDK.NoInstallationState || prevInstallationState == "" {
		// Validating configurations
		if err := i.validateConfigurations(); err != nil {
			s.Failure()
			return nil, err
		}

		// Checking installation source
		i.checkInstallationSource()

		// Loading installation files
		files, err := i.prepareFiles()
		if err != nil {
			s.Failure()
			return nil, err
		}

		// Requesting Kyma Installer to install Kyma
		if err := i.triggerInstallation(files); err != nil {
			s.Failure()
			return nil, err
		}
		s.Successf("Preparations done")

	} else {
		s.Successf(logInfo)
	}

	if prevInstallationState != "Installed" && !i.Options.NoWait {
		if prevInstallationState == installationSDK.NoInstallationState || prevInstallationState == "" {
			i.newStep("Waiting for installation to start")
		} else {
			i.newStep("Re-attaching installation status")
		}
		if err := i.waitForInstaller(); err != nil {
			return nil, err
		}
	}

	duration := time.Since(installationTimer)

	result, err := i.buildResult(duration)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (i *Installation) checkPrevInstallation() (string, string, error) {
	prevInstallationState, err := i.Service.CheckInstallationState(i.K8s.RestConfig())
	if err != nil {
		installErr := installationSDK.InstallationError{}
		if errors.As(err, &installErr) {
			prevInstallationState.State = "Error"
		} else {
			return "", "", fmt.Errorf("Failed to get installation state: %s", err.Error())
		}
	}

	var kymaVersion string
	if prevInstallationState.State != installationSDK.NoInstallationState && prevInstallationState.State != "" {
		kymaVersion, err = version.KymaVersion(i.K8s)
		if err != nil {
			return "", "", err
		}
	}

	return prevInstallationState.State, kymaVersion, nil
}

func (i *Installation) getInstallationLogInfo(prevInstallationState string, kymaVersion string) string {
	var logInfo string
	switch prevInstallationState {
	case "Installed":
		logInfo = fmt.Sprintf("Kyma is already installed in version '%s'", kymaVersion)

	case "InProgress", "Error":
		// when installation is in in "Error" state, it doesn't mean that the installation has failed
		// Installer might sill recover from the error and install Kyma successfully
		logInfo = fmt.Sprintf("Installation in version '%s' is already in progress", kymaVersion)
	}

	return logInfo
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

		if !i.Options.IsLocal && i.Options.CustomImage == "" {
			return pkgErrors.New("You must specify --custom-image to install Kyma from local sources to a remote cluster.")
		}

	//Install the latest version (latest master)
	case strings.EqualFold(i.Options.Source, sourceLatest):
		latest, err := getMasterHash()
		if err != nil {
			return pkgErrors.Wrap(err, "unable to get latest version of kyma")
		}
		i.Options.releaseVersion = fmt.Sprintf("master-%s", latest)
		i.Options.configVersion = "master"
		i.Options.registryTemplate = registryImagePattern

	case strings.EqualFold(i.Options.Source, sourceLatestPublished):
		latest, err := getLatestAvailableMasterHash(i.currentStep, i.Options.FallbackLevel)
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

func (i *Installation) checkInstallationSource() {
	if i.Options.fromLocalSources {
		i.currentStep.LogInfof("Installing Kyma from local path: '%s'", i.Options.LocalSrcPath)
	} else {
		if i.Options.releaseVersion != i.Options.configVersion {
			i.currentStep.LogInfof("Using the installation configuration from '%s'", i.Options.configVersion)
		}
		if i.Options.remoteImage != "" {
			i.currentStep.LogInfof("Installing Kyma with installer image '%s' ", i.Options.remoteImage)
		} else {
			i.currentStep.LogInfof("Installing Kyma in version '%s' ", i.Options.releaseVersion)
		}
	}
}

func (i *Installation) prepareFiles() (map[string]*File, error) {
	files, err := i.loadInstallationFiles()
	if err != nil {
		return nil, err
	}

	if i.Options.fromLocalSources {
		//In case of local installation from local sources, build installer image using Minikube Docker client.
		if i.Options.IsLocal {
			i.Docker, err = docker.NewKymaClient(i.Options.IsLocal, i.Options.Verbose, i.Options.LocalCluster.Profile, i.Options.Timeout)
			if err != nil {
				return nil, err
			}
			imageName, err := getInstallerImage(files[installerFile])
			if err != nil {
				return nil, err
			}

			err = i.Docker.BuildKymaInstaller(i.Options.LocalSrcPath, imageName)
			if err != nil {
				return nil, err
			}
			//In case of remote cluster installation from local sources, build installer image using default Docker client and push the image.
		} else {
			i.Docker, err = docker.NewKymaClient(i.Options.IsLocal, i.Options.Verbose, i.Options.LocalCluster.Profile, i.Options.Timeout)
			if err != nil {
				return nil, err
			}
			err = i.Docker.BuildKymaInstaller(i.Options.LocalSrcPath, i.Options.CustomImage)
			if err != nil {
				return nil, err
			}

			err = i.Docker.PushKymaInstaller(i.Options.CustomImage, i.currentStep)
			if err != nil {
				return nil, err
			}

			err = replaceInstallerImage(files[installerFile], i.Options.CustomImage)
			if err != nil {
				return nil, err
			}
		}
	} else {
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

func (i *Installation) triggerInstallation(files map[string]*File) error {
	var err error
	files, err = loadStringContent(files)
	if err != nil {
		return fmt.Errorf("Failed to load installation files: %s", err.Error())
	}

	installerFileContent := files[installerFile].StringContent
	installerCRFileContent := files[installerCRFile].StringContent
	configuration, err := i.loadConfigurations(files)
	if err != nil {
		return pkgErrors.Wrap(err, "unable to load the configurations")
	}

	err = i.Service.TriggerInstallation(installerFileContent, installerCRFileContent, configuration)
	if err != nil {
		return fmt.Errorf("Failed to start installation: %s", err.Error())
	}

	return i.K8s.WaitPodStatusByLabel("kyma-installer", "name", "kyma-installer", corev1.PodRunning)
}

func (i *Installation) waitForInstaller() error {
	currentDesc := ""
	var errorOccured bool
	var timeout <-chan time.Time
	if i.Options.Timeout > 0 {
		timeout = time.After(i.Options.Timeout)
	}

	for {
		select {
		case <-timeout:
			i.currentStep.Failure()
			if _, err := i.Service.CheckInstallationState(i.K8s.RestConfig()); err != nil {
				installationError := installationSDK.InstallationError{}
				if ok := errors.As(err, &installationError); ok {
					i.currentStep.LogErrorf("Installation error occurred while installing Kyma: %s. Details: %s", installationError.Error(), installationError.Details())
				}
			}
			return errors.New("Timeout reached while waiting for installation to complete")
		default:
			installationState, err := i.Service.CheckInstallationState(i.K8s.RestConfig())
			if err != nil {
				if !errorOccured {
					errorOccured = true
					installErr := installationSDK.InstallationError{}
					if errors.As(err, &installErr) {
						i.currentStep.LogErrorf("%s, which may be OK. Will retry later...", installErr.Error())
						i.currentStep.LogInfo("To fetch the error logs from the installer, run: kubectl get installation kyma-installation -o go-template --template='{{- range .status.errorLog }}{{printf \"%s:\\n %s\\n\" .component .log}}{{- end}}'")
						i.currentStep.LogInfo("To fetch the application logs from the installer, run: kubectl logs -n kyma-installer -l name=kyma-installer")
					} else {
						i.currentStep.LogErrorf("Failed to get installation state, which may be OK. Will retry later...\nError: %s", err)
					}
				}
				time.Sleep(10 * time.Second)
				continue
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

func (i *Installation) buildResult(duration time.Duration) (*Result, error) {
	// In case that noWait flag is set, check that Kyma was actually installed before building the Result
	if i.Options.NoWait {
		installationState, err := i.Service.CheckInstallationState(i.K8s.RestConfig())
		if err != nil {
			return nil, err
		}
		if installationState.State != "Installed" {
			return nil, nil
		}
	}

	v, err := version.KymaVersion(i.K8s)
	if err != nil {
		return nil, err
	}

	adm, err := i.K8s.Static().CoreV1().Secrets("kyma-system").Get(context.Background(), "admin-user", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var consoleURL string
	vs, err := i.K8s.Istio().NetworkingV1alpha3().VirtualServices("kyma-system").Get(context.Background(), "console-web", metav1.GetOptions{})
	switch {
	case apiErrors.IsNotFound(err):
		consoleURL = "not installed"
	case err != nil:
		return nil, err
	case vs != nil && len(vs.Spec.Hosts) > 0:
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
		Host:          i.K8s.RestConfig().Host,
		Console:       consoleURL,
		AdminEmail:    string(adm.Data["email"]),
		AdminPassword: string(adm.Data["password"]),
		Warnings:      []string{warning},
		Duration:      duration,
	}, nil
}

func (i *Installation) releaseFile(path string) string {
	return fmt.Sprintf(releaseResourcePattern, i.Options.configVersion, path)
}
