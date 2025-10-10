package envs

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	corev1 "k8s.io/api/core/v1"
)

func BuildFromFile(envs types.SourcedEnvArray) ([]corev1.EnvVar, error) {
	var result []corev1.EnvVar
	for _, e := range envs.Values {
		if e.Location == "" {
			return nil, fmt.Errorf("missing file path in env: '%s'", e.String())
		}

		data, err := godotenv.Read(e.Location)
		if err != nil {
			return nil, err
		}

		if e.Name != "" {
			// single env var from file
			value, found := data[e.LocationKey]
			if !found {
				return nil, fmt.Errorf("key '%s' not found in env file '%s'", e.LocationKey, e.Location)
			}

			result = append(result, corev1.EnvVar{
				Name:  e.Name,
				Value: value,
			})
		} else {
			// multi env vars from file
			result = append(result, buildAllFileEnvs(data, e.LocationKeysPrefix)...)
		}
	}

	return result, nil
}

func buildAllFileEnvs(data map[string]string, prefix string) []corev1.EnvVar {
	var result []corev1.EnvVar
	for key, value := range data {
		result = append(result, corev1.EnvVar{
			Name:  fmt.Sprintf("%s%s", prefix, key),
			Value: value,
		})
	}

	return result
}
