package modules

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	modulesfake "github.com/kyma-project/cli.v3/internal/modules/fake"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
)

func TestGetRunningResourcesOfCommunityModule(t *testing.T) {
	t.Run("fails to list installed community modules with provided name", func(t *testing.T) {
		ctx := context.Background()
		fakeModuleTemplatesRepo := modulesfake.ModuleTemplatesRepo{
			CommunityInstalledByNameErr: errors.New("CommunityInstallByNameError"),
		}

		_, err := GetRunningResourcesOfCommunityModule(ctx, &fakeModuleTemplatesRepo, "module")

		expectedCliErr := clierror.Wrap(
			errors.New("failed to retrieve a list of installed community modules: CommunityInstallByNameError"),
			clierror.New("failed to retrieve the module module"),
		)

		require.NotNil(t, err)
		require.Equal(t, expectedCliErr, err)
	})

	t.Run("fails to find installed module", func(t *testing.T) {
		ctx := context.Background()
		fakeModuleTemplatesRepo := modulesfake.ModuleTemplatesRepo{
			ReturnCommunityInstalledByName: []kyma.ModuleTemplate{},
		}

		_, err := GetRunningResourcesOfCommunityModule(ctx, &fakeModuleTemplatesRepo, "test")

		expectedCliErr := clierror.Wrap(
			errors.New("failed to find any version of the module test"),
			clierror.New("failed to retrieve the module test"),
		)

		require.NotNil(t, err)
		require.Equal(t, expectedCliErr, err)
	})

	t.Run("fails to determine module when multiple versions installed", func(t *testing.T) {
		ctx := context.Background()
		fakeModuleTemplatesRepo := modulesfake.ModuleTemplatesRepo{
			ReturnCommunityInstalledByName: []kyma.ModuleTemplate{{}, {}},
		}

		_, err := GetRunningResourcesOfCommunityModule(ctx, &fakeModuleTemplatesRepo, "test")

		expectedCliErr := clierror.Wrap(
			errors.New("failed to determine module version for test"),
			clierror.New("failed to retrieve the module test"),
		)

		require.NotNil(t, err)
		require.Equal(t, expectedCliErr, err)
	})

	t.Run("fails to retrieve running resources", func(t *testing.T) {
		ctx := context.Background()
		fakeModuleTemplatesRepo := modulesfake.ModuleTemplatesRepo{
			ReturnCommunityInstalledByName:        []kyma.ModuleTemplate{{}},
			RunningAssociatedResourcesOfModuleErr: errors.New("RunningAssociatedResourcesOfModuleError"),
		}

		_, err := GetRunningResourcesOfCommunityModule(ctx, &fakeModuleTemplatesRepo, "test")

		expectedCliErr := clierror.Wrap(
			errors.New("RunningAssociatedResourcesOfModuleError"),
			clierror.New("failed to retrieve running resources of the test module"),
		)

		require.NotNil(t, err)
		require.Equal(t, expectedCliErr, err)
	})

	t.Run("successfully returns a list of running resources", func(t *testing.T) {
		ctx := context.Background()
		fakeModuleTemplatesRepo := modulesfake.ModuleTemplatesRepo{
			ReturnCommunityInstalledByName: []kyma.ModuleTemplate{{}},
			ReturnRunningAssociatedResourcesOfModule: []unstructured.Unstructured{
				{
					Object: map[string]any{
						"metadata": map[string]any{
							"name": "resource-1",
						},
						"kind": "CustomResource",
					},
				},
				{
					Object: map[string]any{
						"metadata": map[string]any{
							"name": "resource-2",
						},
						"kind": "CustomResource",
					},
				},
			},
		}

		res, err := GetRunningResourcesOfCommunityModule(ctx, &fakeModuleTemplatesRepo, "test")
		require.Nil(t, err)
		require.NotNil(t, res)
		require.Equal(t, []string{"resource-1 (CustomResource)", "resource-2 (CustomResource)"}, res)
	})
}

func TestDisableCommunity(t *testing.T) {
	t.Run("fails to list installed community modules", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		ctx := context.Background()
		fakeModuleTemplatesRepo := modulesfake.ModuleTemplatesRepo{
			CommunityInstalledByNameErr: errors.New("CommunityInstalledErr"),
		}

		err := uninstall(out.NewToWriter(buffer), ctx, &fakeModuleTemplatesRepo, "test")

		expectedCliErr := clierror.Wrap(
			errors.New("failed to retrieve a list of installed community modules: CommunityInstalledErr"),
			clierror.New("failed to retrieve the module test"),
		)

		require.NotNil(t, err)
		require.Equal(t, expectedCliErr, err)
		require.Equal(t, "removing test community module from the target Kyma environment\n", buffer.String())
	})

	t.Run("fails to find any version for the module", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		ctx := context.Background()
		fakeModuleTemplatesRepo := modulesfake.ModuleTemplatesRepo{
			ReturnCommunityInstalledByName: []kyma.ModuleTemplate{},
		}

		err := uninstall(out.NewToWriter(buffer), ctx, &fakeModuleTemplatesRepo, "test")

		expectedCliErr := clierror.Wrap(
			errors.New("failed to find any version of the module test"),
			clierror.New("failed to retrieve the module test"),
		)

		require.NotNil(t, err)
		require.Equal(t, expectedCliErr, err)
		require.Equal(t, "removing test community module from the target Kyma environment\n", buffer.String())
	})

	t.Run("fails to determine version to remove", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		ctx := context.Background()
		fakeModuleTemplatesRepo := modulesfake.ModuleTemplatesRepo{
			ReturnCommunityInstalledByName: []kyma.ModuleTemplate{
				{
					Spec: kyma.ModuleTemplateSpec{
						ModuleName: "test",
						Version:    "1.0.1",
					},
				},
				{
					Spec: kyma.ModuleTemplateSpec{
						ModuleName: "test",
						Version:    "1.0.2",
					},
				},
			},
		}

		err := uninstall(out.NewToWriter(buffer), ctx, &fakeModuleTemplatesRepo, "test")

		expectedCliErr := clierror.Wrap(
			errors.New("failed to determine module version for test"),
			clierror.New("failed to retrieve the module test"),
		)

		require.NotNil(t, err)
		require.Equal(t, expectedCliErr, err)
		require.Equal(t, "removing test community module from the target Kyma environment\n", buffer.String())
	})

	t.Run("fails to get modules resources", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		ctx := context.Background()
		fakeModuleTemplatesRepo := modulesfake.ModuleTemplatesRepo{
			ReturnCommunityInstalledByName: []kyma.ModuleTemplate{
				{
					Spec: kyma.ModuleTemplateSpec{
						ModuleName: "test",
						Version:    "1.0.1",
					},
				},
			},
			ResourcesErr: errors.New("ResourcesError"),
		}

		err := uninstall(out.NewToWriter(buffer), ctx, &fakeModuleTemplatesRepo, "test")

		expectedCliErr := clierror.Wrap(
			errors.New("ResourcesError"),
			clierror.New("failed to get resources for module test"),
		)

		require.NotNil(t, err)
		require.Equal(t, expectedCliErr, err)
	})

	t.Run("fails to remove resources", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		ctx := context.Background()
		fakeModuleTemplatesRepo := modulesfake.ModuleTemplatesRepo{
			ReturnCommunityInstalledByName: []kyma.ModuleTemplate{
				{
					Spec: kyma.ModuleTemplateSpec{
						ModuleName: "test",
						Version:    "1.0.1",
					},
				},
			},
			ReturnResources: []map[string]any{
				{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"metadata": map[string]any{
						"name": "test-namespace",
					},
				},
				{
					"apiVersion": "v1",
					"kind":       "Secret",
					"metadata": map[string]any{
						"name":      "test-secret",
						"namespace": "test-namespace",
					},
					"data": map[string]any{
						"key": "value",
					},
				},
			},
			DeleteResourceReturnWatcherErr: errors.New("DeleteResourceReturnWatcherError"),
		}

		err := uninstall(out.NewToWriter(buffer), ctx, &fakeModuleTemplatesRepo, "test")

		require.Nil(t, err)
		require.Equal(t, "removing test community module from the target Kyma environment\nfailed to delete resource test-secret (Secret): DeleteResourceReturnWatcherError\nfailed to delete resource test-namespace (Namespace): DeleteResourceReturnWatcherError\nsome errors occured during the test community module removal\n", buffer.String())
	})

	t.Run("timeouts waiting for the resource removal", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		ctx, cancel := context.WithCancel(context.Background())
		fakeWatcher := watch.NewFake()
		fakeModuleTemplatesRepo := modulesfake.ModuleTemplatesRepo{
			ReturnCommunityInstalledByName: []kyma.ModuleTemplate{
				{
					Spec: kyma.ModuleTemplateSpec{
						ModuleName: "test",
						Version:    "1.0.1",
					},
				},
			},
			ReturnResources: []map[string]any{
				{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"metadata": map[string]any{
						"name": "test-namespace",
					},
				},
				{
					"apiVersion": "v1",
					"kind":       "Secret",
					"metadata": map[string]any{
						"name":      "test-secret",
						"namespace": "test-namespace",
					},
					"data": map[string]any{
						"key": "value",
					},
				},
			},
			ReturnDeleteResourceReturnWatcher: fakeWatcher,
		}

		go func() {
			fakeWatcher.Add(nil)
			// emit any event and then cancel ctx to simluate timeout
			cancel()
		}()

		err := uninstall(out.NewToWriter(buffer), ctx, &fakeModuleTemplatesRepo, "test")

		expectedCliErr := clierror.Wrap(
			errors.New("context canceled"),
			clierror.New("context timeout"),
		)

		require.NotNil(t, err)
		require.Equal(t, expectedCliErr, err)
		require.Equal(t, buffer.String(), "removing test community module from the target Kyma environment\nwaiting for resource deletion: test-secret (Secret)\n")
	})

	t.Run("successfully removes the module", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		ctx := context.Background()
		fakeWatcher := watch.NewFake()
		fakeModuleTemplatesRepo := modulesfake.ModuleTemplatesRepo{
			ReturnCommunityInstalledByName: []kyma.ModuleTemplate{
				{
					Spec: kyma.ModuleTemplateSpec{
						ModuleName: "test",
						Version:    "1.0.1",
					},
				},
			},
			ReturnResources: []map[string]any{
				{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"metadata": map[string]any{
						"name": "test-namespace",
					},
				},
				{
					"apiVersion": "v1",
					"kind":       "Secret",
					"metadata": map[string]any{
						"name":      "test-secret",
						"namespace": "test-namespace",
					},
					"data": map[string]any{
						"key": "value",
					},
				},
			},
			ReturnDeleteResourceReturnWatcher: fakeWatcher,
		}

		go func() {
			// successfully deletes both reources
			fakeWatcher.Delete(nil)
			fakeWatcher.Delete(nil)
		}()

		err := uninstall(out.NewToWriter(buffer), ctx, &fakeModuleTemplatesRepo, "test")

		require.Nil(t, err)
		require.Equal(t, buffer.String(), "removing test community module from the target Kyma environment\nwaiting for resource deletion: test-secret (Secret)\nwaiting for resource deletion: test-namespace (Namespace)\ntest community module successfully removed\n")
	})
}
