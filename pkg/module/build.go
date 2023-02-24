package module

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kyma-project/cli/pkg/module/git"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	ocm "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
)

// Build creates a component archive with the given configuration
func Build(fs vfs.FileSystem, def *Definition) (*comparch.ComponentArchive, error) {
	if err := def.validate(); err != nil {
		return nil, err
	}

	compDescFilePath := filepath.Join(def.ArchivePath, comparch.ComponentDescriptorFileName)
	if !def.Overwrite {
		_, err := fs.Stat(compDescFilePath)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if err == nil {
			//Overwrite == false and component descriptor exists
			return buildWithoutOverwriting(fs, def)
		}
	}

	//Overwrite == true OR (Overwrite == false AND the component descriptor does not exist)
	return buildFull(fs, def)
}

// buildWithoutOverwriting builds over an existing descriptor without overwriting
func buildWithoutOverwriting(fs vfs.FileSystem, def *Definition) (*comparch.ComponentArchive, error) {
	// add the input to the comparch format
	archiveFs, err := projectionfs.New(fs, def.ArchivePath)
	if err != nil {
		return nil, fmt.Errorf("unable to create projectionfilesystem: %w", err)
	}

	archive, err := comparch.New(
		cpi.DefaultContext(),
		accessobj.ACC_WRITABLE, archiveFs,
		nil,
		nil,
		vfs.ModePerm,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to parse component archive from %s: %w", def.ArchivePath, err)
	}

	cd := archive.GetDescriptor()

	if err := addSources(cd, def); err != nil {
		return nil, err
	}

	if def.Name != "" {
		if cd.Name != "" && cd.Name != def.Name {
			return nil, errors.New("unable to overwrite the existing component name: forbidden")
		}
		cd.Name = def.Name
	}

	if def.Version != "" {
		if cd.Version != "" && cd.Version != def.Version {
			return nil, errors.New("unable to overwrite the existing component version: forbidden")
		}
		cd.Version = def.Version
	}

	ocm.DefaultResources(cd)

	if err = ocm.Validate(cd); err != nil {
		return nil, fmt.Errorf("invalid component descriptor: %w", err)
	}

	return archive, nil
}

func buildFull(fs vfs.FileSystem, def *Definition) (*comparch.ComponentArchive, error) {
	// build minimal archive

	if err := fs.MkdirAll(def.ArchivePath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create component-archive path %q: %w", def.ArchivePath, err)
	}
	archiveFs, err := projectionfs.New(fs, def.ArchivePath)
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
