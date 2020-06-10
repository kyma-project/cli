package installation

import (
	"context"
	"fmt"
	"time"

	pkgErrors "github.com/pkg/errors"

	"github.com/kyma-incubator/hydroform/install/installation"

	"k8s.io/client-go/rest"
)

const (
	tillerWaitTime = 10 * time.Minute
	installAction  = "installation"
	upgradeAction  = "upgrade"
)

type Service interface {
	CheckInstallationState(kubeconfig *rest.Config) (installation.InstallationState, error)
	TriggerInstallation(kubeconfig *rest.Config, tillerYaml string, installerYaml string, configuration installation.Configuration) error
	TriggerUpgrade(kubeconfig *rest.Config, tillerYaml string, installerYaml string, configuration installation.Configuration) error
	TriggerUninstall(kubeconfig *rest.Config) error
}

func NewInstallationService(kubeconfig *rest.Config, installationTimeout time.Duration, clusterCleanupResourceSelector string) (Service, error) {
	installer, err := installation.NewKymaInstaller(kubeconfig, installation.WithTillerWaitTime(tillerWaitTime))
	if err != nil {
		return nil, err
	}

	return &installationService{
		kymaInstallationTimeout:        installationTimeout,
		kymaInstaller:                  *installer,
		clusterCleanupResourceSelector: clusterCleanupResourceSelector,
	}, nil
}

type installationService struct {
	kymaInstallationTimeout        time.Duration
	kymaInstaller                  installation.Installer
	clusterCleanupResourceSelector string
}

func (s *installationService) TriggerInstallation(kubeconfig *rest.Config, tillerYaml string, installerYaml string, configuration installation.Configuration) error {
	return s.triggerAction(tillerYaml, installerYaml, configuration, s.kymaInstaller, s.kymaInstaller.PrepareInstallation, installAction)
}

func (s *installationService) TriggerUpgrade(kubeconfig *rest.Config, tillerYaml string, installerYaml string, configuration installation.Configuration) error {
	return s.triggerAction(tillerYaml, installerYaml, configuration, s.kymaInstaller, s.kymaInstaller.PrepareUpgrade, upgradeAction)
}

func (s *installationService) triggerAction(
	tillerYaml string,
	installerYaml string,
	configuration installation.Configuration,
	installer installation.Installer,
	prepareFunction func(installation.Installation) error,
	actionName string) error {

	installationConfig := installation.Installation{
		TillerYaml:    tillerYaml,
		InstallerYaml: installerYaml,
		Configuration: configuration,
	}

	err := prepareFunction(installationConfig)
	if err != nil {
		return pkgErrors.Wrap(err, fmt.Sprintf("Failed to prepare %s", actionName))
	}

	installationCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// We are not waiting for events, just triggering installation
	_, _, err = installer.StartInstallation(installationCtx)
	if err != nil {
		return pkgErrors.Wrap(err, fmt.Sprintf("Failed to start Kyma %s", actionName))
	}

	return nil
}

func (s *installationService) CheckInstallationState(kubeconfig *rest.Config) (installation.InstallationState, error) {
	return installation.CheckInstallationState(kubeconfig)
}

func (s *installationService) TriggerUninstall(kubeconfig *rest.Config) error {
	return installation.TriggerUninstall(kubeconfig)
}
