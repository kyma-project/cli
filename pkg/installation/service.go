package installation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kyma-project/cli/pkg/step"
	pkgErrors "github.com/pkg/errors"

	"github.com/kyma-incubator/hydroform/install/installation"

	"k8s.io/client-go/rest"
)

const (
	tillerWaitTime = 10 * time.Minute
	installAction  = "installation"
	upgradeAction  = "upgrade"
)

//go:generate mockery -name=Service
type Service interface {
	InstallKyma(kubeconfig *rest.Config, tillerYaml string, installerYaml string, configuration installation.Configuration, currentStep step.Step) error
	CheckInstallationState(kubeconfig *rest.Config) (installation.InstallationState, error)
	TriggerInstallation(kubeconfig *rest.Config, tillerYaml string, installerYaml string, configuration installation.Configuration) error
	TriggerUpgrade(kubeconfig *rest.Config, tillerYaml string, installerYaml string, configuration installation.Configuration) error
	TriggerUninstall(kubeconfig *rest.Config) error
	//PerformCleanup(kubeconfig *rest.Config) error
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

// func (s *installationService) PerformCleanup(kubeconfig *rest.Config) error {
// 	cli, err := NewServiceCatalogClient(kubeconfig)
// 	if err != nil {
// 		return err
// 	}
// 	return cli.PerformCleanup(s.clusterCleanupResourceSelector)
// }

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

func (s *installationService) InstallKyma(kubeconfig *rest.Config, tillerYaml string, installerYaml string, configuration installation.Configuration, currentStep step.Step) error {
	installationConfig := installation.Installation{
		TillerYaml:    tillerYaml,
		InstallerYaml: installerYaml,
		Configuration: configuration,
	}

	err := s.kymaInstaller.PrepareInstallation(installationConfig)
	if err != nil {
		return pkgErrors.Wrap(err, "Failed to prepare installation")
	}

	installationCtx, cancel := context.WithTimeout(context.Background(), s.kymaInstallationTimeout)
	defer cancel()

	stateChannel, errChannel, err := s.kymaInstaller.StartInstallation(installationCtx)
	if err != nil {
		return pkgErrors.Wrap(err, "Failed to start Kyma installation")
	}

	err = s.waitForInstallation(currentStep, stateChannel, errChannel)
	if err != nil {
		return pkgErrors.Wrap(err, "Error while waiting for Kyma to install")
	}

	return nil
}

func (s *installationService) CheckInstallationState(kubeconfig *rest.Config) (installation.InstallationState, error) {
	return installation.CheckInstallationState(kubeconfig)
}

func (s *installationService) TriggerUninstall(kubeconfig *rest.Config) error {
	return installation.TriggerUninstall(kubeconfig)
}

func (s *installationService) waitForInstallation(currentStep step.Step, stateChannel <-chan installation.InstallationState, errorChannel <-chan error) error {
	for {
		select {
		case state, ok := <-stateChannel:
			if !ok {
				return nil
			}
			currentStep.LogInfof("Installing Kyma. Description: %s, State: %s", state.Description, state.State)
		case err, ok := <-errorChannel:
			if !ok {
				continue
			}

			installationError := installation.InstallationError{}
			if errors.Is(err, installationError) {
				currentStep.LogInfof("Warning: installation error occurred while installing Kyma: %s. Details: %s", installationError.Error(), installationError.Details())
				continue
			}

			return fmt.Errorf("an error occurred while installing Kyma: %s.", err.Error())
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
