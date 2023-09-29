package e2e

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"github.com/onsi/ginkgo/v2/dsl/core"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	errKymaDeployCommandFailed  = errors.New("failed to run kyma alpha deploy")
	errModuleTemplateNotApplied = errors.New("failed to apply ModuleTemplate")
	errModuleEnablingFailed     = errors.New("failed to enable module")
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

func ExecuteKymaDeployCommand() error {
	deployCmd := exec.Command("kyma", "alpha", "deploy")
	deployOut, err := deployCmd.CombinedOutput()
	if err != nil {
		return errKymaDeployCommandFailed
	}

	if !strings.Contains(string(deployOut), "Kyma CR deployed and Ready") {
		return errKymaDeployCommandFailed
	}

	return nil
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
	cmd := exec.Command("kubectl", "apply", "-f", moduleTemplatePath)

	_, err := cmd.CombinedOutput()
	if err != nil {
		return errModuleTemplateNotApplied
	}

	return nil
}

func EnableModuleOnKyma(moduleName string) error {
	cmd := exec.Command("kyma", "alpha", "enable", "module", moduleName)
	enableOut, err := cmd.CombinedOutput()
	core.GinkgoWriter.Println(string(err.Error()))
	if err != nil {
		return err
	}
	core.GinkgoWriter.Println(enableOut)

	return nil
}
