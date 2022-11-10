package module

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/apis/v2/cdutils"
	cdvalidation "github.com/gardener/component-spec/bindings-go/apis/v2/validation"
	"github.com/gardener/component-spec/bindings-go/ctf"
	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/pkg/module/blob"
	"github.com/kyma-project/cli/pkg/module/kubebuilder"
	"github.com/kyma-project/cli/pkg/step"
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
func AddResources(archive *ctf.ComponentArchive, modDef *Definition, log *zap.SugaredLogger, fs vfs.FileSystem) error {
	resources, err := generateResources(log, modDef.Version, modDef.Layers...)
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
		if err := WriteComponentDescriptor(fs, archive.ComponentDescriptor, modDef.ArchivePath, ctf.ComponentDescriptorFileName); err != nil {
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
func generateResources(log *zap.SugaredLogger, version string, defs ...Layer) ([]ResourceDescriptor, error) {
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

// Inspect analyzes the contents of a module and updates the module definition provided as parameter with all information contained in the module (layers, metadata and resources).
// Inspect supports:
// Kubebuilder projects: if a PROJECT file is found and correctly parsed, the project will automatically be generated and layered.
// Custom module: If not a kubebuilder project, the user has complete freedom to layer the contents as desired via customDefs. Any contents of path not included in the customDefs will be added to the base layer
func Inspect(def *Definition, customDefs []string, s step.Step, log *zap.SugaredLogger) error {
	log.Debugf("Inspecting module contents at [%s]:", def.Source)
	layers := []Layer{}

	for _, d := range customDefs {
		rd, err := LayerFromString(d)
		if err != nil {
			return err
		}

		layers, err = appendDefIfValid(layers, rd, log)
		if err != nil {
			return err
		}
	}

	p, err := kubebuilder.ParseProject(def.Source)
	if err == nil {
		// Kubebuilder project
		log.Debug("Kubebuilder project detected.")
		return inspectProject(def, p, layers, s)

	} else if errors.Is(err, os.ErrNotExist) {
		// custom module
		log.Debug("No kubebuilder project detected, bundling module in a single layer.")
		return inspectCustom(def, layers, log)
	}
	// there is an error other than not exists
	return err
}

func inspectProject(def *Definition, p *kubebuilder.Project, layers []Layer, s step.Step) error {
	// use kubebuilder project name if no override given
	if def.Name == "" {
		def.Name = p.FullName()
	}
	if err := def.validate(); err != nil {
		return err
	}

	// generated chart -> layer 1
	chartPath, err := p.Build(def.Name, def.Version, def.RegistryURL) // TODO switch from charts to pure manifests when mod-mngr is ready
	if err != nil {
		return err
	}
	// config.yaml -> layer 2
	configPath, err := p.Config()
	if err != nil {
		return err
	}

	// Add default CR if generating template
	cr := []byte{}
	if def.RegistryURL != "" {
		if def.DefaultCRPath == "" {
			cr, err = p.DefaultCR(s)
			if err != nil {
				return err
			}
		} else {
			cr, err := os.ReadFile(def.DefaultCRPath)
			if err != nil {
				return fmt.Errorf("could not read CR file %q: %w", def.DefaultCRPath, err)
			}
			def.DefaultCR = cr
		}
	}

	charts := Layer{
		name:         filepath.Base(chartPath),
		resourceType: typeHelmChart,
		path:         chartPath,
	}
	config := Layer{
		name:         configLayerName,
		resourceType: typeYaml,
		path:         configPath,
	}

	def.Repo = p.Repo
	def.DefaultCR = cr
	def.Layers = append(def.Layers, charts, config)
	def.Layers = append(def.Layers, layers...)

	return nil
}

func inspectCustom(def *Definition, layers []Layer, log *zap.SugaredLogger) error {
	if err := def.validate(); err != nil {
		return err
	}
	absPath, err := filepath.Abs(def.Source)
	if err != nil {
		return fmt.Errorf("could not get absolute path to %q: %w", def.Source, err)
	}

	l := Layer{
		name:         filepath.Base(absPath),
		path:         absPath,
		resourceType: typeHelmChart,
	}
	// exclude any custom resources that overlap with module root to avoid bundling them twice
	for _, d := range layers {
		if strings.HasPrefix(d.path, l.path) {
			l.excludedFiles = append(l.excludedFiles, d.path)
		}
	}
	// prepend the base resource def
	base, err := appendDefIfValid(nil, l, log)
	if err != nil {
		return err
	}

	def.Layers = append(base, layers...)
	return nil
}

func appendDefIfValid(defs []Layer, r Layer, log *zap.SugaredLogger) ([]Layer, error) {
	info, err := os.Stat(r.path)
	if err != nil {
		return nil, fmt.Errorf("could not validate resource %q: %w", r.path, err)
	}

	if info.IsDir() {
		empty, err := files.IsDirEmpty(r.path)
		if err != nil {
			return nil, fmt.Errorf("could not determine if directory %q is empty: %w", r.path, err)
		}
		if empty {
			log.Debugf("Resource %q has no content, skipping", r.Path())
			return defs, nil
		}
	} else if info.Size() == 0 {
		log.Debugf("Resource %q has no content, skipping", r.Path())
		return defs, nil
	}
	log.Debugf("Added layer %+v", r)
	return append(defs, r), nil
}
