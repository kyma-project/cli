package templates

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/types"
	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_remove(t *testing.T) {
	t.Run("build proper command", func(t *testing.T) {
		cmd := fixDeleteCommand(bytes.NewBuffer([]byte{}), &mockGetter{})

		require.Equal(t, "delete <resource_name> [flags]", cmd.Use)
		require.Equal(t, "delete test deploy", cmd.Short)
		require.Equal(t, "use this to delete test deploy", cmd.Long)

		require.NoError(t, cmd.ValidateArgs([]string{"resource_name"}))
		require.NotNil(t, cmd.Flag("namespace"))
	})

	t.Run("delete resource", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{})
		fakeClient := &fake.RootlessDynamicClient{}
		mock := mockGetter{
			client: &fake.KubeClient{
				TestRootlessDynamicInterface: fakeClient,
			},
		}

		cmd := fixDeleteCommand(buf, &mock)

		cmd.SetArgs([]string{"test-deploy", "--namespace", "test-namespace"})

		err := cmd.Execute()
		require.NoError(t, err)

		require.Len(t, fakeClient.RemovedObjs, 1)
		require.Equal(t, fixDeletedUnstructuredDeployment(), fakeClient.RemovedObjs[0])
	})

	t.Run("failed to get client", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{})
		mock := mockGetter{
			clierror: clierror.New("test error"),
			client:   nil,
		}

		err := deleteResource(&deleteArgs{
			out:           buf,
			ctx:           context.Background(),
			clientGetter:  &mock,
			deleteOptions: fixDeleteOptions(),
		})
		require.Equal(t, clierror.New("test error"), err)
	})

	t.Run("failed to delete object", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{})
		fakeClient := &fake.RootlessDynamicClient{
			ReturnRemoveErr: errors.New("test error"),
		}
		mock := mockGetter{
			client: &fake.KubeClient{
				TestRootlessDynamicInterface: fakeClient,
			},
		}

		err := deleteResource(&deleteArgs{
			out:           buf,
			ctx:           context.Background(),
			clientGetter:  &mock,
			deleteOptions: fixDeleteOptions(),
		})
		require.Equal(t, clierror.Wrap(errors.New("test error"), clierror.New("failed to delete resource")), err)
	})
}

func fixDeleteCommand(writer io.Writer, getter KubeClientGetter) *cobra.Command {
	return buildDeleteCommand(writer, getter, fixDeleteOptions())
}

func fixDeleteOptions() *DeleteOptions {
	return &DeleteOptions{
		DeleteCommand: types.DeleteCommand{
			Description:     "delete test deploy",
			DescriptionLong: "use this to delete test deploy",
		},
		ResourceInfo: types.ResourceInfo{
			Scope:   types.NamespaceScope,
			Kind:    "Deployment",
			Group:   "apps",
			Version: "v1",
		},
	}
}

func fixDeletedUnstructuredDeployment() unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "test-deploy",
				"namespace": "test-namespace",
			},
		},
	}
}
