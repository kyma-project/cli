package modules

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	modulesfake "github.com/kyma-project/cli.v3/internal/modules/fake"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
)

var (
	testKedaTemplate = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleTemplate",
			"metadata": map[string]interface{}{
				"name":      "keda",
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"moduleName": "keda",
				"version":    "0.0.1",
				"data":       testKedaCR.Object,
			},
		},
	}
)

func TestDisableCore(t *testing.T) {
	t.Run("disable module with CreateAndDelete policy", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		fakeKymaClient := fake.KymaClient{
			ReturnErr: nil,
			ReturnModuleInfo: kyma.KymaModuleInfo{
				Spec: kyma.Module{
					CustomResourcePolicy: kyma.CustomResourcePolicyCreateAndDelete,
				},
			},
		}
		fakeKubeClient := fake.KubeClient{
			TestKymaInterface: &fakeKymaClient,
		}

		err := disableCore(buffer, context.Background(), &fakeKubeClient, "keda")
		require.Nil(t, err)
		require.Equal(t, []string{"keda"}, fakeKymaClient.DisabledModules)
		require.Equal(t, "removing keda module from the Kyma CR\nkeda module disabled\n", buffer.String())
	})

	t.Run("disable module with Ignore policy for module with no CR", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		fakeKymaClient := fake.KymaClient{
			ReturnErr: nil,
			ReturnModuleInfo: kyma.KymaModuleInfo{
				Spec: kyma.Module{
					CustomResourcePolicy: kyma.CustomResourcePolicyIgnore,
				},
			},
			ReturnModuleTemplate: kyma.ModuleTemplate{
				Spec: kyma.ModuleTemplateSpec{
					Data: unstructured.Unstructured{
						// empty
					},
				},
			},
		}
		fakeKubeClient := fake.KubeClient{
			TestKymaInterface: &fakeKymaClient,
		}

		err := disableCore(buffer, context.Background(), &fakeKubeClient, "keda")
		require.Nil(t, err)
		require.Equal(t, []string{"keda"}, fakeKymaClient.DisabledModules)
		require.Equal(t, "removing keda module from the Kyma CR\nkeda module disabled\n", buffer.String())
	})

	t.Run("disable module with Ignore policy for module", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		fakeKymaClient := fake.KymaClient{
			ReturnErr: nil,
			ReturnModuleInfo: kyma.KymaModuleInfo{
				Spec: kyma.Module{
					CustomResourcePolicy: kyma.CustomResourcePolicyIgnore,
				},
			},
			ReturnModuleTemplate: kyma.ModuleTemplate{
				Spec: kyma.ModuleTemplateSpec{
					Data: testKedaTemplate,
				},
			},
		}
		fakeWatcher := watch.NewFakeWithChanSize(1, false)
		fakeRootlessDynamicClient := fake.RootlessDynamicClient{
			ReturnWatcher: fakeWatcher,
			ReturnListObjs: &unstructured.UnstructuredList{
				Items: []unstructured.Unstructured{
					testKedaCR,
				},
			},
		}
		fakeKubeClient := fake.KubeClient{
			TestKymaInterface:            &fakeKymaClient,
			TestRootlessDynamicInterface: &fakeRootlessDynamicClient,
		}

		fakeWatcher.Delete(nil)

		err := disableCore(buffer, context.Background(), &fakeKubeClient, "keda")
		require.Nil(t, err)
		require.Equal(t, []string{"keda"}, fakeKymaClient.DisabledModules)
		require.Equal(t, "removing kyma-system/default CR\nwaiting for kyma-system/default CR to be removed\nremoving keda module from the Kyma CR\nkeda module disabled\n", buffer.String())
	})

	t.Run("failed to disable module", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		fakeKymaClient := fake.KymaClient{
			ReturnDisableModuleErr: errors.New("test error"),
			ReturnModuleInfo: kyma.KymaModuleInfo{
				Spec: kyma.Module{
					CustomResourcePolicy: kyma.CustomResourcePolicyCreateAndDelete,
				},
			},
		}
		fakeKubeClient := fake.KubeClient{
			TestKymaInterface: &fakeKymaClient,
		}

		expectedCliErr := clierror.Wrap(
			errors.New("test error"),
			clierror.New("failed to disable module"),
		)

		err := disableCore(buffer, context.Background(), &fakeKubeClient, "keda")
		require.Equal(t, expectedCliErr, err)
		require.Equal(t, "removing keda module from the Kyma CR\n", buffer.String())
	})

	t.Run("failed to get module info", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		fakeKymaClient := fake.KymaClient{
			ReturnGetModuleInfoErr: errors.New("test error"),
		}
		fakeKubeClient := fake.KubeClient{
			TestKymaInterface: &fakeKymaClient,
		}

		expectedCliErr := clierror.Wrap(
			errors.New("test error"),
			clierror.New("failed to get module info from the Kyma CR"),
		)

		err := disableCore(buffer, context.Background(), &fakeKubeClient, "keda")
		require.Equal(t, expectedCliErr, err)
		require.Empty(t, fakeKymaClient.DisabledModules)
		require.Empty(t, buffer.String())
	})

	t.Run("failed to get ModuleTemplate for module", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		fakeKymaClient := fake.KymaClient{
			ReturnGetModuleTemplateErr: errors.New("test error"),
			ReturnModuleInfo: kyma.KymaModuleInfo{
				Spec: kyma.Module{
					CustomResourcePolicy: kyma.CustomResourcePolicyIgnore,
				},
			},
		}
		fakeKubeClient := fake.KubeClient{
			TestKymaInterface: &fakeKymaClient,
		}

		expectedCliErr := clierror.Wrap(
			errors.New("test error"),
			clierror.New("failed to get ModuleTemplate CR for module"),
		)

		err := disableCore(buffer, context.Background(), &fakeKubeClient, "keda")
		require.Equal(t, expectedCliErr, err)
		require.Empty(t, fakeKymaClient.DisabledModules)
		require.Empty(t, buffer.String())
	})

	t.Run("failed to remove module cr", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		fakeKymaClient := fake.KymaClient{
			ReturnErr: nil,
			ReturnModuleInfo: kyma.KymaModuleInfo{
				Spec: kyma.Module{
					CustomResourcePolicy: kyma.CustomResourcePolicyIgnore,
				},
			},
			ReturnModuleTemplate: kyma.ModuleTemplate{
				Spec: kyma.ModuleTemplateSpec{
					Data: testKedaTemplate,
				},
			},
		}
		fakeRootlessDynamicClient := fake.RootlessDynamicClient{
			ReturnWatcher:   watch.NewFake(),
			ReturnRemoveErr: errors.New("test error"),
			ReturnListObjs: &unstructured.UnstructuredList{
				Items: []unstructured.Unstructured{
					testKedaCR,
				},
			},
		}
		fakeKubeClient := fake.KubeClient{
			TestKymaInterface:            &fakeKymaClient,
			TestRootlessDynamicInterface: &fakeRootlessDynamicClient,
		}

		expectedCliErr := clierror.Wrap(
			errors.New("test error"),
			clierror.New("failed to remove kyma-system/default cr"),
		)

		err := disableCore(buffer, context.Background(), &fakeKubeClient, "keda")
		require.Equal(t, expectedCliErr, err)
		require.Empty(t, fakeKymaClient.DisabledModules)
		require.Equal(t, "removing kyma-system/default CR\n", buffer.String())
	})

	t.Run("failed to watch resource", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		fakeKymaClient := fake.KymaClient{
			ReturnErr: nil,
			ReturnModuleInfo: kyma.KymaModuleInfo{
				Spec: kyma.Module{
					CustomResourcePolicy: kyma.CustomResourcePolicyIgnore,
				},
			},
			ReturnModuleTemplate: kyma.ModuleTemplate{
				Spec: kyma.ModuleTemplateSpec{
					Data: testKedaTemplate,
				},
			},
		}
		fakeRootlessDynamicClient := fake.RootlessDynamicClient{
			ReturnWatcher:  nil,
			ReturnWatchErr: errors.New("test error"),
			ReturnListObjs: &unstructured.UnstructuredList{
				Items: []unstructured.Unstructured{
					testKedaCR,
				},
			},
		}
		fakeKubeClient := fake.KubeClient{
			TestKymaInterface:            &fakeKymaClient,
			TestRootlessDynamicInterface: &fakeRootlessDynamicClient,
		}

		expectedCliErr := clierror.Wrap(
			errors.New("test error"),
			clierror.New("failed to watch resource kyma-system/default"),
		)

		err := disableCore(buffer, context.Background(), &fakeKubeClient, "keda")
		require.Equal(t, expectedCliErr, err)
		require.Empty(t, fakeKymaClient.DisabledModules)
	})

	t.Run("wait for resource ctx done error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		buffer := bytes.NewBuffer([]byte{})
		fakeKymaClient := fake.KymaClient{
			ReturnErr: nil,
			ReturnModuleInfo: kyma.KymaModuleInfo{
				Spec: kyma.Module{
					CustomResourcePolicy: kyma.CustomResourcePolicyIgnore,
				},
			},
			ReturnModuleTemplate: kyma.ModuleTemplate{
				Spec: kyma.ModuleTemplateSpec{
					Data: testKedaTemplate,
				},
			},
		}
		fakeWatcher := watch.NewFake()
		fakeRootlessDynamicClient := fake.RootlessDynamicClient{
			ReturnWatcher: fakeWatcher,
			ReturnListObjs: &unstructured.UnstructuredList{
				Items: []unstructured.Unstructured{
					testKedaCR,
				},
			},
		}
		fakeKubeClient := fake.KubeClient{
			TestKymaInterface:            &fakeKymaClient,
			TestRootlessDynamicInterface: &fakeRootlessDynamicClient,
		}

		expectedCliErr := clierror.Wrap(
			errors.New("context canceled"),
			clierror.New("context timeout"),
		)

		go func() {
			fakeWatcher.Add(nil)
			// emit any event and then cancel ctx to simluate timeout
			cancel()
		}()

		err := disableCore(buffer, ctx, &fakeKubeClient, "keda")
		require.Equal(t, expectedCliErr, err)
		require.Equal(t, "removing kyma-system/default CR\nwaiting for kyma-system/default CR to be removed\n", buffer.String())
	})
}

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
					},
				},
				{
					Object: map[string]any{
						"metadata": map[string]any{
							"name": "resource-2",
						},
					},
				},
			},
		}

		res, err := GetRunningResourcesOfCommunityModule(ctx, &fakeModuleTemplatesRepo, "test")
		require.Nil(t, err)
		require.NotNil(t, res)
		require.Equal(t, []string{"resource-1", "resource-2"}, res)
	})
}

func TestDisableCommunity(t *testing.T) {
	t.Run("fails to list installed community modules", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		ctx := context.Background()
		fakeModuleTemplatesRepo := modulesfake.ModuleTemplatesRepo{
			CommunityInstalledByNameErr: errors.New("CommunityInstalledErr"),
		}

		err := disableCommunity(buffer, ctx, &fakeModuleTemplatesRepo, "test")

		expectedCliErr := clierror.Wrap(
			errors.New("failed to retrieve a list of installed community modules: CommunityInstalledErr"),
			clierror.New("failed to retrieve the module test"),
		)

		require.NotNil(t, err)
		require.Equal(t, expectedCliErr, err)
		require.Equal(t, "removing test community module from the cluster\n", buffer.String())
	})

	t.Run("fails to find any version for the module", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		ctx := context.Background()
		fakeModuleTemplatesRepo := modulesfake.ModuleTemplatesRepo{
			ReturnCommunityInstalledByName: []kyma.ModuleTemplate{},
		}

		err := disableCommunity(buffer, ctx, &fakeModuleTemplatesRepo, "test")

		expectedCliErr := clierror.Wrap(
			errors.New("failed to find any version of the module test"),
			clierror.New("failed to retrieve the module test"),
		)

		require.NotNil(t, err)
		require.Equal(t, expectedCliErr, err)
		require.Equal(t, "removing test community module from the cluster\n", buffer.String())
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

		err := disableCommunity(buffer, ctx, &fakeModuleTemplatesRepo, "test")

		expectedCliErr := clierror.Wrap(
			errors.New("failed to determine module version for test"),
			clierror.New("failed to retrieve the module test"),
		)

		require.NotNil(t, err)
		require.Equal(t, expectedCliErr, err)
		require.Equal(t, "removing test community module from the cluster\n", buffer.String())
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

		err := disableCommunity(buffer, ctx, &fakeModuleTemplatesRepo, "test")

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

		err := disableCommunity(buffer, ctx, &fakeModuleTemplatesRepo, "test")

		require.Nil(t, err)
		require.Equal(t, "removing test community module from the cluster\nfailed to delete resource test-secret (Secret): DeleteResourceReturnWatcherError\nfailed to delete resource test-namespace (Namespace): DeleteResourceReturnWatcherError\nsome errors occured during the test community module removal\n", buffer.String())
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

		err := disableCommunity(buffer, ctx, &fakeModuleTemplatesRepo, "test")

		expectedCliErr := clierror.Wrap(
			errors.New("context canceled"),
			clierror.New("context timeout"),
		)

		require.NotNil(t, err)
		require.Equal(t, expectedCliErr, err)
		require.Equal(t, buffer.String(), "removing test community module from the cluster\nwaiting for resource deletion: test-secret (Secret)\n")
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

		err := disableCommunity(buffer, ctx, &fakeModuleTemplatesRepo, "test")

		require.Nil(t, err)
		require.Equal(t, buffer.String(), "removing test community module from the cluster\nwaiting for resource deletion: test-secret (Secret)\nwaiting for resource deletion: test-namespace (Namespace)\ntest community module successfully removed\n")
	})
}
