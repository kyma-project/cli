package installation

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/hydroform/install/installation"
	"github.com/kyma-project/cli/pkg/installation/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
)

func TestTriggerInstallation(t *testing.T) {
	t.Parallel()
	// set all mocked outcomes
	mockInstaller := &mocks.Installer{}

	// happy installation preparation
	mockInstaller.On("PrepareInstallation", installation.Installation{
		InstallerYaml:   "installer stuff",
		InstallerCRYaml: "Installer CR stuff",
		Configuration:   installation.Configuration{},
	}).Return(nil)

	// error preparing installation
	mockInstaller.On("PrepareInstallation", installation.Installation{
		InstallerYaml:   "installer CORRUPTED stuff",
		InstallerCRYaml: "Installer CR stuff",
		Configuration:   installation.Configuration{},
	}).Return(errors.New("installer YAML corrupted"))

	// successful installation trigger
	mockInstaller.On("StartInstallation", mock.Anything).Return(nil, nil, nil)

	cases := []struct {
		name        string
		service     Service
		installCfg  installation.Installation
		shouldFail  bool
		expectedErr string
	}{
		{
			name: "Happy path",
			service: &installationService{
				clusterCleanupResourceSelector: "dummy",
				kymaInstaller:                  mockInstaller,
				kymaInstallationTimeout:        10 * time.Minute,
			},
			installCfg: installation.Installation{
				InstallerYaml:   "installer stuff",
				InstallerCRYaml: "Installer CR stuff",
				Configuration:   installation.Configuration{},
			},
		},
		{
			name: "Corrputed install yaml",
			service: &installationService{
				clusterCleanupResourceSelector: "dummy",
				kymaInstaller:                  mockInstaller,
				kymaInstallationTimeout:        10 * time.Minute,
			},
			installCfg: installation.Installation{
				InstallerYaml:   "installer CORRUPTED stuff",
				InstallerCRYaml: "Installer CR stuff",
				Configuration:   installation.Configuration{},
			},
			expectedErr: "Failed to prepare installation: installer YAML corrupted",
		},
	}

	for _, c := range cases {
		err := c.service.TriggerInstallation(c.installCfg.InstallerYaml, c.installCfg.InstallerCRYaml, c.installCfg.Configuration)
		if c.expectedErr != "" {
			require.EqualError(t, err, c.expectedErr, fmt.Sprintf("Test Case: %s", c.name))
		} else {
			require.NoError(t, err, fmt.Sprintf("Test Case: %s", c.name))
		}
	}
}

func TestTriggerUpgrade(t *testing.T) {
	t.Parallel()
	// set all mocked outcomes
	mockInstaller := &mocks.Installer{}

	// happy upgrade preparation
	mockInstaller.On("PrepareUpgrade", installation.Installation{
		InstallerYaml:   "installer stuff",
		InstallerCRYaml: "Installer CR stuff",
		Configuration:   installation.Configuration{},
	}).Return(nil)

	// error preparing upgrade
	mockInstaller.On("PrepareUpgrade", installation.Installation{
		InstallerYaml:   "installer CORRUPTED stuff",
		InstallerCRYaml: "Installer CR stuff",
		Configuration:   installation.Configuration{},
	}).Return(errors.New("installer YAML corrupted"))

	// successful installation trigger
	mockInstaller.On("StartInstallation", mock.Anything).Return(nil, nil, nil)

	cases := []struct {
		name        string
		service     Service
		installCfg  installation.Installation
		shouldFail  bool
		expectedErr string
	}{
		{
			name: "Happy path",
			service: &installationService{
				clusterCleanupResourceSelector: "dummy",
				kymaInstaller:                  mockInstaller,
				kymaInstallationTimeout:        10 * time.Minute,
			},
			installCfg: installation.Installation{
				InstallerYaml:   "installer stuff",
				InstallerCRYaml: "Installer CR stuff",
				Configuration:   installation.Configuration{},
			},
		},
		{
			name: "Corrputed install yaml",
			service: &installationService{
				clusterCleanupResourceSelector: "dummy",
				kymaInstaller:                  mockInstaller,
				kymaInstallationTimeout:        10 * time.Minute,
			},
			installCfg: installation.Installation{
				InstallerYaml:   "installer CORRUPTED stuff",
				InstallerCRYaml: "Installer CR stuff",
				Configuration:   installation.Configuration{},
			},
			expectedErr: "Failed to prepare upgrade: installer YAML corrupted",
		},
	}

	for _, c := range cases {
		err := c.service.TriggerUpgrade(c.installCfg.InstallerYaml, c.installCfg.InstallerCRYaml, c.installCfg.Configuration)
		if c.expectedErr != "" {
			require.EqualError(t, err, c.expectedErr, fmt.Sprintf("Test Case: %s", c.name))
		} else {
			require.NoError(t, err, fmt.Sprintf("Test Case: %s", c.name))
		}
	}
}

func TestGetInstallationCRModificationFunc(t *testing.T) {
	t.Parallel()
	comps := []v1alpha1.KymaComponent{
		{
			Name: "Comp 1",
		},
		{
			Name: "Comp 2",
		},
	}
	i := v1alpha1.Installation{}

	iFunc := GetInstallationCRModificationFunc(comps)

	iFunc(&i)

	require.Equal(t, comps, i.Spec.Components)
}
