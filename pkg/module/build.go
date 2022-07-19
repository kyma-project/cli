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
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"sigs.k8s.io/yaml"
)

// ComponentConfig contains all configurable fields for a component descriptor
type ComponentConfig struct {
	ComponentArchivePath string // Location of the component descriptor. If it does not exist, it is created.
	Name                 string // Name of the module (mandatory)
	Version              string // Version of the module (mandatory)
	RegistryURL          string // Registry URL to push the image to (optional)
	Overwrite            bool   // If true, existing module is overwritten if the configuration differs.
}

// Build creates a component archive with the given configuration
func Build(fs vfs.FileSystem, c *ComponentConfig) (*ctf.ComponentArchive, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	compDescFilePath := filepath.Join(c.ComponentArchivePath, ctf.ComponentDescriptorFileName)
	if !c.Overwrite {
		_, err := fs.Stat(compDescFilePath)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if err == nil {
			// add the input to the ctf format
			archiveFs, err := projectionfs.New(fs, c.ComponentArchivePath)
			if err != nil {
				return nil, fmt.Errorf("unable to create projectionfilesystem: %w", err)
			}

			archive, err := ctf.NewComponentArchiveFromFilesystem(archiveFs, codec.DisableValidation(true))
			if err != nil {
				return nil, fmt.Errorf("unable to parse component archive from %s: %w", c.ComponentArchivePath, err)
			}

			cd := archive.ComponentDescriptor

			if c.Name != "" {
				if cd.Name != "" && cd.Name != c.Name {
					return nil, errors.New("unable to overwrite the existing component name: forbidden")
				}
				cd.Name = c.Name
			}

			if c.Version != "" {
				if cd.Version != "" && cd.Version != c.Version {
					return nil, errors.New("unable to overwrite the existing component version: forbidden")
				}
				cd.Version = c.Version
			}

			if err = cdvalidation.Validate(cd); err != nil {
				return nil, fmt.Errorf("invalid component descriptor: %w", err)
			}

			return archive, nil
		}
	}

	// build minimal archive

	if err := fs.MkdirAll(c.ComponentArchivePath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create component-archive path %q: %w", c.ComponentArchivePath, err)
	}
	archiveFs, err := projectionfs.New(fs, c.ComponentArchivePath)
	if err != nil {
		return nil, fmt.Errorf("unable to create projectionfilesystem: %w", err)
	}

	cd := &cdv2.ComponentDescriptor{}
	cd.Metadata.Version = cdv2.SchemaVersion
	cd.ComponentSpec.Name = c.Name
	cd.ComponentSpec.Version = c.Version
	cd.Provider = cdv2.InternalProvider
	cd.RepositoryContexts = make([]*cdv2.UnstructuredTypedObject, 0)
	if len(c.RegistryURL) != 0 {
		repoCtx, err := cdv2.NewUnstructured(cdv2.NewOCIRegistryRepository(c.RegistryURL, cdv2.OCIRegistryURLPathMapping))
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

func (cfg *ComponentConfig) validate() error {
	if cfg.Name == "" {
		return errors.New("The module name cannot be empty")
	}
	if cfg.Version == "" {
		return errors.New("The module version cannot be empty")
	}
	if cfg.ComponentArchivePath == "" {
		return errors.New("The module version cannot be empty")
	}
	return nil
}
