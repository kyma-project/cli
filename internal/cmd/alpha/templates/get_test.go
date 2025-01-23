package templates

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/types"
	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_get(t *testing.T) {
	t.Run("build proper command for namespaced resource", func(t *testing.T) {
		cmd := buildGetCommand(bytes.NewBuffer([]byte{}), &mockGetter{}, &GetOptions{
			GetCommand: types.GetCommand{
				Description:     "get test deploy",
				DescriptionLong: "use this to get test deploy",
			},
			ResourceInfo: types.ResourceInfo{
				Scope: types.NamespaceScope,
			},
		})

		require.Equal(t, "get", cmd.Use)
		require.Equal(t, "get test deploy", cmd.Short)
		require.Equal(t, "use this to get test deploy", cmd.Long)

		require.NotNil(t, cmd.Flag("name"))
		require.NotNil(t, cmd.Flag("namespace"))
		require.NotNil(t, cmd.Flag("all-namespaces"))
	})

	t.Run("get resources with custom field", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{})
		fakeClient := &fake.RootlessDynamicClient{
			ReturnListObjs: &unstructured.UnstructuredList{
				Items: fixGetResources(),
			},
		}
		mock := mockGetter{
			client: &fake.KubeClient{
				TestRootlessDynamicInterface: fakeClient,
			},
		}
		cmd := buildGetCommand(buf, &mock, &GetOptions{
			GetCommand: types.GetCommand{
				Description:     "get test deploy",
				DescriptionLong: "use this to get test deploy",
				Parameters: []types.Parameter{
					{
						Path: ".metadata.generation",
						Name: "gen",
					},
				},
			},
			ResourceInfo: types.ResourceInfo{
				Scope: types.NamespaceScope,
			},
		})

		err := cmd.Execute()
		require.NoError(t, err)

		expectedOutput := "NAME  \tNAMESPACE\tGEN \n" +
			"name-1\tkyma     \t1  \t\n" +
			"name-2\tkyma     \t5  \t\n" +
			"name-3\tkyma     \t12 \t\n" +
			"name-4\tkyma     \t43 \t\n"
		require.Equal(t, expectedOutput, buf.String())
	})

	t.Run("failed to get client", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{})
		mock := mockGetter{
			clierror: clierror.New("test error"),
			client:   nil,
		}

		err := getResources(&getArgs{
			out:          buf,
			ctx:          context.Background(),
			clientGetter: &mock,
			getOptions: &GetOptions{
				ResourceInfo: types.ResourceInfo{
					Scope: types.NamespaceScope,
				},
			},
		})
		require.Equal(t, clierror.New("test error"), err)
	})

	t.Run("failed to delete object", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{})
		fakeClient := &fake.RootlessDynamicClient{
			ReturnErr: errors.New("test error"),
		}
		mock := mockGetter{
			client: &fake.KubeClient{
				TestRootlessDynamicInterface: fakeClient,
			},
		}

		err := getResources(&getArgs{
			out:          buf,
			ctx:          context.Background(),
			clientGetter: &mock,
			getOptions: &GetOptions{
				ResourceInfo: types.ResourceInfo{
					Scope: types.NamespaceScope,
				},
			},
		})
		require.Equal(t, clierror.Wrap(errors.New("test error"), clierror.New("failed to get resource")), err)
	})
}

func fixGetResources() []unstructured.Unstructured {
	return []unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":       "name-1",
					"namespace":  "kyma",
					"generation": 1,
				},
			},
		},
		{
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":       "name-2",
					"namespace":  "kyma",
					"generation": 5,
				},
			},
		},
		{
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":       "name-3",
					"namespace":  "kyma",
					"generation": 12,
				},
			},
		},
		{
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":       "name-4",
					"namespace":  "kyma",
					"generation": 43,
				},
			},
		},
	}
}
