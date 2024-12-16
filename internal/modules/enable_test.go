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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	testKedaCR = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "test/v1",
			"kind":       "Keda",
			"metadata": map[string]interface{}{
				"name":      "default",
				"namespace": "kyma-system",
			},
		},
	}
)

func TestEnable(t *testing.T) {
	t.Run("enable module", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		kymaClient := fake.KymaClient{
			ReturnErr: nil,
		}
		client := fake.KubeClient{
			TestKymaInterface: &kymaClient,
		}

		expectedEnabledModule := fake.FakeEnabledModule{
			Name:                 "keda",
			Channel:              "fast",
			CustomResourcePolicy: kyma.CustomResourcePolicyCreateAndDelete,
		}

		err := enable(buffer, context.Background(), &client, "keda", "fast", true)
		require.Nil(t, err)
		require.Equal(t, "adding keda module to the Kyma CR\nkeda module enabled\n", buffer.String())
		require.Equal(t, []fake.FakeEnabledModule{expectedEnabledModule}, kymaClient.EnabledModules)
	})

	t.Run("enable module and add custom cr", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		kymaClient := fake.KymaClient{
			ReturnErr: nil,
		}
		rootlessDynamicClient := fake.RootlessDynamicClient{}
		client := fake.KubeClient{
			TestKymaInterface:            &kymaClient,
			TestRootlessDynamicInterface: &rootlessDynamicClient,
		}

		expectedEnabledModule := fake.FakeEnabledModule{
			Name:                 "keda",
			Channel:              "fast",
			CustomResourcePolicy: kyma.CustomResourcePolicyIgnore,
		}

		err := enable(buffer, context.Background(), &client, "keda", "fast", false, testKedaCR)
		require.Nil(t, err)
		require.Equal(t, "adding keda module to the Kyma CR\nwaiting for module to be ready\napplying kyma-system/default cr\nkeda module enabled\n", buffer.String())
		require.Equal(t, []fake.FakeEnabledModule{expectedEnabledModule}, kymaClient.EnabledModules)
		require.Equal(t, []unstructured.Unstructured{testKedaCR}, rootlessDynamicClient.ApplyObjs)
	})

	t.Run("failed to enable module", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		kymaClient := fake.KymaClient{
			ReturnErr: errors.New("test error"),
		}
		client := fake.KubeClient{
			TestKymaInterface: &kymaClient,
		}

		expectedCliErr := clierror.Wrap(
			errors.New("test error"),
			clierror.New("failed to enable module"),
		)

		err := enable(buffer, context.Background(), &client, "keda", "fast", true)
		require.Equal(t, expectedCliErr, err)
		require.Equal(t, "adding keda module to the Kyma CR\n", buffer.String())
	})

	t.Run("failed to wait for module to be ready", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		kymaClient := fake.KymaClient{
			ReturnWaitForModuleErr: errors.New("test error"),
		}
		client := fake.KubeClient{
			TestKymaInterface: &kymaClient,
		}

		expectedCliErr := clierror.Wrap(
			errors.New("test error"),
			clierror.New("failed to check module state"),
		)

		err := enable(buffer, context.Background(), &client, "keda", "fast", false, testKedaCR)
		require.Equal(t, expectedCliErr, err)
		require.Equal(t, "adding keda module to the Kyma CR\nwaiting for module to be ready\n", buffer.String())
	})

	t.Run("failed to apply custom resource", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		kymaClient := fake.KymaClient{
			ReturnErr: nil,
		}
		rootlessDynamicClient := fake.RootlessDynamicClient{
			ReturnErr: errors.New("test error"),
		}
		client := fake.KubeClient{
			TestKymaInterface:            &kymaClient,
			TestRootlessDynamicInterface: &rootlessDynamicClient,
		}

		expectedCliErr := clierror.Wrap(
			errors.New("test error"),
			clierror.New("failed to apply custom cr from path"),
		)

		err := enable(buffer, context.Background(), &client, "keda", "fast", false, testKedaCR)
		require.Equal(t, expectedCliErr, err)
		require.Equal(t, "adding keda module to the Kyma CR\nwaiting for module to be ready\napplying kyma-system/default cr\n", buffer.String())
	})
}
