package modules

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/out"
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

		err := disable(out.NewToWriter(buffer), context.Background(), &fakeKubeClient, "keda")
		require.Nil(t, err)
		require.Equal(t, []string{"keda"}, fakeKymaClient.DisabledModules)
		require.Equal(t, "removing the keda module from the target Kyma environment\nkeda module disabled\n", buffer.String())
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

		err := disable(out.NewToWriter(buffer), context.Background(), &fakeKubeClient, "keda")
		require.Nil(t, err)
		require.Equal(t, []string{"keda"}, fakeKymaClient.DisabledModules)
		require.Equal(t, "removing the keda module from the target Kyma environment\nkeda module disabled\n", buffer.String())
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

		err := disable(out.NewToWriter(buffer), context.Background(), &fakeKubeClient, "keda")
		require.Nil(t, err)
		require.Equal(t, []string{"keda"}, fakeKymaClient.DisabledModules)
		require.Equal(t, "removing kyma-system/default CR\nwaiting for kyma-system/default CR to be removed\nremoving keda module from the target Kyma environment\nkeda module disabled\n", buffer.String())
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
			clierror.New("failed to disable the module"),
		)

		err := disable(out.NewToWriter(buffer), context.Background(), &fakeKubeClient, "keda")
		require.Equal(t, expectedCliErr, err)
		require.Equal(t, "removing the keda module from the target Kyma environment\n", buffer.String())
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
			clierror.New("failed to get the module info from the target Kyma environment"),
		)

		err := disable(out.NewToWriter(buffer), context.Background(), &fakeKubeClient, "keda")
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

		err := disable(out.NewToWriter(buffer), context.Background(), &fakeKubeClient, "keda")
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

		err := disable(out.NewToWriter(buffer), context.Background(), &fakeKubeClient, "keda")
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

		err := disable(out.NewToWriter(buffer), context.Background(), &fakeKubeClient, "keda")
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

		err := disable(out.NewToWriter(buffer), ctx, &fakeKubeClient, "keda")
		require.Equal(t, expectedCliErr, err)
		require.Equal(t, "removing kyma-system/default CR\nwaiting for kyma-system/default CR to be removed\n", buffer.String())
	})
}
