package module

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//nolint:gosec
const OCIRegistryCredLabel = "oci-registry-cred"

func CreateCredMatchLabels(registryCredSelector string) ([]byte, error) {
	var matchLabels []byte
	if registryCredSelector != "" {
		selector, err := metav1.ParseToLabelSelector(registryCredSelector)
		if err != nil {
			return nil, err
		}
		matchLabels, err = json.Marshal(selector.MatchLabels)
		if err != nil {
			return nil, err
		}
	}
	return matchLabels, nil
}
