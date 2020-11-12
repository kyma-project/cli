package installation

import (
	"errors"
	"testing"

	k8sTesting "k8s.io/client-go/testing"

	"github.com/kyma-project/cli/internal/kube/mocks"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetClusterInfoFromConfigMap(t *testing.T) {
	t.Parallel()
	kymaMock := &mocks.KymaKube{}

	// Happy path
	k8sMock := fake.NewSimpleClientset(
		&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kyma-cluster-info",
				Namespace: "kube-system",
			},
			Data: map[string]string{
				"isLocal":       "true",
				"provider":      "nimbus",
				"profile":       "test",
				"localIP":       "0.0.0.0",
				"localVMDriver": "hyperfake",
			},
		},
	)
	kymaMock.On("Static").Return(k8sMock).Once()

	ci, err := GetClusterInfoFromConfigMap(kymaMock)
	require.NoError(t, err, "Test case: Happy path")
	require.Equal(t, ci.Provider, "nimbus", "Test case: Happy path")
	require.Equal(t, ci.LocalIP, "0.0.0.0", "Test case: Happy path")
	require.Equal(t, ci.LocalVMDriver, "hyperfake", "Test case: Happy path")
	require.Equal(t, ci.Profile, "test", "Test case: Happy path")
	require.True(t, ci.IsLocal, "Test case: Happy path")

	// corrupted cluster info
	k8sMock = fake.NewSimpleClientset(
		&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kyma-cluster-info",
				Namespace: "kube-system",
			},
			Data: map[string]string{
				"isLocal": "dwjkefwb·$%·$hdfbsjd234",
			},
		},
	)
	kymaMock.On("Static").Return(k8sMock).Once()

	ci, err = GetClusterInfoFromConfigMap(kymaMock)
	require.NoError(t, err, "Test case: Cluster info not found")
	require.False(t, ci.IsLocal, "Test case: Corrupted cluster info")

	// cluster info not found
	k8sMock = fake.NewSimpleClientset()
	kymaMock.On("Static").Return(k8sMock).Once()

	ci, err = GetClusterInfoFromConfigMap(kymaMock)
	require.NoError(t, err, "Test case: Cluster info not found")
	require.Equal(t, ci.Provider, "", "Test case: Cluster info not found")
	require.Equal(t, ci.LocalIP, "", "Test case: Cluster info not found")
	require.Equal(t, ci.LocalVMDriver, "", "Test case: Cluster info not found")
	require.Equal(t, ci.Profile, "", "Test case: Cluster info not found")
	require.False(t, ci.IsLocal, "Test case: Cluster info not found")

	// error getting cluster info
	k8sMock = fake.NewSimpleClientset()
	k8sMock.PrependReactor("get", "configmaps", func(action k8sTesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("Something went wrong")
	})
	kymaMock.On("Static").Return(k8sMock).Once()

	ci, err = GetClusterInfoFromConfigMap(kymaMock)
	require.Error(t, err, "Test case: Error getting cluster info")
}
