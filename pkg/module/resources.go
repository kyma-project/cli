package module

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/pkg/module/blob"
	"github.com/kyma-project/cli/pkg/module/kubebuilder"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	ocmv1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"
)

// ResourceDescriptor contains all information to describe a resource
type ResourceDescriptor struct {
	compdesc.Resource
	Input *blob.Input `json:"input,omitempty"`
}

// ResourceDescriptorList contains a list of all information to describe a resource.
type ResourceDescriptorList struct {
	Resources []ResourceDescriptor `json:"resources"`
}

// AddResources adds the resources in the given resource definitions into the archive and its FS.
// A resource definition is a string with format: NAME:TYPE@PATH, where NAME and TYPE can be omitted and will default to the last path element name and "helm-chart" respectively
func AddResources(
	archive *comparch.ComponentArchive,
	modDef *Definition,
	log *zap.SugaredLogger,
	fs vfs.FileSystem,
	registryCredSelector string,
) error {

	resources, err := generateResources(log, modDef.Version, registryCredSelector, modDef.Layers...)
	if err != nil {
		return err
	}

	log.Debugf("Adding %d resources...", len(resources))
	for i, resource := range resources {
		if resource.Input != nil {
			if err := addBlob(fs, archive, &resources[i]); err != nil {
				return err
			}
			log.Debugf("Added input blob from %q", resource.Input.Path)
		}
	}

	log.Debugf("Successfully added all resources to component descriptor")
	return nil
}

// generateResources generates resources by parsing the given definitions.
// Definitions have the following format: NAME:TYPE@PATH
// If a definition does not have a name or type, the name of the last path element is used,
// and it is assumed to be a helm-chart type.
func generateResources(log *zap.SugaredLogger, version, registryCredSelector string, defs ...Layer) ([]ResourceDescriptor, error) {
	var res []ResourceDescriptor
	credMatchLabels, err := CreateCredMatchLabels(registryCredSelector)
	if err != nil {
		return nil, err
	}
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

		if len(credMatchLabels) > 0 {
			r.SetLabels([]ocmv1.Label{{
				Name:  OCIRegistryCredLabel,
				Value: credMatchLabels,
			}})
		}

		log.Debugf("Generated resource:\n%s", r)
		res = append(res, r)
	}
	return res, nil
}

func addBlob(fs vfs.FileSystem, archive *comparch.ComponentArchive, resource *ResourceDescriptor) error {
	access, err := blob.AccessForFileOrFolder(fs, resource.Input)
	if err != nil {
		return err
	}

	blobAccess, err := archive.AddBlob(
		accessobj.CachedBlobAccessForDataAccess(archive.GetContext(), access.MimeType(), access), string(resource.Input.Type),
		resource.Resource.Name, nil,
	)
	if err != nil {
		return err
	}

	return archive.SetResource(&resource.ResourceMeta, blobAccess, cpi.ModifyResource(true))
}

func (rd ResourceDescriptor) String() string {
	y, err := yaml.Marshal(rd)
	if err != nil {
		return err.Error()
	}
	return string(y)
}

// Inspect updates the module definition provided as parameter with necessary data.
// Inspect supports a single source file path.
func Inspect(def *Definition, log *zap.SugaredLogger) error {
	log.Debugf("Inspecting module contents at [%s]:", def.Source)

	if err := def.validate(); err != nil {
		return err
	}

	// generated raw manifest -> layer 1
	def.Layers = append(def.Layers, Layer{
		name:         RawManifestLayerName,
		resourceType: TypeYaml,
		path:         def.SingleManifestPath,
	})

	// Add default CR if generating template
	var cr []byte
	if def.RegistryURL != "" {
		if def.DefaultCRPath != "" {
			var err error
			cr, err = os.ReadFile(def.DefaultCRPath)
			if err != nil {
				return fmt.Errorf("could not read CR file %q: %w", def.DefaultCRPath, err)
			}
		}
	}

	def.DefaultCR = cr

	return nil
}

// InspectLegacy analyzes the contents of a module and updates the module definition provided as parameter with all information contained in the module (layers, metadata and resources).
// InspectLegacy supports:
// Kubebuilder projects: if a PROJECT file is found and correctly parsed, the project will automatically be generated and layered.
// Custom module: If not a kubebuilder project, the user has complete freedom to layer the contents as desired via customDefs. Any contents of path not included in the customDefs will be added to the base layer
// Deprecated.
func InspectLegacy(def *Definition, customDefs []string, s step.Step, log *zap.SugaredLogger) error {
	log.Debugf("Inspecting module contents at [%s]:", def.Source)
	var layers []Layer

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

	// generated raw manifest -> layer 1
	renderedManifestPath, err := p.Build(def.Name)
	if err != nil {
		return err
	}

	// Add default CR if generating template
	var cr []byte
	if def.RegistryURL != "" {
		if def.DefaultCRPath == "" {
			cr, err = p.DefaultCR(s)
			if err != nil {
				return err
			}
		} else {
			cr, err = os.ReadFile(def.DefaultCRPath)
			if err != nil {
				return fmt.Errorf("could not read CR file %q: %w", def.DefaultCRPath, err)
			}
		}
	}

	def.Repo = p.Repo
	def.DefaultCR = cr
	def.Layers = append(def.Layers, Layer{
		name:         RawManifestLayerName,
		resourceType: TypeYaml,
		path:         renderedManifestPath,
	})
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
