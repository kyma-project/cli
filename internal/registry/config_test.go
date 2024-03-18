package registry

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetConfig(t *testing.T) {
	t.Run("Should return the RegistryConfig", func(t *testing.T) {
		// given
		client := fake.NewSimpleClientset(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      registrySecretName,
				Namespace: serverlessNamespace,
			},
			Data: map[string][]byte{
				".dockerconfigjson": []byte(`{"auths": {}}`),
				"username":          []byte("testUsername"),
				"password":          []byte("testPassword"),
				"pullRegAddr":       []byte("pullRegAddr"),
				"pushRegAddr":       []byte("pushRegAddr"),
				"isInternal":        []byte("true"),
			},
		})
		expectedRegistryConfig := &RegistryConfig{
			DockerConfigJson: `{"auths": {}}`,
			Username:         "testUsername",
			Password:         "testPassword",
			PullRegAddr:      "pullRegAddr",
			PushRegAddr:      "pushRegAddr",
			IsInternal:       true,
		}

		// when
		config, err := GetConfig(context.Background(), client)

		// then
		require.NoError(t, err)
		require.Equal(t, expectedRegistryConfig, config)
	})

	t.Run("Should return an error when the secret does not exist", func(t *testing.T) {
		// given
		client := fake.NewSimpleClientset()
		expectedErrorMsg := "secrets \"serverless-registry-config-default\" not found"

		// when
		config, err := GetConfig(context.Background(), client)

		// then
		require.Error(t, err)
		require.Nil(t, config)
		require.Contains(t, err.Error(), expectedErrorMsg)
	})
}
