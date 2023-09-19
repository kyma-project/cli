package module

import (
	"fmt"
	"os"
	"strings"

	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compatattr"
	ocm "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	compdescv2 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/v2"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"

	"github.com/kyma-project/cli/pkg/module/gitsource"
)

type Source interface {
	FetchSource(ctx cpi.Context, path, repo, version string) (*ocm.Source, error)
}

// CreateArchive creates a component archive with the given configuration.
// An empty vfs.FileSystem causes a FileSystem to be created in
// the temporary OS folder
func CreateArchive(fs vfs.FileSystem, path, gitRemote string, def *Definition, isTargetDirAGitRepo bool) (*comparch.ComponentArchive, error) {
	if err := def.validate(); err != nil {
		return nil, err
	}

	// build minimal archive

	if err := fs.MkdirAll(path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create component-archive path %q: %w", fs.Normalize(path), err)
	}
	archiveFs, err := projectionfs.New(fs, path)
	if err != nil {
		return nil, fmt.Errorf("unable to create projectionfilesystem: %w", err)
	}

	ctx := cpi.DefaultContext()
	if err := compatattr.Set(ctx, def.SchemaVersion == compdescv2.SchemaVersion); err != nil {
		return nil, fmt.Errorf("could not set compatibility attribute for v2: %w", err)
	}

	archive, err := comparch.New(
		ctx,
		accessobj.ACC_CREATE, archiveFs,
		nil,
		nil,
		vfs.ModePerm,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to build archive for minimal descriptor: %w", err)
	}

	cd := archive.GetDescriptor()
	cd.Metadata.ConfiguredVersion = def.SchemaVersion
	builtByCLI, err := v1.NewLabel("kyma-project.io/built-by", "cli", v1.WithVersion("v1"))
	if err != nil {
		return nil, err
	}

	if compatattr.Get(ctx) {
		cd.Provider = v1.Provider{Name: "internal"}
	} else {
		cd.Provider = v1.Provider{Name: "kyma-project.io", Labels: v1.Labels{*builtByCLI}}
	}

	if isTargetDirAGitRepo {
		if err := addSources(ctx, cd, def, gitRemote); err != nil {
			return nil, err
		}
	}
	cd.ComponentSpec.SetName(def.Name)
	cd.ComponentSpec.SetVersion(def.Version)

	ocm.DefaultResources(cd)

	if err := ocm.Validate(cd); err != nil {
		return nil, fmt.Errorf("unable to validate component descriptor: %w", err)
	}

	return archive, nil
}

// addSources adds the sources to the component descriptor. If the def.Source is a git repository
func addSources(ctx cpi.Context, cd *ocm.ComponentDescriptor, def *Definition, gitRemote string) error {
	if strings.HasSuffix(def.Source, ".git") {
		gitSource := gitsource.NewGitSource()
		if def.Repo == "" {
			var err error
			if def.Repo, err = gitSource.DetermineRepositoryURL(gitRemote, def.Repo, def.Source); err != nil {
				return err
			}
		}
		src, err := gitSource.FetchSource(ctx, def.Source, def.Repo, def.Version)

		if err != nil {
			return err
		}
		appendSourcesForCd(cd, src)
	}

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
