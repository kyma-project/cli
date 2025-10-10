package envs

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	corev1 "k8s.io/api/core/v1"
)

func Build(envs types.EnvMap) []corev1.EnvVar {
	var result []corev1.EnvVar
	for k, v := range envs.Values {
		result = append(result, corev1.EnvVar{
			Name:  k,
			Value: fmt.Sprintf("%v", v),
		})
	}

	return result
}
