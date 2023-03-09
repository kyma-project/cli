package deploy

import (
	"context"
	"github.com/kyma-project/cli/internal/config"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/kustomize"
	"sigs.k8s.io/kustomize/api/filters/fieldspec"
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/resid"
)

const modulesKustomization = "https://github.com/kyma-project/kyma/modules@" + config.DefaultKyma2Version

func ModuleTemplates(ctx context.Context, k8s kube.KymaKube, templates []string, target string, force, dryRun bool) error {
	var defs []kustomize.Definition
	for _, k := range templates {
		parsed, err := kustomize.ParseKustomization(k)
		if err != nil {
			return err
		}
		defs = append(defs, parsed)
	}

	filter := fieldspec.Filter{
		FieldSpec: types.FieldSpec{
			Gvk: resid.Gvk{
				Group: "operator.kyma-project.io",
				Kind:  "ModuleTemplate",
			},
			Path:               "spec/target",
			CreateIfNotPresent: false,
		},
		SetValue: filtersutil.SetScalar(target),
	}

	manifests, err := kustomize.BuildMany(defs, []kio.Filter{kio.FilterAll(filter)})
	if err != nil {
		return err
	}

	return applyManifests(
		ctx, k8s, manifests, applyOpts{
			dryRun, force, defaultRetries, defaultInitialBackoff},
	)
}

func DefaultModuleTemplates(ctx context.Context, k8s kube.KymaKube, target string, force, dryRun bool) error {
	return ModuleTemplates(ctx, k8s, []string{modulesKustomization}, target, force, dryRun)
}
