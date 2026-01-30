package precheck

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeKubeClientConfig struct {
	returnClient kube.Client
	returnErr    clierror.Error
}

func (f *fakeKubeClientConfig) GetKubeClient() (kube.Client, error) {
	if f.returnErr != nil {
		return nil, errors.New("test error")
	}
	return f.returnClient, nil
}

func (f *fakeKubeClientConfig) GetKubeClientWithClierr() (kube.Client, clierror.Error) {
	return f.returnClient, f.returnErr
}

func TestRunClusterKLMManagedCheck_KubeClientError(t *testing.T) {
	expectedErr := clierror.New("failed to get kube client")
	kymaConfig := &cmdcommon.KymaConfig{
		KubeClientConfig: &fakeKubeClientConfig{
			returnClient: nil,
			returnErr:    expectedErr,
		},
		Ctx: context.Background(),
	}

	err := RunClusterKLMManagedCheck(kymaConfig)

	assert.NotNil(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestRunClusterKLMManagedCheck_ClusterManagedByKLM_NoError(t *testing.T) {
	fakeKymaClient := &fake.KymaClient{
		ReturnErr: nil,
		ReturnDefaultKyma: kyma.Kyma{
			Spec: kyma.KymaSpec{
				Modules: []kyma.Module{},
			},
		},
	}
	fakeKubeClient := &fake.KubeClient{
		TestKymaInterface: fakeKymaClient,
	}

	kymaConfig := &cmdcommon.KymaConfig{
		KubeClientConfig: &fakeKubeClientConfig{
			returnClient: fakeKubeClient,
			returnErr:    nil,
		},
		Ctx: context.Background(),
	}

	err := RunClusterKLMManagedCheck(kymaConfig)

	assert.Nil(t, err)
}

func TestRunClusterKLMManagedCheck_ClusterNotManagedByKLM_ReturnsError(t *testing.T) {
	fakeKymaClient := &fake.KymaClient{
		ReturnErr: errors.New("kyma resource not found"),
	}
	fakeKubeClient := &fake.KubeClient{
		TestKymaInterface: fakeKymaClient,
	}

	kymaConfig := &cmdcommon.KymaConfig{
		KubeClientConfig: &fakeKubeClientConfig{
			returnClient: fakeKubeClient,
			returnErr:    nil,
		},
		Ctx: context.Background(),
	}

	err := RunClusterKLMManagedCheck(kymaConfig)

	require.NotNil(t, err)
	errStr := err.String()
	assert.Contains(t, errStr, "not managed by KLM")
	assert.Contains(t, errStr, "Custom Resource Definitions")
	assert.Contains(t, errStr, "kyma module catalog")
	assert.Contains(t, errStr, "kyma module pull")
}

func TestRunClusterKLMManagedCheck_ErrorMessageContent(t *testing.T) {
	fakeKymaClient := &fake.KymaClient{
		ReturnErr: errors.New("not found"),
	}
	fakeKubeClient := &fake.KubeClient{
		TestKymaInterface: fakeKymaClient,
	}

	kymaConfig := &cmdcommon.KymaConfig{
		KubeClientConfig: &fakeKubeClientConfig{
			returnClient: fakeKubeClient,
			returnErr:    nil,
		},
		Ctx: context.Background(),
	}

	err := RunClusterKLMManagedCheck(kymaConfig)

	require.NotNil(t, err)
	errStr := err.String()

	assert.Contains(t, errStr, "List available modules from the community modules catalog")
	assert.Contains(t, errStr, "Pull a community module to your cluster")
	assert.Contains(t, errStr, "Pulling a community module will install the required dependencies")
	assert.Contains(t, errStr, "For more information, refer to the documentation")
	assert.Contains(t, errStr, "kyma module --help")
}
