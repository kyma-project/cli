package e2e

import (
	"context"
	"os"
	"os/exec"

	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"github.com/onsi/ginkgo/v2/dsl/core"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var errModuleTemplateNotApplied = errors.New("failed to apply moduletemplate")

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
	namespace string) bool {
	var deployment appsv1.Deployment
	err := k8sClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      deploymentName,
	}, &deployment)

	if err != nil || deployment.Status.AvailableReplicas == 0 {
		return false
	}

	return true
}

func IsKymaCRInReadyState(ctx context.Context,
	k8sClient client.Client,
	kymaName string,
	namespace string) bool {
	var kyma v1beta2.Kyma
	err := k8sClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      kymaName,
	}, &kyma)

	if err != nil || kyma.Status.State != v1beta2.StateReady {
		return false
	}

	return true
}

func ApplyModuleTemplate(
	moduleTemplatePath string) error {
	//cmd := exec.Command("kubectl", "apply", "-f", moduleTemplatePath)
	cmd := exec.Command("pwd")

	currentPath, err := cmd.CombinedOutput()

	core.GinkgoWriter.Println(string(currentPath))
	if err != nil {
		return err
	}

	return nil
}
