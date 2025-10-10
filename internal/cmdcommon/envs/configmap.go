package envs

import (
	"context"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildFromConfigmap(ctx context.Context, client kube.Client, namespace string, envs types.SourcedEnvArray) ([]corev1.EnvVar, error) {
	var result []corev1.EnvVar
	for _, e := range envs.Values {
		if e.Location == "" {
			return nil, fmt.Errorf("missing configmap name in env: '%s'", e.String())
		}

		if e.Name != "" {
			// single env var from configmap
			if e.LocationKey == "" {
				return nil, fmt.Errorf("missing configmap key in env: '%s'", e.String())
			}

			result = append(result, corev1.EnvVar{
				Name: e.Name,
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						Key: e.LocationKey,
						LocalObjectReference: corev1.LocalObjectReference{
							Name: e.Location,
						},
					},
				},
			})
			continue
		}

		// multi env vars from configmap
		data, err := getConfigmapData(ctx, client, namespace, e.Location)
		if err != nil {
			return nil, fmt.Errorf("while reading configmap '%s': %w", e.Location, err)
		}

		cmEnvs := buildConfigmapAllKeyEnvs(data, e.Location, e.LocationKeysPrefix)
		result = append(result, cmEnvs...)
	}

	return result, nil
}

func getConfigmapData(ctx context.Context, client kube.Client, namespace, name string) (map[string]string, error) {
	cm, err := client.Static().CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return cm.Data, nil
}

func buildConfigmapAllKeyEnvs(data map[string]string, resName string, prefix string) []corev1.EnvVar {
	var result []corev1.EnvVar
	for k := range data {
		result = append(result, corev1.EnvVar{
			Name: fmt.Sprintf("%s%s", prefix, k),
			ValueFrom: &corev1.EnvVarSource{
				ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
					Key: k,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: resName,
					},
				},
			},
		})
	}

	return result
}
