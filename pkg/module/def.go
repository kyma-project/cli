package module

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	DefaultResourceType = "helm-chart"
)

// ResourceDef represents a resource definition that can be added to a module as a layer
type ResourceDef struct {
	name          string
	resourceType  string
	path          string
	excludedFiles []string
}

// ResourceDefFromString creates a resource definition from a string with format NAME:TYPE@PATH
// NAME and TYPE can be omitted and will default to the last path element name and "helm-chart" respectively.
func ResourceDefFromString(def string) (ResourceDef, error) {
	items := strings.FieldsFunc(def, func(r rune) bool { return r == ':' || r == '@' }) // split the elements of the format NAME:TYPE@PATH
	if len(items) == 0 || len(items) == 2 {
		return ResourceDef{}, fmt.Errorf("the given resource %q could not be parsed. At least, it must contain a path and follow the format NAME:TYPE@PATH", def)
	}

	// only path given: infer name and use default type
	if len(items) == 1 {
		name := filepath.Base(items[0])
		name = strings.TrimSuffix(name, filepath.Ext(name)) // remove extension
		name = strings.ReplaceAll(name, ".", "-")           // replace dots in the name

		return ResourceDef{
			name:         name,
			resourceType: DefaultResourceType,
			path:         items[0],
		}, nil
	}

	return ResourceDef{
		name:         items[0],
		resourceType: items[1],
		path:         items[2],
	}, nil
}

func (d ResourceDef) Name() string {
	return d.name
}

func (d ResourceDef) Type() string {
	return d.resourceType
}

func (d ResourceDef) Path() string {
	return d.path
}
