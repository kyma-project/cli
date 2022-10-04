package clusterinfo

import (
	"context"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestDiscover(t *testing.T) {
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

		info, err := Discover(context.Background(), clientset)
		require.NoError(t, err)

		gardener, isGardener := info.(Gardener)
		require.True(t, isGardener)
		require.Equal(t, "my.cool.gardener.domain.com", gardener.Domain)
	})

	t.Run("k3d cluster", func(t *testing.T) {
		clientset := fake.NewSimpleClientset(&corev1.NodeList{
			Items: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "k3d-cool-cluster-server-0",
						Labels: map[string]string{
							"node-role.kubernetes.io/master": "true",
						},
					},
				},
			},
		})

		info, err := Discover(context.Background(), clientset)
		require.NoError(t, err)

		k3d, isK3d := info.(K3d)
		require.True(t, isK3d)
		require.Equal(t, "cool-cluster", k3d.ClusterName)
	})

	t.Run("gke cluster", func(t *testing.T) {
		clientset := fake.NewSimpleClientset(&corev1.NodeList{
			Items: []corev1.Node{
				{
					Status: corev1.NodeStatus{
						NodeInfo: corev1.NodeSystemInfo{
							KubeProxyVersion: "v1.23.10-gke.1000",
						},
					},
				},
			},
		})

		info, err := Discover(context.Background(), clientset)
		require.NoError(t, err)

		_, isGke := info.(GKE)
		require.True(t, isGke)
	})

	t.Run("unrecognized cluster", func(t *testing.T) {
		clientset := fake.NewSimpleClientset(&corev1.NodeList{
			Items: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gke-cluster-server-0",
					},
				},
			},
		})

		info, err := Discover(context.Background(), clientset)
		require.NoError(t, err)

		_, isUnrecognized := info.(Unrecognized)
		require.True(t, isUnrecognized)
	})
}
