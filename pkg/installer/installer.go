package installer

import (
	installer_api "github.com/kyma-project/kyma/components/installer/pkg/apis/installer/v1alpha1"
	installer "github.com/kyma-project/kyma/components/installer/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func GetComponents(kubeConfig *rest.Config) ([]installer_api.KymaComponent, error) {
	installerClient, err := installer.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	installation, err := installerClient.InstallerV1alpha1().Installations("kyma-installer").Get("kyma-installation", metav1.GetOptions{})
	if err != nil {
		installation, err = installerClient.InstallerV1alpha1().Installations("default").Get("kyma-installation", metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}
	return installation.Spec.Components, nil
}
