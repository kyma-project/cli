package deploy

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/avast/retry-go"

	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/kustomize"
)

const (
	defaultRetries            = 3
	defaultInitialBackoff     = 3 * time.Second
	wildCardRoleAndAssignment = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kyma-cli-provisioned-wildcard
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: lifecycle-manager-wildcard
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kyma-cli-provisioned-wildcard
subjects:
- kind: ServiceAccount
  name: lifecycle-manager-controller-manager
  namespace: kcp-system`
)

// Bootstrap deploys the kustomization files for the prerequisites for Kyma.
// Returns true if the Kyma CRD was deployed.
func Bootstrap(
	ctx context.Context, kustomizations []string, k8s kube.KymaKube, addWildCard, dryRun bool,
) (bool, error) {
	var defs []kustomize.Definition
	// defaults
	for _, k := range kustomizations {
		parsed, err := kustomize.ParseKustomization(k)
		if err != nil {
			return false, err
		}
		defs = append(defs, parsed)
	}

	// build manifests
	manifests, err := build(defs)
	if err != nil {
		return false, err
	}

	if addWildCard {
		manifests = append(manifests, []byte(wildCardRoleAndAssignment)...)
	}

	// apply manifests with incremental retry
	if dryRun {
		fmt.Println(string(manifests))
	} else {
		err := retry.Do(
			func() error {
				return k8s.Apply(context.Background(), manifests)
			}, retry.Attempts(defaultRetries), retry.Delay(defaultInitialBackoff), retry.DelayType(retry.BackOffDelay),
			retry.LastErrorOnly(false), retry.Context(ctx),
		)

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
