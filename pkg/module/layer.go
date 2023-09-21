package module

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	RawManifestLayerName = "raw-manifest"
	typeHelmChart        = "helm-chart"
	TypeYaml             = "yaml"
)

// Layer encapsulates all necessary data to create an OCI layer
type Layer struct {
	name          string
	resourceType  string
	path          string
	excludedFiles []string
}

func NewLayer(name, resType, path string, excludedFiles ...string) *Layer {
	return &Layer{
		name:          name,
		resourceType:  resType,
		path:          path,
		excludedFiles: excludedFiles,
	}
}

// LayerFromString creates a layer resource from a string with format NAME:TYPE@PATH
// NAME and TYPE can be omitted and will default to the last path element name and "helm-chart" respectively.
func LayerFromString(def string) (Layer, error) {
	items := strings.FieldsFunc(def, func(r rune) bool { return r == ':' || r == '@' }) // split the elements of the format NAME:TYPE@PATH
	if len(items) == 0 || len(items) == 2 {
		return Layer{}, fmt.Errorf("the given resource %q could not be parsed. At least, it must contain a path and follow the format NAME:TYPE@PATH", def)
	}

	// only path given: infer name and use default type
	if len(items) == 1 {
		name := filepath.Base(items[0])
		name = strings.TrimSuffix(name, filepath.Ext(name)) // remove extension
		name = strings.ReplaceAll(name, ".", "-")           // replace dots in the name

		return Layer{
			name:         name,
			resourceType: typeHelmChart,
			path:         items[0],
		}, nil
	}

	return Layer{
		name:         items[0],
		resourceType: items[1],
		path:         items[2],
	}, nil
}

func (d Layer) Name() string {
	return d.name
}

func (d Layer) Type() string {
	return d.resourceType
}

func (d Layer) Path() string {
	return d.path
}

func (d Layer) ExcludedFiles() []string {
	return d.excludedFiles
}
