package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kyma-project/lifecycle-manager/api/shared"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	. "github.com/onsi/ginkgo/v2" //nolint:stylecheck,revive
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1extensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	errKymaDeployCommandFailed           = errors.New("failed to run kyma alpha deploy")
	errModuleTemplateNotApplied          = errors.New("failed to apply ModuleTemplate")
	errModuleEnablingFailed              = errors.New("failed to enable module")
	errModuleDisablingFailed             = errors.New("failed to disable module")
	ErrCreateModuleFailedWithSameVersion = errors.New(
		"failed to create module with same version exists message")
)

const (
	exitCodeNoError = 0
	exitCodeWarning = 2
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
		return fmt.Errorf("%w: %v", errKymaDeployCommandFailed, err)
	}

	if !strings.Contains(string(deployOut), "Kyma CR deployed and Ready") {
		return errKymaDeployCommandFailed
	}

	return nil
}

func DeploymentIsReady(ctx context.Context,
	k8sClient client.Client,
	deploymentName string,
	namespace string) bool {
	var deployment appsv1.Deployment
	err := k8sClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      deploymentName,
	}, &deployment)

	GinkgoWriter.Println("Available replicas:", deployment.Status.AvailableReplicas)
	return err == nil && deployment.Status.AvailableReplicas != 0
}

func KymaCRIsInReadyState(ctx context.Context,
	k8sClient client.Client,
	kymaName string,
	namespace string) bool {
	var kyma v1beta2.Kyma
	err := k8sClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      kymaName,
	}, &kyma)

	return err == nil && kyma.Status.State == shared.StateReady
}

func ApplyModuleTemplate(
	moduleTemplatePath string) error {
	cmd := exec.Command("kubectl", "apply", "-f", moduleTemplatePath)

	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %v", errModuleTemplateNotApplied, err)
	}

	return nil
}

func enableKymaModuleWithExpectedExitCode(moduleName string, expectedExitCode int) error {
	cmd := exec.Command("kyma", "alpha", "enable", "module", moduleName, "-w")
	err := cmd.Run()
	var exitCode int
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else {
		exitCode = cmd.ProcessState.ExitCode()
	}

	GinkgoWriter.Println("Exit code", exitCode)
	if exitCode != expectedExitCode {
		return fmt.Errorf("%w: %v", errModuleEnablingFailed, err)
	}
	return nil
}

func EnableKymaModuleWithReadyState(moduleName string) error {
	return enableKymaModuleWithExpectedExitCode(moduleName, exitCodeNoError)
}

func EnableKymaModuleWithWarningState(moduleName string) error {
	return enableKymaModuleWithExpectedExitCode(moduleName, exitCodeWarning)

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
		return fmt.Errorf("%w: %v", errModuleDisablingFailed, err)
	}
	return nil
}

func CRDIsAvailable(ctx context.Context,
	k8sClient client.Client,
	name string) bool {
	var crd v1extensions.CustomResourceDefinition
	err := k8sClient.Get(ctx, client.ObjectKey{
		Name: name,
	}, &crd)

	return err == nil
}

func CrIsInExpectedState(resourceType string,
	resourceName string,
	namespace string,
	expectedState shared.State) bool {
	cmd := exec.Command("kubectl", "get", resourceType, resourceName, "-n",
		namespace, "-o", "jsonpath='{.status.state}'")

	statusOutput, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	GinkgoWriter.Println(string(statusOutput))

	return err == nil && strings.Contains(string(statusOutput), string(expectedState))
}

func KymaContainsModuleInExpectedState(ctx context.Context,
	k8sClient client.Client,
	kymaName string,
	namespace string,
	moduleName string,
	expectedState shared.State) bool {
	var kyma v1beta2.Kyma
	err := k8sClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      kymaName,
	}, &kyma)

	GinkgoWriter.Println(kyma.Status.Modules)

	return err == nil && kyma.Status.Modules != nil && kyma.Status.Modules[0].Name == moduleName &&
		kyma.Status.Modules[0].State == expectedState
}

func ModuleResourcesAreReady(ctx context.Context,
	k8sClient client.Client,
	crdName string,
	deploymentNamesAndNamespaces map[string]string) bool {
	if !CRDIsAvailable(ctx, k8sClient, crdName) {
		return false
	}
	for k, v := range deploymentNamesAndNamespaces {
		if !DeploymentIsReady(ctx, k8sClient, k, v) {
			return false
		}
	}

	return true
}

func Flatten(labels v1.Labels) map[string]string {
	labelsMap := make(map[string]string)
	for _, l := range labels {
		var value string
		_ = yaml.Unmarshal(l.Value, &value)
		labelsMap[l.Name] = value
	}

	return labelsMap
}

func CreateModuleCommand(versionOverwrite bool,
	path, registry, configFilePath, version, secScannerConfig string) error {
	var createModuleCmd *exec.Cmd
	if versionOverwrite {
		createModuleCmd = exec.Command("kyma", "alpha", "create", "module",
			"--path", path, "--registry", registry, "--insecure", "--module-config-file", configFilePath,
			"--version", version, "--module-archive-version-overwrite", "--sec-scanners-config", secScannerConfig)
	} else {
		createModuleCmd = exec.Command("kyma", "alpha", "create", "module",
			"--path", path, "--registry", registry, "--insecure", "--module-config-file", configFilePath,
			"--version", version, "--sec-scanners-config", secScannerConfig)
	}
	createOut, err := createModuleCmd.CombinedOutput()

	if err != nil {
		if strings.Contains(string(createOut),
			fmt.Sprintf("version %s already exists with different content", version)) {
			return ErrCreateModuleFailedWithSameVersion
		}
		return fmt.Errorf("create module command failed with err %s", err)
	}
	return nil
}
