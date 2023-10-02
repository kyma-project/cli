package e2e

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	. "github.com/onsi/ginkgo/v2" //nolint:stylecheck,revive
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1extensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	errKymaDeployCommandFailed  = errors.New("failed to run kyma alpha deploy")
	errModuleTemplateNotApplied = errors.New("failed to apply ModuleTemplate")
	errModuleEnablingFailed     = errors.New("failed to enable module")
	errModuleDisablingFailed    = errors.New("failed to disable module")
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

	GinkgoWriter.Println("Available replicas:", deployment.Status.AvailableReplicas)
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

func EnableModuleOnKymaWithReadyStateModule(moduleName string) error {
	cmd := exec.Command("kyma", "alpha", "enable", "module", moduleName, "-w")
	err := cmd.Run()
	var exitCode int
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else {
		exitCode = cmd.ProcessState.ExitCode()
	}

	GinkgoWriter.Println("Exit code", exitCode)
	if exitCode != 0 {
		return errModuleEnablingFailed
	}
	return nil
}

func EnableModuleOnKymaWithWarningStateModule(moduleName string) error {
	cmd := exec.Command("kyma", "alpha", "enable", "module", moduleName, "-w")
	err := cmd.Run()
	var exitCode int
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else {
		exitCode = cmd.ProcessState.ExitCode()
	}

	GinkgoWriter.Println("Exit code", exitCode)
	if exitCode != 2 {
		return errModuleEnablingFailed
	}
	return nil
}

func DisableModuleOnKyma(moduleName string) error {
	cmd := exec.Command("kyma", "alpha", "disable", "module", moduleName, "-w")
	err := cmd.Run()
	var exitCode int
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else {
		exitCode = cmd.ProcessState.ExitCode()
	}

	GinkgoWriter.Println("Exit code", exitCode)
	if exitCode != 0 {
		return errModuleDisablingFailed
	}
	return nil
}

func IsCRDAvailable(ctx context.Context,
	k8sClient client.Client,
	name string) bool {
	var crd v1extensions.CustomResourceDefinition
	err := k8sClient.Get(ctx, client.ObjectKey{
		Name: name,
	}, &crd)

	return err == nil
}

func IsCRReady(resourceType string,
	resourceName string,
	namespace string) bool {
	cmd := exec.Command("kubectl", "get", resourceType, resourceName, "-n",
		namespace, "-o", "jsonpath='{.status.state}'")

	statusOutput, err := cmd.CombinedOutput()
	GinkgoWriter.Println(string(statusOutput))
	if err != nil || string(statusOutput) != "'Ready'" {
		return false
	}

	return true
}

func IsCRInWarningState(resourceType string,
	resourceName string,
	namespace string) bool {
	cmd := exec.Command("kubectl", "get", resourceType, resourceName, "-n",
		namespace, "-o", "jsonpath='{.status.state}'")

	statusOutput, err := cmd.CombinedOutput()
	GinkgoWriter.Println(string(statusOutput))
	if err != nil || string(statusOutput) != "'Warning'" {
		GinkgoWriter.Print(err)
		GinkgoWriter.Print(string(statusOutput) != "'Warning'")
		return false
	}

	return true
}

func IsModuleReadyInKymaStatus(ctx context.Context,
	k8sClient client.Client,
	kymaName string,
	namespace string,
	moduleName string) bool {
	var kyma v1beta2.Kyma
	err := k8sClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      kymaName,
	}, &kyma)

	GinkgoWriter.Println(kyma.Status.Modules)
	if err != nil || kyma.Status.Modules == nil || kyma.Status.Modules[0].Name != moduleName ||
		kyma.Status.Modules[0].State != v1beta2.StateReady {
		return false
	}

	return true
}

func IsModuleInWarningStateInKymaStatus(ctx context.Context,
	k8sClient client.Client,
	kymaName string,
	namespace string,
	moduleName string) bool {
	var kyma v1beta2.Kyma
	err := k8sClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      kymaName,
	}, &kyma)

	GinkgoWriter.Println(kyma.Status.Modules)
	if err != nil || kyma.Status.Modules == nil || kyma.Status.Modules[0].Name != moduleName ||
		kyma.Status.Modules[0].State != v1beta2.StateWarning {
		return false
	}

	return true
}
