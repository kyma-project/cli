package create_module_test

import (
	"testing"

	"github.com/kyma-project/cli/tests/e2e"
	"github.com/stretchr/testify/assert"
)

func Test_SameVersion_ModuleCreation(t *testing.T) {
	path := "../../../template-operator"
	registry := "http://k3d-oci.localhost:5001"
	configFilePath := "../../../template-operator/module-config.yaml"
	version := "v1.0.0"

	t.Run("Create same version module with module-archive-version-overwrite flag", func(t *testing.T) {
		err := e2e.CreateModuleCommand(true, path, registry, configFilePath, version)
		assert.Nil(t, err)
	})

	t.Run("Create same version module without module-archive-version-overwrite flag", func(t *testing.T) {
		err := e2e.CreateModuleCommand(false, path, registry, configFilePath, version)
		assert.Equal(t, e2e.ErrCreateModuleFailedWithSameVersion, err)
	})
}
