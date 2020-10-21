package installation

import (
	"fmt"
	"time"

	"github.com/Masterminds/semver"
	installationSDK "github.com/kyma-incubator/hydroform/install/installation"
	"github.com/kyma-project/cli/internal/net"
	pkgErrors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

// UpgradeKyma triggers the upgrade of a Kyma cluster.
func (i *Installation) UpgradeKyma() (*Result, error) {
	// Start timer for the upgrade
	upgradeTimer := time.Now()

	if i.Options.CI || i.Options.NonInteractive {
		i.Factory.NonInteractive = true
	}

	s := i.newStep("Preparing Upgrade")
	// Checking existence of previous installation
	prevInstallationState, currVersion, err := i.checkPrevInstallation()
	if err != nil {
		s.Failure()
		return nil, err
	}
	logInfo, err := i.getUpgradeLogInfo(prevInstallationState, currVersion)
	if err != nil {
		s.Failure()
		return nil, err
	}

	if prevInstallationState == "Installed" {
		// Validating configurations
		if err := i.validateConfigurations(); err != nil {
			s.Failure()
			return nil, err
		}

		// Getting target Kyma version in a suitable format to be shown to the user
		targetVersion := i.getTargetVersion()

		// Getting current Kyma semantic version (if possible)
		isCurrReleaseVersion, currSemVersion := i.getSemVersion(currVersion)

		// Getting target Kyma semantic version (if possible)
		isTargetReleaseVersion, targetSemVersion := i.getSemVersion(targetVersion)

		// Prompting upgrade warning
		if !i.Options.NonInteractive {
			if err := i.promptUpgradeWarning(isCurrReleaseVersion, currVersion, isTargetReleaseVersion, targetVersion); err != nil {
				s.Failure()
				return nil, err
			}
		}

		// Check for upgrade compatibility and prompt migration guide only if both current and target Kyma versions are release versions
		if isCurrReleaseVersion && isTargetReleaseVersion {
			// Checking upgrade compatibility
			if err := i.checkUpgradeCompatibility(currSemVersion, targetSemVersion); err != nil {
				s.Failure()
				return nil, err
			}
			if !i.Options.NonInteractive {
				// Prompting migration guide
				if err := i.promptMigrationGuide(currSemVersion, targetSemVersion); err != nil {
					s.Failure()
					return nil, err
				}
			}
		}

		// Logging current Kyma version and upgrade target version
		i.logVersionUpgrade(currVersion, targetVersion)

		// Loading upgrade files
		files, err := i.prepareFiles()
		if err != nil {
			s.Failure()
			return nil, err
		}

		// Requesting Kyma Installer to upgrade Kyma
		if err := i.triggerUpgrade(files); err != nil {
			s.Failure()
			return nil, err
		}
		s.Successf("Upgrade is ready")

	} else {
		s.Successf(logInfo)
	}

	if !i.Options.NoWait {
		if prevInstallationState == "Installed" {
			i.newStep("Waiting for upgrade to start")
		} else {
			i.newStep("Re-attaching installation status")
		}
		if err := i.waitForInstaller(); err != nil {
			return nil, err
		}
	}

	duration := time.Since(upgradeTimer)

	result, err := i.buildResult(duration)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (i *Installation) getUpgradeLogInfo(prevInstallationState string, kymaVersion string) (string, error) {
	var logInfo string
	switch prevInstallationState {
	case installationSDK.NoInstallationState, "":
		return "", fmt.Errorf("It is not possible to upgrade, since Kyma is not installed on the cluster. Run \"kyma install\" to install Kyma")

	case "InProgress", "Error":
		// when installation is in in "Error" state, it doesn't mean that the installation has failed
		// Installer might sill recover from the error and install Kyma successfully
		logInfo = fmt.Sprintf("Installation in version '%s' is already in progress", kymaVersion)
	}

	return logInfo, nil
}

func (i *Installation) getTargetVersion() string {
	if i.Options.fromLocalSources {
		return i.Options.LocalSrcPath
	} else if i.Options.remoteImage != "" {
		return i.Options.remoteImage
	} else {
		return i.Options.releaseVersion
	}
}

func (i *Installation) getSemVersion(version string) (bool, *semver.Version) {
	isReleaseVersion := true
	semVersion, err := semver.NewVersion(version)
	if err != nil {
		isReleaseVersion = false
	}
	return isReleaseVersion, semVersion
}

func (i *Installation) promptUpgradeWarning(isCurrReleaseVersion bool, currVersion string, isTargetReleaseVersion bool, targetVersion string) error {
	var warning string
	if !isCurrReleaseVersion && !isTargetReleaseVersion {
		warning = fmt.Sprintf("Current Kyma version '%s' and target Kyma version '%s' are not release versions", currVersion, targetVersion)
	} else if !isCurrReleaseVersion {
		warning = fmt.Sprintf("Current Kyma version '%s' is not a release version", currVersion)
	} else if !isTargetReleaseVersion {
		warning = fmt.Sprintf("Target Kyma version '%s' is not a release version", targetVersion)
	}

	if !isCurrReleaseVersion || !isTargetReleaseVersion {
		warning += ", so it is not possible to check the upgrade compatibility.\n" +
			"If you choose to continue the upgrade, you can compromise the functionality of your cluster.\n" +
			"Are you sure you want to continue? "
	}
	continueUpgrade := i.currentStep.PromptYesNo(warning)
	if !continueUpgrade {
		return fmt.Errorf("Aborting upgrade")
	}

	return nil
}

func (i *Installation) checkUpgradeCompatibility(currSemVersion *semver.Version, targetSemVersion *semver.Version) error {
	if currSemVersion.GreaterThan(targetSemVersion) {
		return fmt.Errorf("Current Kyma version '%s' is greater than the target version '%s'. Kyma does not support a dedicated downgrade procedure", currSemVersion.String(), targetSemVersion.String())
	} else if currSemVersion.Major() != targetSemVersion.Major() {
		return fmt.Errorf("Mismatch between current Kyma version '%s' and target version '%s' is more than one minor version", currSemVersion.String(), targetSemVersion.String())
	} else if currSemVersion.Minor() != targetSemVersion.Minor() && currSemVersion.Minor()+1 != targetSemVersion.Minor() {
		return fmt.Errorf("Mismatch between current Kyma version '%s' and target version '%s' is more than one minor version", currSemVersion.String(), targetSemVersion.String())
	}

	return nil
}

func (i *Installation) promptMigrationGuide(currSemVersion *semver.Version, targetSemVersion *semver.Version) error {
	guideURL := fmt.Sprintf(
		"https://github.com/kyma-project/kyma/blob/release-%v.%v/docs/migration-guides/%v.%v-%v.%v.md",
		targetSemVersion.Major(), targetSemVersion.Minor(),
		currSemVersion.Major(), currSemVersion.Minor(),
		targetSemVersion.Major(), targetSemVersion.Minor(),
	)
	statusCode, err := net.DoGet(guideURL)
	if err != nil {
		return fmt.Errorf("Unable to check migration guide url: %v", err)
	}
	if statusCode == 404 {
		// no migration guide for this release
		i.currentStep.LogInfof("No migration guide available for upgrade from version '%s' to version '%s'", currSemVersion.String(), targetSemVersion.String())
		return nil
	}
	if statusCode != 200 {
		return fmt.Errorf("Unexpected status code %v when checking migration guide url", statusCode)
	}

	promptMsg := fmt.Sprintf("Did you check the migration guide? '%s' ", guideURL)
	isGuideChecked := i.currentStep.PromptYesNo(promptMsg)
	if !isGuideChecked {
		return fmt.Errorf("Migration guide must be checked before Kyma upgrade")
	}
	return nil
}

func (i *Installation) logVersionUpgrade(currVersion string, targetVersion string) {
	i.currentStep.LogInfof("Upgrading Kyma from version '%s' to version '%s'", currVersion, targetVersion)
}

func (i *Installation) triggerUpgrade(files map[string]*File) error {
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

	err = i.Service.TriggerUpgrade(installerFileContent, installerCRFileContent, configuration)
	if err != nil {
		return fmt.Errorf("Failed to start upgrade: %s", err.Error())
	}

	return i.K8s.WaitPodStatusByLabel("kyma-installer", "name", "kyma-installer", corev1.PodRunning)
}
