package module

import (
	"fmt"
	"os"
	"strings"

	"github.com/kyma-project/cli/pkg/module/gitsource"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	ocm "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	ocmV2 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/v2"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/runtime"
	"sigs.k8s.io/yaml"
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
	cdInit := ocmV2.ComponentDescriptor{
		Metadata: v1.Metadata{Version: def.SchemaVersion},
		ComponentSpec: ocmV2.ComponentSpec{
			ObjectMeta: ocmV2.ObjectMeta{
				Name:    def.Name,
				Version: def.Version,
			},
		},
	}
	if cdInit.RepositoryContexts == nil {
		cdInit.RepositoryContexts = make([]*runtime.UnstructuredTypedObject, 0)
	}
	if cdInit.Sources == nil {
		cdInit.Sources = make([]ocmV2.Source, 0)
	}
	if cdInit.Resources == nil {
		cdInit.Resources = make([]ocmV2.Resource, 0)
	}
	if cdInit.ComponentReferences == nil {
		cdInit.ComponentReferences = make([]ocmV2.ComponentReference, 0)
	}
	builtByCLI, err := v1.NewLabel("kyma-project.io/built-by", "cli", v1.WithVersion("v1"))
	if err != nil {
		return nil, err
	}
	cdInit.Provider = "CLI"

	data, err := yaml.Marshal(cdInit)
	if err != nil {
		return nil, fmt.Errorf("unable to build archive for minimal descriptor: %w", err)
	}
	//data := fmt.Sprintf("component:\n  name: %s\n  version: %s\nmeta:\n  schemaVersion: v2\n", def.Name, def.Version)
	file, err := archiveFs.Create("component-descriptor.yaml")
	if err != nil {
		return nil, fmt.Errorf("unable to build archive for minimal descriptor: %w", err)
	}
	_, err = file.Write([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("unable to build archive for minimal descriptor: %w", err)
	}
	file.Close()
	ctx := cpi.DefaultContext()

	archive, err := comparch.New(
		ctx,
		accessobj.ACC_CREATE, archiveFs, nil,
		nil,
		vfs.ModePerm,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to build archive for minimal descriptor: %w", err)
	}
	cd := archive.GetDescriptor()
	cd.Metadata.ConfiguredVersion = def.SchemaVersion

	cd.Provider = v1.Provider{Name: "kyma-project.io", Labels: v1.Labels{*builtByCLI}}

	if isTargetDirAGitRepo {
		if err := addSources(cd, def, gitRemote); err != nil {
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
func addSources(cd *ocm.ComponentDescriptor, def *Definition, gitRemote string) error {
	if strings.HasSuffix(def.Source, ".git") {
		gitSource := gitsource.NewGitSource()
		var err error
		if def.Repo, err = gitSource.DetermineRepositoryURL(gitRemote, def.Repo, def.Source); err != nil {
			return err
		}
		src, err := gitSource.FetchSource(def.Source, def.Repo, def.Version)

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
