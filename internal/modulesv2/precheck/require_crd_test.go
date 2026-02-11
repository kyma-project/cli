package precheck

import (
	"context"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube"
	kubefake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type fakeKubeClientConfig struct {
	returnClient kube.Client
	returnErr    clierror.Error
}

func (f *fakeKubeClientConfig) GetKubeClient() (kube.Client, error) {
	return f.returnClient, nil
}

func (f *fakeKubeClientConfig) GetKubeClientWithClierr() (kube.Client, clierror.Error) {
	return f.returnClient, f.returnErr
}

func TestRequireCRD_KubeClientError(t *testing.T) {
	expectedErr := clierror.New("failed to get kube client")
	kymaConfig := &cmdcommon.KymaConfig{
		KubeClientConfig: &fakeKubeClientConfig{
			returnClient: nil,
			returnErr:    expectedErr,
		},
		Ctx: context.Background(),
	}

	err := RequireCRD(kymaConfig, CmdGroupStable)

	assert.NotNil(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestRequireCRD_CRDPresent_NoError(t *testing.T) {
	// Create a fake rootless dynamic client that returns the CRD
	fakeRootlessDynamic := &kubefake.RootlessDynamicClient{
		ReturnGetObj: unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "apiextensions.k8s.io/v1",
				"kind":       "CustomResourceDefinition",
				"metadata": map[string]any{
					"name": "moduletemplates.operator.kyma-project.io",
				},
			},
		},
		ReturnGetErr: nil,
	}
	fakeKubeClient := &kubefake.KubeClient{
		TestRootlessDynamicInterface: fakeRootlessDynamic,
	}

	kymaConfig := &cmdcommon.KymaConfig{
		KubeClientConfig: &fakeKubeClientConfig{
			returnClient: fakeKubeClient,
			returnErr:    nil,
		},
		Ctx: context.Background(),
	}

	err := RequireCRD(kymaConfig, CmdGroupStable)

	assert.Nil(t, err)
}

func TestRequireCRD_CRDNotPresent_ReturnsError(t *testing.T) {
	// Create a fake rootless dynamic client that returns NotFound error
	fakeRootlessDynamic := &kubefake.RootlessDynamicClient{
		ReturnGetObj: unstructured.Unstructured{},
		ReturnGetErr: k8serrors.NewNotFound(schema.GroupResource{Group: "apiextensions.k8s.io", Resource: "customresourcedefinitions"}, "moduletemplates.operator.kyma-project.io"),
	}
	fakeKubeClient := &kubefake.KubeClient{
		TestRootlessDynamicInterface: fakeRootlessDynamic,
	}

	kymaConfig := &cmdcommon.KymaConfig{
		KubeClientConfig: &fakeKubeClientConfig{
			returnClient: fakeKubeClient,
			returnErr:    nil,
		},
		Ctx: context.Background(),
	}

	err := RequireCRD(kymaConfig, CmdGroupStable)

	require.NotNil(t, err)
	errStr := err.String()
	assert.Contains(t, errStr, "not managed by KLM")
	assert.Contains(t, errStr, "Custom Resource Definitions")
	assert.Contains(t, errStr, "kyma module catalog")
	assert.Contains(t, errStr, "kyma module pull")
}

func TestRequireCRD_ErrorMessageContent(t *testing.T) {
	// Create a fake rootless dynamic client that returns NotFound error
	fakeRootlessDynamic := &kubefake.RootlessDynamicClient{
		ReturnGetObj: unstructured.Unstructured{},
		ReturnGetErr: k8serrors.NewNotFound(schema.GroupResource{Group: "apiextensions.k8s.io", Resource: "customresourcedefinitions"}, "moduletemplates.operator.kyma-project.io"),
	}
	fakeKubeClient := &kubefake.KubeClient{
		TestRootlessDynamicInterface: fakeRootlessDynamic,
	}

	kymaConfig := &cmdcommon.KymaConfig{
		KubeClientConfig: &fakeKubeClientConfig{
			returnClient: fakeKubeClient,
			returnErr:    nil,
		},
		Ctx: context.Background(),
	}

	err := RequireCRD(kymaConfig, CmdGroupStable)

	require.NotNil(t, err)
	errStr := err.String()

	assert.Contains(t, errStr, "List available modules from the community modules catalog")
	assert.Contains(t, errStr, "Pull a community module to your cluster")
	assert.Contains(t, errStr, "Pulling a community module will install the required dependencies")
	assert.Contains(t, errStr, "For more information, refer to the documentation")
	assert.Contains(t, errStr, "kyma module --help")
}
