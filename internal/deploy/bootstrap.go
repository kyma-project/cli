package deploy

import (
	"bytes"
	"fmt"
	"regexp"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/kustomize"
)

const (
	defaultLifecycleManager = "https://github.com/kyma-project/lifecycle-manager/operator/config/default"
	defaultModuleManager    = "https://github.com/kyma-project/module-manager/operator/config/default"
	defaultRetries          = 3
	defaultInitialBackoff   = 3 * time.Second
)

// Bootstrap deploys the kustomization files for the prerequisites for Kyma.
// Returns true if the Kyma CRD was deployed.
func Bootstrap(kustomizations []string, k8s kube.KymaKube, dryRun bool) (bool, error) {
	defs := []kustomize.Definition{}
	// defaults
	if len(kustomizations) == 0 {
		lm, err := kustomize.ParseKustomization(defaultLifecycleManager)
		if err != nil {
			return false, err
		}
		mm, err := kustomize.ParseKustomization(defaultModuleManager)
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

	// build manifests
	manifests, err := build(defs)
	if err != nil {
		return false, err
	}

	// apply manifests with incremental retry
	if dryRun {
		fmt.Println(string(manifests))
	} else {
		err := retry.Do(func() error {
			return k8s.Apply(manifests)
		}, retry.Attempts(defaultRetries), retry.Delay(defaultInitialBackoff), retry.DelayType(retry.BackOffDelay), retry.LastErrorOnly(false))

		if err != nil {
			return false, err
		}
	}

	return hasKyma(string(manifests))
}

func build(kustomizations []kustomize.Definition) ([]byte, error) {
	ms := bytes.Buffer{}
	for _, k := range kustomizations {
		manifest, err := kustomize.Build(k)
		if err != nil {
			return nil, fmt.Errorf("could not build manifest for %s: %w", k.Name, err)
		}
		ms.Write(manifest)
		ms.WriteString("\n---\n")
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
