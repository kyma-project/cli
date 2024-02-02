package create_module_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/cli/tests/e2e"
)

const (
	ociRepositoryEnvVar         = "OCI_REPOSITORY_URL"
	moduleTemplateVersionEnvVar = "MODULE_TEMPLATE_VERSION"
)

func Test_SameVersion_ModuleCreation(t *testing.T) {
	path := "../" + "../../../template-operator"
	configFilePath := "../" + "../../../template-operator/module-config.yaml"
	secScannerConfigFile := "../" + "../../../template-operator/sec-scanners-config.yaml"
	changedSecScannerConfigFile := "../" + "../../../template-operator/sec-scanners-config-changed.yaml"
	version := os.Getenv(moduleTemplateVersionEnvVar)
	registry := os.Getenv(ociRepositoryEnvVar)

	t.Run("Create same version module with module-archive-version-overwrite flag", func(t *testing.T) {
		err := e2e.CreateModuleCommand(true, path, registry, configFilePath, version, secScannerConfigFile)
		moveDir(".bak1")
		assert.Nil(t, err)
	})

	t.Run("Create same version module and same content without module-archive-version-overwrite flag",
		func(t *testing.T) {
			err := e2e.CreateModuleCommand(false, path, registry, configFilePath, version, secScannerConfigFile)
			moveDir(".bak2")
			assert.Nil(t, err)
		})

	t.Run("Create same version module, but different content without module-archive-version-overwrite flag",
		func(t *testing.T) {
			err := e2e.CreateModuleCommand(false, path, registry, configFilePath, version, changedSecScannerConfigFile)
			moveDir(".bak3")
			assert.Equal(t, e2e.ErrCreateModuleFailedWithSameVersion, err)
		})
}

func moveDir(suffix string) {
	oldName := "mod"
	fi, err := os.Stat(oldName)
	if err != nil {
		fmt.Printf("error during os.Stat(\"mod\"): %s\n", err.Error())
		return
	}
	if fi.IsDir() {
		newName := oldName + suffix
		err = os.Rename(oldName, newName)
		if err != nil {
			fmt.Printf("error during os.Rename(oldName,newName): %s\n", err.Error())
		}
	} else {
		fmt.Printf("%s is not a dir!\n", oldName)
	}
}
