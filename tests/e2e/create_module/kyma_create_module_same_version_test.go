package create_module_test

import (
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
	path := "../../../template-operator"
	configFilePath := "../../../template-operator/module-config.yaml"
	secScannerConfigFile := "../../../template-operator/sec-scanners-config.yaml"
	changedSecScannerConfigFile := "../../../template-operator/sec-scanners-config-changed.yaml"
	version := os.Getenv(moduleTemplateVersionEnvVar)
	registry := os.Getenv(ociRepositoryEnvVar)

	t.Run("Create same version module with module-archive-version-overwrite flag", func(t *testing.T) {
		err := e2e.CreateModuleCommand(true, path, registry, configFilePath, version, secScannerConfigFile)
		assert.Nil(t, err)
	})

	t.Run("Create same version module and same content without module-archive-version-overwrite flag",
		func(t *testing.T) {
			err := e2e.CreateModuleCommand(false, path, registry, configFilePath, version, secScannerConfigFile)
			assert.Nil(t, err)
		})

	t.Run("Create same version module, but different content without module-archive-version-overwrite flag",
		func(t *testing.T) {
			err := e2e.CreateModuleCommand(false, path, registry, configFilePath, version, changedSecScannerConfigFile)
			assert.Equal(t, e2e.ErrCreateModuleFailedWithSameVersion, err)
		})
}
