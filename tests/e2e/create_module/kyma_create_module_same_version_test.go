package create_module_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/cli/tests/e2e"
)

func Test_SameVersion_ModuleCreation(t *testing.T) {
	path := "../../../template-operator"
	registry := "http://k3d-oci.localhost:5001"
	configFilePath := "../../../template-operator/module-config.yaml"
	secScannerConfigFile := "../../../template-operator/sec-scanners-config.yaml"
	changedsecScannerConfigFile := "../../../template-operator/sec-scanners-config-changed.yaml"
	version := "1.0.0"

	t.Run("Create same version module with module-archive-version-overwrite flag", func(t *testing.T) {
		err := e2e.CreateModuleCommand(true, path, registry, configFilePath, version, secScannerConfigFile)
		assert.Nil(t, err)
	})

	t.Run("Create same version module and same content without module-archive-version-overwrite flag",
		func(t *testing.T) {
			err := e2e.CreateModuleCommand(false, path, registry, configFilePath, version, secScannerConfigFile)
			assert.Nil(t, err)
		})

	t.Run("Create same version module but different content without module-archive-version-overwrite flag",
		func(t *testing.T) {
			err := e2e.CreateModuleCommand(false, path, registry, configFilePath, version, changedsecScannerConfigFile)
			assert.Equal(t, e2e.ErrCreateModuleFailedWithSameVersion, err)
		})
}
