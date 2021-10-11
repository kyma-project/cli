package values

import (
	"encoding/base64"
	"github.com/kyma-project/cli/internal/clusterinfo"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/kyma-incubator/reconciler/pkg/reconciler/workspace"
	"github.com/kyma-project/cli/internal/resolve"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/strvals"
)

type Values map[string]interface{}

func Merge(opts Sources, workspace *workspace.Workspace) (Values, error) {
	builder := &builder{}

	if err := addDefaultValues(builder, workspace); err != nil {
		return nil, err
	}

	if err := addClusterSpecificDefaults(builder); err != nil {
		return nil, err
	}

	if err := addValueFiles(builder, opts, workspace); err != nil {
		return nil, err
	}

	if err := addValues(builder, opts); err != nil {
		return nil, err
	}

	if err := addDomainValues(builder, opts); err != nil {
		return nil, err
	}

	vals, err := builder.build()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build values")
	}

	return vals, nil
}

func addDefaultValues(builder *builder, workspace *workspace.Workspace) error {
	kyma2OverridesPath := path.Join(workspace.InstallationResourceDir, "values.yaml")
	if err := builder.addValuesFile(kyma2OverridesPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return errors.Wrap(err, "failed to add default values file")
	}

	return nil
}

func addClusterSpecificDefaults(builder *builder) error {
	var info clusterinfo.Info
	switch info.Provider {
	case clusterinfo.K3d:
		return builder.addGlobalDomainName(defaultLocalKymaDomain)
	case clusterinfo.Gardener:
		return builder.addGlobalDomainName(info.Domain)
	default:
		return builder.addGlobalDomainName(defaultRemoteKymaDomain)
	}

	//builder.addInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, newCertificateInterceptor("global.tlsCrt", "global.tlsKey", info))
	//builder.addInterceptor([]string{"serverless.dockerRegistry.internalServerAddress", "serverless.dockerRegistry.serverAddress", "serverless.dockerRegistry.registryAddress"}, newRegistryInterceptor(info))
	//builder.addInterceptor([]string{"serverless.dockerRegistry.enableInternal"}, newRegistryDisableInterceptor(info))
}

func addValueFiles(builder *builder, opts Sources, workspace *workspace.Workspace) error {
	valueFiles, err := resolve.Files(opts.ValueFiles, filepath.Join(workspace.WorkspaceDir, "tmp"))
	if err != nil {
		return errors.Wrap(err, "failed to resolve value files")
	}
	for _, file := range valueFiles {
		if err := builder.addValuesFile(file); err != nil {
			return errors.Wrap(err, "failed to add a values file")
		}
	}

	return nil
}

func addValues(builder *builder, opts Sources) error {
	for _, value := range opts.Values {
		nested, err := strvals.Parse(value)
		if err != nil {
			return errors.Wrapf(err, "failed to parse %s", value)
		}

		if err := builder.addValues(nested); err != nil {
			return errors.Wrapf(err, "failed to add values %s", value)
		}
	}

	return nil
}

func addDomainValues(builder *builder, opts Sources) error {
	domainOverrides := make(map[string]interface{})
	if opts.Domain != "" {
		domainOverrides["domainName"] = opts.Domain
	}

	if opts.TLSCrtFile != "" && opts.TLSKeyFile != "" {
		tlsCrt, err := readFileAndEncode(opts.TLSCrtFile)
		if err != nil {
			return errors.Wrap(err, "failed to read")
		}
		tlsKey, err := readFileAndEncode(opts.TLSKeyFile)
		if err != nil {
			return errors.Wrap(err, "failed to read")
		}
		domainOverrides["tlsKey"] = tlsKey
		domainOverrides["tlsCrt"] = tlsCrt
	}

	if len(domainOverrides) > 0 {
		if err := builder.addValues(map[string]interface{}{
			"global": domainOverrides,
		}); err != nil {
			return err
		}
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
