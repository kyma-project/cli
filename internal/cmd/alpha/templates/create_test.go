package templates

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/parameters"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/types"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_create(t *testing.T) {
	t.Run("build proper command", func(t *testing.T) {
		cmd := fixCreateCommand(bytes.NewBuffer([]byte{}), &mockGetter{})

		require.Equal(t, "create", cmd.Use)
		require.Equal(t, "create test deploy", cmd.Short)
		require.Equal(t, "use this to create test deploy", cmd.Long)

		require.NotNil(t, cmd.Flag("name"))
		require.NotNil(t, cmd.Flag("namespace"))
		require.NotNil(t, cmd.Flag("replicas"))

	})

	t.Run("create custom resource", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{})
		fakeClient := &fake.RootlessDynamicClient{}
		mock := mockGetter{
			client: &fake.KubeClient{
				TestRootlessDynamicInterface: fakeClient,
			},
		}
		cmd := fixCreateCommand(buf, &mock)

		cmd.SetArgs([]string{"--name", "test-deploy", "--namespace", "test-namespace", "--replicas", "2"})
		err := cmd.Execute()
		require.NoError(t, err)

		require.Equal(t, "resource test-namespace/test-deploy applied\n", buf.String())

		require.Len(t, fakeClient.ApplyObjs, 1)
		require.Equal(t, fixUnstructuredDeployment(), fakeClient.ApplyObjs[0])
	})

	t.Run("failed to get client from getter", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{})
		mock := mockGetter{
			clierror: clierror.New("test error"),
			client:   nil,
		}

		err := createResource(&createArgs{
			out:           buf,
			ctx:           context.Background(),
			clientGetter:  &mock,
			createOptions: fixCreateOptions(),
		})
		require.Equal(t, clierror.New("test error"), err)
	})

	t.Run("failed to apply resource", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{})
		fakeClient := &fake.RootlessDynamicClient{
			ReturnErr: errors.New("test error"),
		}
		mock := mockGetter{
			client: &fake.KubeClient{
				TestRootlessDynamicInterface: fakeClient,
			},
		}

		err := createResource(&createArgs{
			out:           buf,
			ctx:           context.Background(),
			clientGetter:  &mock,
			createOptions: fixCreateOptions(),
			extraValues: []parameters.Value{
				parameters.NewTyped(types.StringCustomFlagType, ".metadata.name", "test-name"),
				parameters.NewTyped(types.StringCustomFlagType, ".metadata.namespace", "test-namespace"),
				parameters.NewTyped(types.IntCustomFlagType, ".spec.replicas", 1),
			},
		})
		require.Equal(t, clierror.Wrap(errors.New("test error"), clierror.New("failed to create resource")), err)
	})
}

type mockGetter struct {
	clierror clierror.Error
	client   kube.Client
}

func (m *mockGetter) GetKubeClientWithClierr() (kube.Client, clierror.Error) {
	return m.client, m.clierror
}

func fixCreateCommand(writer io.Writer, clietGetter KubeClientGetter) *cobra.Command {
	return buildCreateCommand(writer, clietGetter, fixCreateOptions())
}

func fixCreateOptions() *CreateOptions {
	return &CreateOptions{
		ResourceInfo: types.ResourceInfo{
			Scope:   types.NamespaceScope,
			Kind:    "Deployment",
			Group:   "apps",
			Version: "v1",
		},
		CreateCommand: types.CreateCommand{
			Description:     "create test deploy",
			DescriptionLong: "use this to create test deploy",
			CustomFlags: []types.CreateCustomFlag{
				{
					Type:         types.IntCustomFlagType,
					Name:         "replicas",
					Description:  "test flag",
					Shorthand:    "r",
					Path:         ".spec.replicas",
					DefaultValue: 3,
					Required:     false,
				},
			},
		},
	}
}

func fixUnstructuredDeployment() unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "test-deploy",
				"namespace": "test-namespace",
			},
			"spec": map[string]interface{}{
				"replicas": int64(2),
			},
		},
	}
}
