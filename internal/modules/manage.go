package modules

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modules/repo"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var ErrModuleInstalledVersionNotInKymaChannel = errors.New("version of the installed module doesn't exist in the configured release channel")

func ModuleExistsInKymaCR(ctx context.Context, client kube.Client, moduleName string) (bool, error) {
	defaultKyma, err := client.Kyma().GetDefaultKyma(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get kyma cr: %v", err)
	}
	for _, m := range defaultKyma.Spec.Modules {
		if moduleName == m.Name {
			return true, nil
		}
	}
	return false, nil
}

func ManageModuleInKymaCR(ctx context.Context, client kube.Client, moduleName, policy string) error {
	err := client.Kyma().ManageModule(ctx, moduleName, policy)
	if err != nil {
		return fmt.Errorf("failed to set module as managed: %v", err)
	}

	err = client.Kyma().WaitForModuleState(ctx, moduleName, "Ready", "Warning")
	if err != nil {
		return fmt.Errorf("failed to check module state: %v", err)
	}

	return nil
}

func ManageModuleMissingInKyma(ctx context.Context, client kube.Client, repo repo.ModuleTemplatesRepository, moduleName, policy string) error {
	installedModuleTemplate, err := findInstalledModuleTemplate(ctx, client, repo, moduleName)
	if err != nil {
		return err
	}

	moduleReleaseMetas, err := client.Kyma().ListModuleReleaseMeta(ctx)
	if err != nil {
		return fmt.Errorf("failed to get module release metas: %v", err)
	}

	channelAssignedToModuleVersion := getAssignedChannel(*moduleReleaseMetas, moduleName, installedModuleTemplate.Spec.Version)
	kymaCR, err := client.Kyma().GetDefaultKyma(ctx)
	if err != nil {
		return fmt.Errorf("failed to get kyma cr")
	}
	expectedChannel := kymaCR.Spec.Channel

	if channelAssignedToModuleVersion != expectedChannel {
		return ErrModuleInstalledVersionNotInKymaChannel
	}

	clierr := Enable(ctx, client, repo, moduleName, channelAssignedToModuleVersion, enableDefaultCr(policy), []unstructured.Unstructured{}...)
	if clierr != nil {
		return fmt.Errorf("failed to manage module: %v", clierr)
	}

	return nil
}

func enableDefaultCr(policy string) bool {
	return policy == kyma.CustomResourcePolicyCreateAndDelete
}

func findInstalledModuleTemplate(ctx context.Context, client kube.Client, repo repo.ModuleTemplatesRepository, moduleName string) (*kyma.ModuleTemplate, error) {
	coreModuleTemplates, err := repo.Core(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get core modules: %n", err)
	}

	for _, coreModuleTemplate := range coreModuleTemplates {
		if coreModuleTemplate.Spec.ModuleName != moduleName {
			continue
		}

		if alreadyExistsInKymaCr(ctx, client, coreModuleTemplate.Spec.ModuleName) {
			continue
		}

		installedManager, err := repo.InstalledManager(ctx, coreModuleTemplate)
		if err != nil {
			fmt.Printf("failed to get installed manager: %v\n", err)
			continue
		}

		if installedManager == nil {
			continue
		}

		installedManagerVersion, err := getManagerVersion(installedManager)
		if err != nil {
			fmt.Printf("failed to determine installed manager version: %v\n", err)
			continue
		}

		if installedManagerVersion == coreModuleTemplate.Spec.Version {
			return &coreModuleTemplate, nil
		}
	}

	return nil, fmt.Errorf("failed to find installed module")
}

func GetAvailableChannelsAndVersions(ctx context.Context, client kube.Client, repo repo.ModuleTemplatesRepository, moduleName string) (map[string]string, error) {
	coreModuleTemplates, err := repo.Core(ctx)
	if err != nil {
		return nil, err
	}

	moduleReleaseMetas, err := client.Kyma().ListModuleReleaseMeta(ctx)
	if err != nil {
		return nil, err
	}

	channelsAndVersions := make(map[string]string)

	for _, coreModuleTemplate := range coreModuleTemplates {
		if coreModuleTemplate.Spec.ModuleName != moduleName {
			continue
		}

		assignedChannel := getAssignedChannel(*moduleReleaseMetas, moduleName, coreModuleTemplate.Spec.Version)

		if assignedChannel != "" {
			channelsAndVersions[assignedChannel] = coreModuleTemplate.Spec.Version
		}
	}

	return channelsAndVersions, nil
}

func alreadyExistsInKymaCr(ctx context.Context, client kube.Client, moduleName string) bool {
	defaultKyma, err := client.Kyma().GetDefaultKyma(ctx)
	if err != nil {
		fmt.Printf("failed to get kyma cr %v", err)
	}

	for _, m := range defaultKyma.Spec.Modules {
		if moduleName == m.Name {
			fmt.Printf("module %s already exists in kyma cr\n", moduleName)
			return true
		}
	}

	return false
}
