package deploy

import (
	"context"
	"github.com/kyma-project/cli/internal/config"

	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/kustomize"
)

const modulesKustomization = "https://github.com/kyma-project/kyma/modules@" + config.DefaultKyma2Version

func ModuleTemplates(ctx context.Context, k8s kube.KymaKube, templates []string, force, dryRun bool) error {
	var defs []kustomize.Definition
	for _, k := range templates {
		parsed, err := kustomize.ParseKustomization(k)
		if err != nil {
			return err
		}
		defs = append(defs, parsed)
	}
	manifests, err := kustomize.BuildMany(defs, nil)
	if err != nil {
		return err
	}

	return applyManifests(
		ctx, k8s, manifests, applyOpts{
			dryRun, force, defaultRetries, defaultInitialBackoff},
	)
}

func DefaultModuleTemplates(ctx context.Context, k8s kube.KymaKube, force, dryRun bool) error {
	return ModuleTemplates(ctx, k8s, []string{modulesKustomization}, force, dryRun)
}
