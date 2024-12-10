package modules

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

func TestDisable(t *testing.T) {
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

		err := disable(buffer, context.Background(), &fakeKubeClient, "keda")
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

		err := disable(buffer, context.Background(), &fakeKubeClient, "keda")
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
		fakeRootlessDynamicClient := fake.RootlessDynamicClient{
			ReturnGetErr: &apierrors.StatusError{
				ErrStatus: metav1.Status{
					Status:  metav1.StatusFailure,
					Reason:  metav1.StatusReasonNotFound,
					Message: "not found",
				},
			},
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

		err := disable(buffer, context.Background(), &fakeKubeClient, "keda")
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

		err := disable(buffer, context.Background(), &fakeKubeClient, "keda")
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

		err := disable(buffer, context.Background(), &fakeKubeClient, "keda")
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

		err := disable(buffer, context.Background(), &fakeKubeClient, "keda")
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

		err := disable(buffer, context.Background(), &fakeKubeClient, "keda")
		require.Equal(t, expectedCliErr, err)
		require.Empty(t, fakeKymaClient.DisabledModules)
		require.Equal(t, "removing kyma-system/default CR\n", buffer.String())
	})
}

func Test_isObjDeleted(t *testing.T) {
	t.Run("return nil when obj is not found", func(t *testing.T) {
		fakeClient := fake.RootlessDynamicClient{
			ReturnGetErr: &apierrors.StatusError{
				ErrStatus: metav1.Status{
					Status:  metav1.StatusFailure,
					Reason:  metav1.StatusReasonNotFound,
					Message: "not found",
				},
			},
		}

		err := isObjDeleted(context.Background(), &fakeClient, testKedaCR.DeepCopy())
		require.NoError(t, err)
	})

	t.Run("return error because of unexpected client error", func(t *testing.T) {
		fakeClient := fake.RootlessDynamicClient{
			ReturnGetErr: errors.New("test error"),
		}

		err := isObjDeleted(context.Background(), &fakeClient, testKedaCR.DeepCopy())
		require.ErrorContains(t, err, "test error")
	})

	t.Run("return error because of client returne object", func(t *testing.T) {
		fakeClient := fake.RootlessDynamicClient{
			ReturnGetErr: nil,
			ReturnGetObj: testKedaCR,
		}

		err := isObjDeleted(context.Background(), &fakeClient, testKedaCR.DeepCopy())
		require.ErrorContains(t, err, "kyma-system/default exists on the cluster")
	})
}
