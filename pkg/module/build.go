package module

import (
	"fmt"
	"os"

	"github.com/kyma-project/cli/pkg/module/git"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	ocm "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
)

// Build creates a component archive with the given configuration.
// An empty vfs.FileSystem causes a FileSystem to be created in
// the temporary OS folder
func Build(fs vfs.FileSystem, def *Definition) (*comparch.ComponentArchive, error) {
	if err := def.validate(); err != nil {
		return nil, err
	}

	//Overwrite == true OR (Overwrite == false AND the component descriptor does not exist)
	return buildFull(fs, "mod", def)
}

func buildFull(fs vfs.FileSystem, path string, def *Definition) (*comparch.ComponentArchive, error) {
	// build minimal archive

	if err := fs.MkdirAll(path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create component-archive path %q: %w", fs.Normalize(path), err)
	}
	archiveFs, err := projectionfs.New(fs, path)
	if err != nil {
		return nil, fmt.Errorf("unable to create projectionfilesystem: %w", err)
	}

	archive, err := comparch.New(
		cpi.DefaultContext(),
		accessobj.ACC_CREATE, archiveFs,
		nil,
		nil,
		vfs.ModePerm,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to build archive for minimal descriptor: %w", err)
	}

	cd := archive.GetDescriptor()
	if err := addSources(cd, def); err != nil {
		return nil, err
	}
	cd.ComponentSpec.SetName(def.Name)
	cd.ComponentSpec.SetVersion(def.Version)
	builtByCLI, err := v1.NewLabel("kyma-project.io/built-by", "cli", v1.WithVersion("v1"))
	if err != nil {
		return nil, err
	}
	cd.Provider = v1.Provider{Name: "kyma-project.io", Labels: v1.Labels{*builtByCLI}}

	ocm.DefaultResources(cd)

	if err := ocm.Validate(cd); err != nil {
		return nil, fmt.Errorf("unable to validate component descriptor: %w", err)
	}

	return archive, nil
}

func addSources(cd *ocm.ComponentDescriptor, def *Definition) error {
	src, err := git.Source(def.Source, def.Repo, def.Version)
	if err != nil {
		return err
	}

	if idx := cd.GetSourceIndex(&src.SourceMeta); idx < 0 {
		cd.Sources = append(cd.Sources, *src)
	} else {
		cd.Sources[idx] = *src
	}

	return nil
}
