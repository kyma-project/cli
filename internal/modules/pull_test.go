package modules_test

import (
	"context"
	"errors"
	"testing"

	kubeFake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modules"
	modulesFake "github.com/kyma-project/cli.v3/internal/modules/fake"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetModuleTemplateFromRemote(t *testing.T) {
	ctx := context.Background()

	t.Run("should return module template when found", func(t *testing.T) {
		// Given
		moduleName := "test-module"
		version := "v1.0.0"

		expectedModule := kyma.ModuleTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-module-template",
			},
			Spec: kyma.ModuleTemplateSpec{
				ModuleName: moduleName,
				Version:    version,
			},
		}

		fakeRepo := &modulesFake.ModuleTemplatesRepo{
			ReturnExternalCommunityByNameAndVersion: []kyma.ModuleTemplate{expectedModule},
		}

		// When
		result, err := modules.GetModuleTemplateFromRemote(ctx, fakeRepo, moduleName, version)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, moduleName, result.Spec.ModuleName)
		assert.Equal(t, version, result.Spec.Version)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Given
		moduleName := "test-module"
		version := "v1.0.0"
		expectedError := errors.New("repository error")

		fakeRepo := &modulesFake.ModuleTemplatesRepo{
			ExternalCommunityByNameAndVersionErr: expectedError,
		}

		// When
		result, err := modules.GetModuleTemplateFromRemote(ctx, fakeRepo, moduleName, version)

		// Then
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get module test-module")
	})

	t.Run("should return error when module not found", func(t *testing.T) {
		// Given
		moduleName := "test-module"
		version := "v1.0.0"

		fakeRepo := &modulesFake.ModuleTemplatesRepo{
			ReturnExternalCommunityByNameAndVersion: []kyma.ModuleTemplate{},
		}

		// When
		result, err := modules.GetModuleTemplateFromRemote(ctx, fakeRepo, moduleName, version)

		// Then
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "module not found in the catalog")
	})

	t.Run("should handle empty module list", func(t *testing.T) {
		// Given
		moduleName := "test-module"
		version := "v1.0.0"

		fakeRepo := &modulesFake.ModuleTemplatesRepo{
			ReturnExternalCommunityByNameAndVersion: []kyma.ModuleTemplate{},
		}

		// When
		result, err := modules.GetModuleTemplateFromRemote(ctx, fakeRepo, moduleName, version)

		// Then
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "module not found in the catalog")
	})
}

func TestPersistModuleTemplateInNamespace(t *testing.T) {
	ctx := context.Background()

	t.Run("should successfully persist module template", func(t *testing.T) {
		// Given
		fakeRootlessDynamic := &kubeFake.RootlessDynamicClient{}
		fakeKubeClient := &kubeFake.KubeClient{
			TestRootlessDynamicInterface: fakeRootlessDynamic,
		}

		moduleTemplate := &kyma.ModuleTemplate{
			TypeMeta: metav1.TypeMeta{
				Kind: "ModuleTemplate",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-module-template",
			},
			Spec: kyma.ModuleTemplateSpec{
				ModuleName: "test-module",
				Version:    "v1.0.0",
			},
		}

		namespace := "custom-namespace"

		// When
		err := modules.PersistModuleTemplateInNamespace(ctx, fakeKubeClient, moduleTemplate, namespace)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, namespace, moduleTemplate.Namespace)
		assert.Len(t, fakeRootlessDynamic.ApplyObjs, 1)

		// Verify the applied object
		appliedObj := fakeRootlessDynamic.ApplyObjs[0]
		assert.Equal(t, "test-module-template", appliedObj.GetName())
		assert.Equal(t, namespace, appliedObj.GetNamespace())
		assert.Equal(t, "ModuleTemplate", appliedObj.GetKind())
	})

	t.Run("should return error when apply fails", func(t *testing.T) {
		// Given
		expectedError := errors.New("apply failed")
		fakeRootlessDynamic := &kubeFake.RootlessDynamicClient{
			ReturnErr: expectedError,
		}
		fakeKubeClient := &kubeFake.KubeClient{
			TestRootlessDynamicInterface: fakeRootlessDynamic,
		}

		moduleTemplate := &kyma.ModuleTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-module-template",
			},
			Spec: kyma.ModuleTemplateSpec{
				ModuleName: "test-module",
				Version:    "v1.0.0",
			},
		}

		namespace := "custom-namespace"

		// When
		err := modules.PersistModuleTemplateInNamespace(ctx, fakeKubeClient, moduleTemplate, namespace)

		// Then
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Equal(t, namespace, moduleTemplate.Namespace)
		assert.Len(t, fakeRootlessDynamic.ApplyObjs, 1)
	})

	t.Run("should update namespace on module template", func(t *testing.T) {
		// Given
		fakeRootlessDynamic := &kubeFake.RootlessDynamicClient{}
		fakeKubeClient := &kubeFake.KubeClient{
			TestRootlessDynamicInterface: fakeRootlessDynamic,
		}

		moduleTemplate := &kyma.ModuleTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-module-template",
				Namespace: "original-namespace", // Original namespace
			},
			Spec: kyma.ModuleTemplateSpec{
				ModuleName: "test-module",
				Version:    "v1.0.0",
			},
		}

		targetNamespace := "target-namespace"

		// When
		err := modules.PersistModuleTemplateInNamespace(ctx, fakeKubeClient, moduleTemplate, targetNamespace)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, targetNamespace, moduleTemplate.Namespace) // Should be updated
		assert.Len(t, fakeRootlessDynamic.ApplyObjs, 1)

		// Verify the applied object has the correct namespace
		appliedObj := fakeRootlessDynamic.ApplyObjs[0]
		assert.Equal(t, targetNamespace, appliedObj.GetNamespace())
	})

	t.Run("should handle empty module template", func(t *testing.T) {
		// Given
		fakeRootlessDynamic := &kubeFake.RootlessDynamicClient{}
		fakeKubeClient := &kubeFake.KubeClient{
			TestRootlessDynamicInterface: fakeRootlessDynamic,
		}

		moduleTemplate := &kyma.ModuleTemplate{}
		namespace := "test-namespace"

		// When
		err := modules.PersistModuleTemplateInNamespace(ctx, fakeKubeClient, moduleTemplate, namespace)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, namespace, moduleTemplate.Namespace)
		assert.Len(t, fakeRootlessDynamic.ApplyObjs, 1)

		// Verify the applied object
		appliedObj := fakeRootlessDynamic.ApplyObjs[0]
		assert.Equal(t, namespace, appliedObj.GetNamespace())
	})
}
