package deploy

import (
	"fmt"
	"github.com/kyma-incubator/reconciler/pkg/reconciler/workspace"
	"github.com/kyma-project/cli/internal/overrides"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/strvals"
	"io/fs"
	"k8s.io/client-go/kubernetes"
	"path"
)

func mergeOverrides(opts *Options, workspace *workspace.Workspace, kubeClient kubernetes.Interface) (map[string]interface{}, error) {
	overridesBuilder := &overrides.Builder{}

	kyma2OverridesPath := path.Join(workspace.InstallationResourceDir, "values.yaml")
	if err := overridesBuilder.AddFile(kyma2OverridesPath); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, errors.Wrap(err, "Could not add overrides for Kyma 2.0")
		}
	}

	for _, value := range opts.Values {
		ovs, err := strvals.Parse(value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse: %s", value)
		}

		overridesBuilder.AddOverrides(ovs)
	}

	overridesBuilder.AddInterceptor([]string{"global.domainName", "global.ingress.domainName"}, overrides.NewDomainNameOverrideInterceptor(kubeClient))
	overridesBuilder.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, overrides.NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", kubeClient))
	overridesBuilder.AddInterceptor([]string{"serverless.dockerRegistry.internalServerAddress", "serverless.dockerRegistry.serverAddress", "serverless.dockerRegistry.registryAddress"}, overrides.NewRegistryInterceptor(kubeClient))
	overridesBuilder.AddInterceptor([]string{"serverless.dockerRegistry.enableInternal"}, overrides.NewRegistryDisableInterceptor(kubeClient))

	ovs, err := overridesBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build: %v", err)
	}

	return ovs.FlattenedMap(), nil
}
