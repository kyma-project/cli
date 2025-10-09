package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	corev1 "k8s.io/api/core/v1"
)

func BuildEnvsFromFile(envs types.SourcedEnvArray) ([]corev1.EnvVar, error) {
	var result []corev1.EnvVar
	for _, e := range envs.Values {
		if e.Location == "" {
			return nil, fmt.Errorf("missing file path in env: %s", e.String())
		}

		data, err := os.ReadFile(e.Location)
		if err != nil {
			return nil, fmt.Errorf("while reading env file %s: %w", e.Location, err)
		}

		if e.Name != "" {
			// single env var from file
			value, found := findFileEnv(data, e.Name)
			if !found {
				return nil, fmt.Errorf("key %s not found in env file %s", e.Name, e.Location)
			}

			result = append(result, corev1.EnvVar{
				Name:  e.Name,
				Value: value,
			})
		} else {
			// multi env vars from file
			fileEnvs, err := buildAllFileEnvs(data, e.LocationKeysPrefix)
			if err != nil {
				return nil, err
			}
			result = append(result, fileEnvs...)
		}
	}

	return result, nil
}

func findFileEnv(data []byte, key string) (string, bool) {
	for _, line := range strings.Split(string(data), "\n") {
		linePrefix := fmt.Sprintf("%s=", key)
		if strings.HasPrefix(line, linePrefix) {
			return strings.TrimPrefix(line, linePrefix), true
		}
	}

	return "", false
}

func buildAllFileEnvs(data []byte, prefix string) ([]corev1.EnvVar, error) {
	var result []corev1.EnvVar
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line format in env file: %s", line)
		}

		result = append(result, corev1.EnvVar{
			Name:  fmt.Sprintf("%s%s", prefix, parts[0]),
			Value: parts[1],
		})
	}

	return result, nil
}
