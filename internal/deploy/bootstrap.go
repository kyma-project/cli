package deploy

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/kustomize"
)

// Bootstrap deploys the kustomization files for the prerequisites for Kyma.
// Returns true if the Kyma CRD was deployed.
func Bootstrap(kustomizations []string, k8s kube.KymaKube, dryRun bool) (bool, error) {
	defs := []kustomize.Definition{}
	// defaults
	if len(kustomizations) == 0 {
		lm, err := kustomize.ParseKustomization("https://github.com/kyma-project/lifecycle-manager/operator/config/default")
		if err != nil {
			return false, err
		}
		mm, err := kustomize.ParseKustomization("https://github.com/kyma-project/module-manager/operator/config/default")
		if err != nil {
			return false, err
		}
		defs = append(defs, lm, mm)
	} else {
		for _, k := range kustomizations {
			parsed, err := kustomize.ParseKustomization(k)
			if err != nil {
				return false, err
			}
			defs = append(defs, parsed)
		}
	}

	// build and apply
	manifests, err := buildApply(defs, k8s, dryRun)
	if err != nil {
		return false, err
	}

	if dryRun {
		fmt.Println(string(manifests))
	}

	return hasKyma(string(manifests))
}

func buildApply(kustomizations []kustomize.Definition, k8s kube.KymaKube, dryRun bool) ([]byte, error) {
	ms := bytes.Buffer{}
	for _, k := range kustomizations {
		m, err := kustomize.Build(k)
		if err != nil {
			return nil, fmt.Errorf("could not build manifest for %s: %w", k.Name, err)
		}
		ms.Write(m)
		ms.WriteString("\n---\n")
	}

	if !dryRun {
		// Dynamic client apply
		if err := k8s.Apply(ms.Bytes()); err != nil {
			return nil, fmt.Errorf("could not apply kustomize manifest: %w", err)
		}
	}
	return ms.Bytes(), nil
}

// hasKyma checks if the given manifest contains the Kyma CRD
func hasKyma(manifest string) (bool, error) {
	r, err := regexp.Compile(`(names:)(?:[.\s\S]*)(kind: Kyma)(?:[.\s\S]*)(plural: kymas)`)
	if err != nil {
		return false, err
	}
	return r.MatchString(manifest), nil
}
