package e2e

import (
	"context"
	"os"

	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"github.com/onsi/ginkgo/v2/dsl/core"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	k8sClient client.Client,
	deploymentName string,
	namespace string) (bool, error) {
	var deployment appsv1.Deployment
	err := k8sClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      deploymentName,
	}, &deployment)

	if err != nil {
		return false, err
	}

	return deployment.Status.AvailableReplicas > 0, nil
}

func IsKymaCRInReadyState(ctx context.Context,
	k8sClient client.Client,
	kymaName string,
	namespace string) (bool, error) {
	var kyma v1beta2.Kyma
	err := k8sClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      kymaName,
	}, &kyma)

	if err != nil {
		return false, err
	}

	// TODO: Remove
	core.GinkgoWriter.Println(kyma)
	return kyma.Status.State == v1beta2.StateReady, nil
}
