package deploy

import (
	"encoding/base64"
	"fmt"
	"github.com/kyma-incubator/reconciler/pkg/reconciler/workspace"
	"github.com/kyma-project/cli/internal/overrides"
	"github.com/kyma-project/cli/internal/resolve"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/strvals"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"path"
	"path/filepath"
)

func mergeOverrides(opts *Options, workspace *workspace.Workspace, kubeClient kubernetes.Interface) (map[string]interface{}, error) {
	builder := &overrides.Builder{}

	if err := setOverrideFiles(builder, opts, workspace); err != nil {
		return nil, err
	}

	if err := setOverrides(builder, opts); err != nil {
		return nil, err
	}

	registerInterceptors(builder, kubeClient)

	ovs, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build: %v", err)
	}

	return ovs.FlattenedMap(), nil
}

func setOverrideFiles(builder *overrides.Builder, opts *Options, workspace *workspace.Workspace) error {
	kyma2OverridesPath := path.Join(workspace.InstallationResourceDir, "values.yaml")
	if err := builder.AddFile(kyma2OverridesPath); err != nil {
		return errors.Wrap(err, "failed to add default values file: %s")
	}

	valueFiles, err := resolve.Files(opts.ValueFiles, filepath.Join(workspace.WorkspaceDir, "tmp"))
	if err != nil {
		return errors.Wrap(err, "failed to resolve value files: %v")
	}
	for _, file := range valueFiles {
		if err := builder.AddFile(file); err != nil {
			return errors.Wrap(err, "failed to add a values file: %s")
		}
	}

	return nil
}

func setOverrides(builder *overrides.Builder, opts *Options) error {
	globalOverrides := make(map[string]interface{})
	if opts.Domain != "" {
		globalOverrides["domainName"] = opts.Domain
	}

	if opts.TLSCrtFile != "" && opts.TLSKeyFile != "" {
		tlsCrt, err := readFileAndEncode(opts.TLSCrtFile)
		if err != nil {
			return fmt.Errorf("failed to read ")
		}
		tlsKey, err := readFileAndEncode(opts.TLSKeyFile)
		if err != nil {
			return err
		}
		globalOverrides["tlsKey"] = tlsKey
		globalOverrides["tlsCrt"] = tlsCrt
	}

	if len(globalOverrides) > 0 {
		if err := builder.AddOverrides(map[string]interface{}{
			"global": globalOverrides,
		}); err != nil {
			return err
		}
	}

	for _, value := range opts.Values {
		ovs, err := strvals.Parse(value)
		if err != nil {
			return fmt.Errorf("failed to parse: %s", value)
		}

		builder.AddOverrides(ovs)
	}

	return nil
}

func readFileAndEncode(filename string) (string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(content), nil
}

func registerInterceptors(builder *overrides.Builder, kubeClient kubernetes.Interface) {
	builder.AddInterceptor([]string{"global.domainName", "global.ingress.domainName"}, overrides.NewDomainNameOverrideInterceptor(kubeClient))
	builder.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, overrides.NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", kubeClient))
	builder.AddInterceptor([]string{"serverless.dockerRegistry.internalServerAddress", "serverless.dockerRegistry.serverAddress", "serverless.dockerRegistry.registryAddress"}, overrides.NewRegistryInterceptor(kubeClient))
	builder.AddInterceptor([]string{"serverless.dockerRegistry.enableInternal"}, overrides.NewRegistryDisableInterceptor(kubeClient))
}
