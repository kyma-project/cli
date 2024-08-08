package cluster

import (
	"context"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_fake "k8s.io/client-go/kubernetes/fake"
	"testing"
)

func Test_AssureNamespace(t *testing.T) {
	t.Run("Should do nothing when namespace exists", func(t *testing.T) {
		staticClient := k8s_fake.NewSimpleClientset(
			&v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "kyma-system",
				},
			},
		)

		cliErr := AssureNamespace(context.Background(), staticClient, "kyma-system")
		require.Nil(t, cliErr)

		ns, err := staticClient.CoreV1().Namespaces().Get(context.Background(), "kyma-system", metav1.GetOptions{})
		require.NoError(t, err)
		require.NotNil(t, ns)
	})
	t.Run("Should create namespace when it does not exist", func(t *testing.T) {
		staticClient := k8s_fake.NewSimpleClientset()

		cliErr := AssureNamespace(context.Background(), staticClient, "kyma-system")
		require.Nil(t, cliErr)

		ns, err := staticClient.CoreV1().Namespaces().Get(context.Background(), "kyma-system", metav1.GetOptions{})
		require.NoError(t, err)
		require.NotNil(t, ns)
	})
}
