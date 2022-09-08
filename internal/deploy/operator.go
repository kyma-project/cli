package deploy

import (
	"fmt"
	"path/filepath"

	"github.com/kyma-project/cli/internal/config"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/kustomize"
)

const (
	managerURLPattern = "https://github.com/kyma-project/%s/%s?ref=%s" // fill pattern with repo, path to kustomization and ref (commit, branch, tag...)
	localSrc          = "local"
)

func Operators(source string, k8s kube.KymaKube, dryRun bool) error {
	// version defaults
	if source == "" {
		source = config.DefaultManagerVersion
	}

	// build and apply
	lifecycleManifest, err := buildApply("lifecycle-manager", source, k8s, dryRun)
	if err != nil {
		return err
	}

	moduleManifest, err := buildApply("module-manager", source, k8s, dryRun)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Printf("%s---\n%s", lifecycleManifest, moduleManifest)
	}

	return nil
}

func buildApply(operator, source string, k8s kube.KymaKube, dryRun bool) ([]byte, error) {
	res := ""               // location of the resources to build. Can be a URL or a local path
	if source == localSrc { // build the manifests and apply them from local sources
		path, err := resolveLocalRepo(filepath.Join("github.com", "kyma-project", operator))
		if err != nil {
			return nil, fmt.Errorf("error resolving %s: %w", operator, err)
		}

		res = filepath.Join(path, "operator", "config", "default")

	} else { // build and apply from github
		res = fmt.Sprintf(managerURLPattern, operator, "operator/config/default", source)
	}

	//build and apply operator
	manifest, err := kustomize.Build(res)
	if err != nil {
		return nil, fmt.Errorf("could not build manifest for %s: %w", operator, err)
	}

	if !dryRun {
		// Dynamic client apply
		if err := k8s.Apply(manifest); err != nil {
			return nil, fmt.Errorf("could not apply operator for %s: %w", operator, err)
		}
	}
	return manifest, err
}
