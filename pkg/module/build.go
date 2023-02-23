package module

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	cdvalidation "github.com/gardener/component-spec/bindings-go/apis/v2/validation"
	"github.com/gardener/component-spec/bindings-go/codec"
	"github.com/gardener/component-spec/bindings-go/ctf"
	"github.com/kyma-project/cli/pkg/module/git"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"regexp"
	"sigs.k8s.io/yaml"
)

// Build creates a component archive with the given configuration
func Build(fs vfs.FileSystem, def *Definition) (*ctf.ComponentArchive, error) {
	if err := def.validate(); err != nil {
		return nil, err
	}

	compDescFilePath := filepath.Join(def.ArchivePath, ctf.ComponentDescriptorFileName)
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
	return buildFull(fs, def, compDescFilePath)
}

// buildWithoutOverwriting builds over an existing descriptor without overwriting
func buildWithoutOverwriting(fs vfs.FileSystem, def *Definition) (*ctf.ComponentArchive, error) {
	// add the input to the ctf format
	archiveFs, err := projectionfs.New(fs, def.ArchivePath)
	if err != nil {
		return nil, fmt.Errorf("unable to create projectionfilesystem: %w", err)
	}

	archive, err := ctf.NewComponentArchiveFromFilesystem(archiveFs, codec.DisableValidation(true))
	if err != nil {
		return nil, fmt.Errorf("unable to parse component archive from %s: %w", def.ArchivePath, err)
	}

	cd := archive.ComponentDescriptor

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

	if err = cdvalidation.Validate(cd); err != nil {
		return nil, fmt.Errorf("invalid component descriptor: %w", err)
	}

	return archive, nil
}

func buildFull(fs vfs.FileSystem, def *Definition, compDescFilePath string) (*ctf.ComponentArchive, error) {
	// build minimal archive

	if err := fs.MkdirAll(def.ArchivePath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create component-archive path %q: %w", def.ArchivePath, err)
	}
	archiveFs, err := projectionfs.New(fs, def.ArchivePath)
	if err != nil {
		return nil, fmt.Errorf("unable to create projectionfilesystem: %w", err)
	}

	cd := &cdv2.ComponentDescriptor{}
	if err := addSources(cd, def); err != nil {
		return nil, err
	}
	cd.Metadata.Version = cdv2.SchemaVersion
	cd.ComponentSpec.Name = def.Name
	cd.ComponentSpec.Version = def.Version
	cd.Provider = cdv2.InternalProvider
	cd.RepositoryContexts = make([]*cdv2.UnstructuredTypedObject, 0)
	if len(def.RegistryURL) != 0 {
		repoCtx, err := cdv2.NewUnstructured(BuildNewOCIRegistryRepository(def.RegistryURL, cdv2.ComponentNameMapping(def.NameMappingMode)))
		if err != nil {
			return nil, fmt.Errorf("unable to create repository context: %w", err)
		}
		cd.RepositoryContexts = []*cdv2.UnstructuredTypedObject{&repoCtx}
	}
	if err := cdv2.DefaultComponent(cd); err != nil {
		return nil, fmt.Errorf("unable to default component descriptor: %w", err)
	}

	if err := cdvalidation.Validate(cd); err != nil {
		return nil, fmt.Errorf("unable to validate component descriptor: %w", err)
	}

	data, err := yaml.Marshal(cd)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal component descriptor: %w", err)
	}
	if err := vfs.WriteFile(fs, compDescFilePath, data, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to write component descriptor to %s: %w", compDescFilePath, err)
	}

	return ctf.NewComponentArchive(cd, archiveFs), nil
}

func addSources(cd *cdv2.ComponentDescriptor, def *Definition) error {
	src, err := git.Source(def.Source, def.Repo, def.Version)
	if err != nil {
		return err
	}
	if src != nil {
		cd.Sources = append(cd.Sources, *src)
	}
	return nil
}

func BuildNewOCIRegistryRepository(registry string, mapping cdv2.ComponentNameMapping) *cdv2.OCIRegistryRepository {
	return cdv2.NewOCIRegistryRepository(noSchemeURL(registry), mapping)
}

func noSchemeURL(url string) string {
	regex := regexp.MustCompile(`^https?://`)
	return regex.ReplaceAllString(url, "")
}
