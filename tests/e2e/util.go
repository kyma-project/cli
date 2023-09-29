package e2e

import (
	"context"
	"os"

	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
)

func ReadModuleTemplate(filepath string) (*v1beta2.ModuleTemplate, error) {
	moduleTemplate := &v1beta2.ModuleTemplate{}
	moduleFile, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(moduleFile, &moduleTemplate)
	return moduleTemplate, err
}

func IsDeploymentReady(ctx context.Context,
	client *kubernetes.Clientset,
	deploymentName string,
	namespace string) (bool, error) {
	deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, deploymentName, v1.GetOptions{})

	if err != nil {
		return false, err
	}

	return deployment.Status.AvailableReplicas > 0, nil
}
