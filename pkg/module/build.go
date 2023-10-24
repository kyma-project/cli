package module

import (
	"fmt"
	"github.com/kyma-project/cli/pkg/module/gitsource"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	ocm "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"os"
	"sigs.k8s.io/yaml"
)

type Source interface {
	FetchSource(ctx cpi.Context, path, repo, version string) (*ocm.Source, error)
}

// CreateArchive creates a component archive with the given configuration.
// An empty vfs.FileSystem causes a FileSystem to be created in
// the temporary OS folder
func CreateArchive(fs vfs.FileSystem, path string, cd *ocm.ComponentDescriptor) (*comparch.ComponentArchive, error) {
	if err := fs.MkdirAll(path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create component-archive path %q: %w", fs.Normalize(path), err)
	}
	archiveFs, err := projectionfs.New(fs, path)
	if err != nil {
		return nil, fmt.Errorf("unable to create projectionfilesystem: %w", err)
	}

	ctx := cpi.DefaultContext()

	if err != nil {
		return nil, fmt.Errorf("unable to build archive for minimal descriptor: %w", err)
	}
	descriptorVersioned, err := ocm.Convert(cd, &ocm.EncodeOptions{SchemaVersion: cd.SchemaVersion()})
	if err != nil {
		return nil, fmt.Errorf("unable to build archive for minimal descriptor: %w", err)
	}
	data, err := yaml.Marshal(descriptorVersioned)
	if err != nil {
		return nil, err
	}
	file, err := archiveFs.Create("component-descriptor.yaml")
	if err != nil {
		return nil, err
	}
	_, err = file.Write(data)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return comparch.New(
		ctx,
		accessobj.ACC_CREATE, archiveFs, nil,
		nil,
		vfs.ModePerm,
	)
}

// AddGitSources adds the git sources to the component descriptor
func AddGitSources(cd *ocm.ComponentDescriptor, def *Definition, gitRemote string) error {
	var err error
	if def.Repo, err = gitsource.DetermineRepositoryURL(gitRemote, def.Repo, def.Source); err != nil {
		return err
	}
	src, err := gitsource.FetchSource(def.Source, def.Repo, def.Version)
	if err != nil {
		return err
	}
	appendSourcesForCd(cd, src)

	return nil
}

// appendSourcesForCd appends the given source to the component descriptor.
func appendSourcesForCd(cd *ocm.ComponentDescriptor, src *ocm.Source) {
	if idx := cd.GetSourceIndex(&src.SourceMeta); idx < 0 {
		cd.Sources = append(cd.Sources, *src)
	} else {
		cd.Sources[idx] = *src
	}
}
