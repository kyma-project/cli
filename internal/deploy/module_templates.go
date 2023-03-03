package deploy

import (
	"context"

	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/kustomize"
)

func ModuleTemplates(ctx context.Context, k8s kube.KymaKube, templates []string, force, dryRun bool) error {
	var defs []kustomize.Definition
	// defaults
	for _, k := range templates {
		parsed, err := kustomize.ParseKustomization(k)
		if err != nil {
			return err
		}
		defs = append(defs, parsed)
	}

	// build manifests
	manifests, err := kustomize.BuildMany(defs, nil)
	if err != nil {
		return err
	}

	return applyManifests(
		ctx, k8s, manifests, applyOpts{
			dryRun, force, defaultRetries, defaultInitialBackoff},
	)
}
