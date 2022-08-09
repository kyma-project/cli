package module

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/apis/v2/cdutils"
	cdvalidation "github.com/gardener/component-spec/bindings-go/apis/v2/validation"
	"github.com/gardener/component-spec/bindings-go/ctf"
	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/pkg/module/blob"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/yaml"
)

// ResourceDescriptor contains all information to describe a resource
type ResourceDescriptor struct {
	cdv2.Resource
	Input *blob.Input `json:"input,omitempty"`
}

// ResourceDescriptorList contains a list of all information to describe a resource.
type ResourceDescriptorList struct {
	Resources []ResourceDescriptor `json:"resources"`
}

// AddResources adds the resources in the given resource definitions into the archive and its FS.
// A resource definition is a string with format: NAME:TYPE@PATH, where NAME and TYPE can be omitted and will default to the last path element name and "helm-chart" respectively
func AddResources(archive *ctf.ComponentArchive, c *ComponentConfig, log *zap.SugaredLogger, fs vfs.FileSystem, defs ...ResourceDef) error {
	resources, err := generateResources(log, c.Version, defs...)
	if err != nil {
		return err
	}

	log.Debugf("Adding %d resources...", len(resources))
	for i, resource := range resources {
		if resource.Input != nil {
			log.Debugf("Added input blob from %q", resource.Input.Path)
			if err := addBlob(context.Background(), fs, archive, &resources[i]); err != nil {
				return err
			}
		} else {
			id := archive.ComponentDescriptor.GetResourceIndex(resource.Resource)
			if id != -1 {
				log.Debugf("Found existing resource in component descriptor, attempt merge...")
				mergedRes := cdutils.MergeResources(archive.ComponentDescriptor.Resources[id], resource.Resource)
				if errList := cdvalidation.ValidateResource(field.NewPath(""), mergedRes); len(errList) != 0 {
					return errList.ToAggregate()
				}
				archive.ComponentDescriptor.Resources[id] = mergedRes
			} else {
				if errList := cdvalidation.ValidateResource(field.NewPath(""), resource.Resource); len(errList) != 0 {
					return errList.ToAggregate()
				}
				archive.ComponentDescriptor.Resources = append(archive.ComponentDescriptor.Resources, resource.Resource)
			}
		}

		if err := cdvalidation.Validate(archive.ComponentDescriptor); err != nil {
			return fmt.Errorf("invalid component descriptor: %w", err)
		}
		if err := WriteComponentDescriptor(fs, archive.ComponentDescriptor, c.ComponentArchivePath, ctf.ComponentDescriptorFileName); err != nil {
			return err
		}
		log.Debugf("Successfully added resource to component descriptor")
	}
	log.Debugf("Successfully added all resources to component descriptor")
	return nil
}

func WriteComponentDescriptor(fs vfs.FileSystem, cd *cdv2.ComponentDescriptor, filePath string, fileName string) error {
	compDescFilePath := filepath.Join(filePath, fileName)
	data, err := yaml.Marshal(cd)
	if err != nil {
		return fmt.Errorf("unable to encode component descriptor: %w", err)
	}
	if err := vfs.WriteFile(fs, compDescFilePath, data, 0664); err != nil {
		return fmt.Errorf("unable to write modified comonent descriptor: %w", err)
	}
	return nil
}

// generateResources generates resources by parsing the given definitions.
// Definitions have the following format: NAME:TYPE@PATH
// If a definition does not have a name or type, the name of the last path element is used and it is assumed to be a helm-chart type.
func generateResources(log *zap.SugaredLogger, version string, defs ...ResourceDef) ([]ResourceDescriptor, error) {
	res := []ResourceDescriptor{}
	for _, d := range defs {
		r := ResourceDescriptor{Input: &blob.Input{}}

		r.Name = d.Name()
		r.Input.Path = d.Path()
		r.Type = d.Type()
		r.Version = version
		r.Relation = "local"
		dir, err := files.IsDir(r.Input.Path)
		if err != nil {
			return nil, errors.Wrap(err, "Could not determine if resource is a directory")
		}
		if dir {
			r.Input.Type = "dir"
			compress := true
			r.Input.CompressWithGzip = &compress
			r.Input.ExcludeFiles = d.excludedFiles
		} else {
			r.Input.Type = "file"
		}

		log.Debugf("Generated resource:\n%s", r)
		res = append(res, r)
	}
	return res, nil
}

func addBlob(ctx context.Context, fs vfs.FileSystem, archive *ctf.ComponentArchive, resource *ResourceDescriptor) error {
	b, err := resource.Input.Read(ctx, fs)
	if err != nil {
		return err
	}

	// default media type to binary data if nothing else is defined
	resource.Input.SetMediaTypeIfNotDefined(blob.MediaTypeOctetStream)

	err = archive.AddResource(&resource.Resource, ctf.BlobInfo{
		MediaType: resource.Input.MediaType,
		Digest:    b.Digest,
		Size:      b.Size,
	}, b.Reader)
	if err != nil {
		b.Reader.Close()
		return fmt.Errorf("unable to add input blob to archive: %w", err)
	}
	if err := b.Reader.Close(); err != nil {
		return fmt.Errorf("unable to close input file: %w", err)
	}
	return nil
}

func (rd ResourceDescriptor) String() string {
	y, err := yaml.Marshal(rd)
	if err != nil {
		return err.Error()
	}
	return string(y)
}

// builtInResources contains a set of resources that will automatically be created if found inspecting a module path
var builtInResources = map[string]ResourceDef{
	"config.yaml": {
		name:         "config",
		resourceType: "yaml",
	},
}

// Inspect analyzes the contents of a module and creates resource definitions for each separate layer the module should be split into
func Inspect(path string, log *zap.SugaredLogger) ([]ResourceDef, error) {
	log.Debugf("Inspecting module contents at [%s]:", path)
	// check path subfolders
	infos, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("could not Inspect module folder %q: %w", path, err)
	}

	defs := []ResourceDef{}
	// for each element in the mod path [charts, config.yaml] create a def and validate it
	for _, i := range infos {
		// special case for charts folder
		if i.Name() == "charts" {
			charts, err := os.ReadDir(filepath.Join(path, "charts"))
			if err != nil {
				return nil, err
			}

			for _, i := range charts {
				log.Debugf("Found chart %q", i.Name())
				r := ResourceDef{
					name:         i.Name(),
					path:         filepath.Join(path, "charts", i.Name()),
					resourceType: "helm-chart",
				}
				if defs, err = appendDefIfValid(defs, r, i, log); err != nil {
					return nil, err
				}

			}
		} else if r, ok := builtInResources[i.Name()]; ok {
			log.Debugf("Found built in layer %q", i.Name())

			r.path = filepath.Join(path, i.Name())
			if defs, err = appendDefIfValid(defs, r, i, log); err != nil {
				return nil, err
			}
		}
	}
	return defs, nil
}

func appendDefIfValid(defs []ResourceDef, r ResourceDef, i os.DirEntry, log *zap.SugaredLogger) ([]ResourceDef, error) {
	if i.IsDir() {
		empty, err := files.IsDirEmpty(r.path)
		if err != nil {
			return nil, fmt.Errorf("could not determine if directory %q is empty: %w", r.path, err)
		}
		if empty {
			log.Debugf("Resource %q has no content, skipping", i.Name())
			return defs, nil
		}
	}
	log.Debugf("Added layer %+v", r)
	return append(defs, r), nil
}
