package clusterinfo

import (
	"context"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestGet(t *testing.T) {
	t.Run("gardener cluster", func(t *testing.T) {
		clientset := fake.NewSimpleClientset(&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "shoot-info",
			},
			Data: map[string]string{
				"domain": "my.cool.gardener.domain.com",
			},
		})

		info, err := Get(context.Background(), clientset)
		require.NoError(t, err)
		require.Equal(t, Gardener, info.ClusterType)
		require.Equal(t, "my.cool.gardener.domain.com", info.Domain)
	})

	t.Run("k3d cluster", func(t *testing.T) {
		clientset := fake.NewSimpleClientset(&corev1.NodeList{
			Items: []corev1.Node{
				
			}
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "shoot-info",
			},
			Data: map[string]string{
				"domain": "my.cool.gardener.domain.com",
			},
		})

		info, err := Get(context.Background(), clientset)
		require.NoError(t, err)
		require.Equal(t, Gardener, info.ClusterType)
		require.Equal(t, "my.cool.gardener.domain.com", info.Domain)
	})
}
