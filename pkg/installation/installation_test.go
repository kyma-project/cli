package installation

import (
	"errors"
	"os"
	"testing"
	"time"

	installSDK "github.com/kyma-incubator/hydroform/install/installation"
	k8sMocks "github.com/kyma-project/cli/internal/kube/mocks"
	"github.com/kyma-project/cli/pkg/installation/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	fakeIstio "istio.io/client-go/pkg/clientset/versioned/fake"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func TestInstallKyma(t *testing.T) {
	t.Parallel()
	// prepare mocks
	kymaMock := k8sMocks.KymaKube{}
	iServiceMock := mocks.Service{}

	// fake k8s with installer pod running and post installation resources
	k8sMock := fake.NewSimpleClientset(
		&v1.Pod{
			ObjectMeta: metaV1.ObjectMeta{Name: "kyma-installer", Namespace: "kyma-installer", Labels: map[string]string{"name": "kyma-installer"}},
			Status:     v1.PodStatus{Phase: v1.PodRunning},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "Installer",
						Image: "fake-registry/installer:1.15.1",
					},
				},
			},
		},
		&v1.Secret{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      "admin-user",
				Namespace: "kyma-system",
			},
			Data: map[string][]byte{
				"email":    []byte("admin@fake.com"),
				"password": []byte("1234-super-secure"),
			},
		},
	)

	// fake istio vService
	istioMock := fakeIstio.NewSimpleClientset(
		&v1alpha3.VirtualService{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      "console-web",
				Namespace: "kyma-system",
			},
			Spec: networkingv1alpha3.VirtualService{
				Hosts: []string{"fake-console-url"},
			},
		},
	)

	i := &Installation{
		K8s:     &kymaMock,
		Service: &iServiceMock,
		Options: &Options{
			NoWait:           false,
			NonInteractive:   true,
			Timeout:          10 * time.Minute,
			Domain:           "irrelevant",
			TLSCert:          "fake-cert",
			TLSKey:           "fake-key",
			Password:         "fake-password",
			OverrideConfigs:  nil,
			ComponentsConfig: "",
			IsLocal:          false,
			Source:           "1.15.1",
		},
	}

	kymaMock.On("Static").Return(k8sMock)
	kymaMock.On("Istio").Return(istioMock)
	kymaMock.On("RestConfig", mock.Anything).Return(&rest.Config{Host: "fake-kubeconfig-host"})

	// There is an existing installation
	iServiceMock.On("CheckInstallationState", mock.Anything).Return(installSDK.InstallationState{State: "Installed"}, nil).Once()

	r, err := i.InstallKyma()
	require.NoError(t, err)
	require.NotEmpty(t, r)

	// Installation in progress
	i.Options.NoWait = true // no need to wait for installation here
	iServiceMock.On("CheckInstallationState", mock.Anything).Return(installSDK.InstallationState{State: "InProgress"}, nil).Times(2)

	r, err = i.InstallKyma()
	require.NoError(t, err)
	require.Empty(t, r)

	// Error getting installation status
	iServiceMock.On("CheckInstallationState", mock.Anything).Return(installSDK.InstallationState{}, errors.New("installation is hiding from us")).Once()

	r, err = i.InstallKyma()
	require.Error(t, err)
	require.Empty(t, r)

	// Empty installation status will be treated the same way as a cluster with no installation, so we should have a happy path
	iServiceMock.On("CheckInstallationState", mock.Anything).Return(installSDK.InstallationState{}, nil).Once()
	iServiceMock.On("TriggerInstallation", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	iServiceMock.On("CheckInstallationState", mock.Anything).Return(installSDK.InstallationState{State: "Installed"}, nil).Once()
	kymaMock.On("WaitPodStatusByLabel", "kyma-installer", "name", "kyma-installer", v1.PodRunning).Return(nil)

	r, err = i.InstallKyma()
	require.NoError(t, err)
	require.NotEmpty(t, r)

	// Happy path
	iServiceMock.On("CheckInstallationState", mock.Anything).Return(installSDK.InstallationState{State: installSDK.NoInstallationState}, nil).Once()
	iServiceMock.On("TriggerInstallation", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	iServiceMock.On("CheckInstallationState", mock.Anything).Return(installSDK.InstallationState{State: "Installed"}, nil).Once()
	kymaMock.On("WaitPodStatusByLabel", "kyma-installer", "name", "kyma-installer", v1.PodRunning).Return(nil)

	r, err = i.InstallKyma()
	require.NoError(t, err)
	require.NotEmpty(t, r)

	// Happy path with commit ID
	iServiceMock.On("CheckInstallationState", mock.Anything).Return(installSDK.InstallationState{State: installSDK.NoInstallationState}, nil).Once()
	iServiceMock.On("TriggerInstallation", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	iServiceMock.On("CheckInstallationState", mock.Anything).Return(installSDK.InstallationState{State: "Installed"}, nil).Once()
	kymaMock.On("WaitPodStatusByLabel", "kyma-installer", "name", "kyma-installer", v1.PodRunning).Return(nil)

	i.Options.Source = "1.24.10"
	r, err = i.InstallKyma()
	require.NoError(t, err)
	require.NotEmpty(t, r)
}

func TestValidateConfigurations(t *testing.T) {
	t.Parallel()
	// Domain is passed, but certificate and key are missing
	i := &Installation{
		Options: &Options{
			Domain: "irrelevant",
			Source: "1.15.1",
		},
	}

	err := i.validateConfigurations()
	require.EqualError(t, err, errorCustomDomainCertMissing)

	// Domain, certificate, and key are passed
	i = &Installation{
		Options: &Options{
			Domain:  "irrelevant",
			TLSCert: "fake-cert",
			TLSKey:  "fake-key",
			Source:  "1.15.1",
		},
	}

	err = i.validateConfigurations()
	require.NoError(t, err)

	// One certificate is missing
	i = &Installation{
		Options: &Options{
			TLSCert: "fake-cert",
			Source:  "1.15.1",
		},
	}

	err = i.validateConfigurations()
	require.EqualError(t, err, errorCertIncomplete)

	i = &Installation{
		Options: &Options{
			TLSKey: "fake-cert",
			Source: "1.15.1",
		},
	}

	err = i.validateConfigurations()
	require.EqualError(t, err, errorCertIncomplete)

	// Unsupported profile is used
	i = &Installation{
		Options: &Options{
			Source:  "1.15.1",
			Profile: "unknown",
		},
	}

	err = i.validateConfigurations()
	require.EqualError(t, err, errorProfileNotSupported)

	// Happy path: allow custom certs without domain
	i = &Installation{
		Options: &Options{
			TLSCert: "fake-cert",
			TLSKey:  "fake-key",
			Source:  "1.15.1",
		},
	}

	err = i.validateConfigurations()
	require.NoError(t, err)

	// Create fake local source path
	fakePath := "/tmp/fake-path-for-cli-test"
	i.Options.LocalSrcPath = fakePath
	err = os.MkdirAll(fakePath+"/installation/resources", 0700)
	require.NoError(t, err)
	defer os.RemoveAll(fakePath)

	// Source "local" and local installation
	i.Options.IsLocal = true
	i.Options.Source = "local"
	err = i.validateConfigurations()
	require.NoError(t, err)

	// Source "local" and cluster installation, without custom-image
	i.Options.IsLocal = false
	i.Options.Source = "local"
	err = i.validateConfigurations()
	require.EqualError(t, err, "You must specify --custom-image to install Kyma from local sources to a remote cluster.")

	// Source "local" and cluster installation
	i.Options.IsLocal = false
	i.Options.Source = "local"
	i.Options.CustomImage = "test-registry/test-image:1.0.0"
	err = i.validateConfigurations()
	require.NoError(t, err)

	// Source "main"
	i.Options.Source = "main"
	err = i.validateConfigurations()
	require.NoError(t, err)

	// Source commit id
	i.Options.Source = "34edf09a"
	err = i.validateConfigurations()
	require.NoError(t, err)

	// Source pull request
	i.Options.Source = "PR-9486"
	err = i.validateConfigurations()
	require.NoError(t, err)

	// Source docker image
	i.Options.Source = "test-registry/test-image:1.0.0"
	err = i.validateConfigurations()
	require.NoError(t, err)

	// Unknown source
	i.Options.Source = "fake-source"
	err = i.validateConfigurations()
	require.Error(t, err)
}
