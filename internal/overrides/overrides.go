package overrides

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
)

var (
	supportedFileExt = []string{"yaml", "yml", "json"}
)

type interceptorOps string

const (
	interceptorOpsString    = "String"
	interceptorOpsIntercept = "Intercept"
)

// Overrides manages override merges
type Builder struct {
	files        []string
	overrides    []map[string]interface{}
	interceptors map[string]OverrideInterceptor
}

// AddFile adds overrides defined in a file to the builder
func (ob *Builder) AddFile(file string) error {
	for _, ext := range supportedFileExt {
		if strings.HasSuffix(file, fmt.Sprintf(".%s", ext)) {
			ob.files = append(ob.files, file)
			return nil
		}
	}
	return fmt.Errorf("Unsupported override file extension. Supported extensions are: %s", strings.Join(supportedFileExt, ", "))
}

// AddOverrides adds overrides for a chart to the builder
func (ob *Builder) AddOverrides(chart string, overrides map[string]interface{}) error {
	if chart == "" {
		return fmt.Errorf("Chart name cannot be empty when adding overrides")
	}
	if len(overrides) < 1 {
		return fmt.Errorf("Empty overrides map provided for chart '%s'", chart)
	}
	overridesMap := make(map[string]interface{})
	overridesMap[chart] = overrides
	ob.overrides = append(ob.overrides, overridesMap)
	return nil
}

// AddInterceptor registers an interceptor for particular override keys
func (ob *Builder) AddInterceptor(overrideKeys []string, interceptor OverrideInterceptor) {
	if ob.interceptors == nil {
		ob.interceptors = make(map[string]OverrideInterceptor)
	}
	for _, overrideKey := range overrideKeys {
		ob.interceptors[overrideKey] = interceptor
	}
}

// Build an overrides object merging all provided sources and applying interceptors
// WARNING: call this function sparingly, it runs all interceptors, potentially incurring heavy computations.
func (ob *Builder) Build() (Overrides, error) {
	o, err := ob.Raw()
	if err != nil {
		return Overrides{}, err
	}

	// assign intercepted overrides back to the original object to not loose the values
	o.overrides, err = o.intercept(interceptorOpsIntercept)
	return o, err
}

// Raw builds an overrides object contining only the raw values in the sources, without applying interceptors.
func (ob *Builder) Raw() (Overrides, error) {
	merged, err := ob.mergeSources()
	if err != nil {
		return Overrides{}, err
	}

	return Overrides{
		overrides:    merged,
		interceptors: ob.interceptors,
	}, nil
}

// mergeSources merges together all overrides sources into a single map
func (ob *Builder) mergeSources() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// merge files
	var fileOverrides map[string]interface{}
	for _, file := range ob.files {
		// read data
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		// unmarshal
		if strings.HasSuffix(file, ".json") {
			err = json.Unmarshal(data, &fileOverrides)
		} else {
			err = yaml.Unmarshal(data, &fileOverrides)
		}
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Failed to process configuration values defined in file '%s'", file))
		}
		// merge
		if err := mergo.Map(&result, fileOverrides, mergo.WithOverride); err != nil {
			return nil, err
		}
	}

	//merge overrides
	for _, override := range ob.overrides {
		if err := mergo.Map(&result, override, mergo.WithOverride); err != nil {
			return nil, err
		}
	}

	return result, nil
}

type Overrides struct {
	overrides    map[string]interface{}
	interceptors map[string]OverrideInterceptor
}

// Map returns a copy of the overrides in map form
func (o Overrides) Map() map[string]interface{} {
	return copyMap(o.overrides)
}

// String all provided overrides
func (o Overrides) String() string {
	in, err := o.intercept(interceptorOpsString)
	if err != nil {
		return fmt.Sprint(err)
	}
	return fmt.Sprintf("%v", in)
}

// Find returns a value on the overrides following a path separated by "." and if it exists.
func (o Overrides) Find(key string) (interface{}, bool) {
	return deepFind(o.overrides, strings.Split(key, "."))
}

func deepFind(m map[string]interface{}, path []string) (interface{}, bool) {
	// if reached the end of the path it means the searched value is a map itself
	// return the map
	if len(path) == 0 {
		return m, true
	}

	if v, ok := m[path[0]].(map[string]interface{}); ok {
		return deepFind(v, path[1:])
	}
	v, ok := m[path[0]]
	return v, ok
}

// setValue recursively traverses a map of maps and sets the given value in the given path separated by ".".
// Should the value type not be assignable to the path an error will be returned
func setValue(m map[string]interface{}, path []string, value interface{}) error {
	// if reached the end of the path it means the user is setting a map or path is wrong
	if len(path) == 0 {
		return errors.New("Error setting value, given key is a map")
	}

	if v, ok := m[path[0]].(map[string]interface{}); ok {
		return setValue(v, path[1:], value)
	}
	if len(path) > 1 {
		return fmt.Errorf("Error setting value, path is incorrect. Map expected at subkey %s but have %T", path[0], m[path[0]])
	}
	m[path[0]] = value
	return nil
}

// intercept runs all interceptors on the overrides and returns a copy of the overrides map with updated values
// The updated map can either be assigned back to the overrides to update the object optionally.
func (o Overrides) intercept(ops interceptorOps) (map[string]interface{}, error) {
	result := copyMap(o.overrides)

	for k, interceptor := range o.interceptors {
		if v, exists := o.Find(k); exists {
			if ops == interceptorOpsString {
				newVal := interceptor.String(v, k)
				if err := setValue(result, strings.Split(k, "."), newVal); err != nil {
					return nil, err
				}
			} else {
				newVal, err := interceptor.Intercept(v, k)
				if err != nil {
					return nil, err
				}
				if err := setValue(result, strings.Split(k, "."), newVal); err != nil {
					return nil, err
				}
			}
		} else {
			if err := interceptor.Undefined(result, k); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

func copyMap(m map[string]interface{}) map[string]interface{} {
	c := make(map[string]interface{}, len(m))
	for k, v := range m {
		if m2, ok := m[k].(map[string]interface{}); ok {
			c[k] = copyMap(m2)
		} else {
			c[k] = v
		}
	}
	return c
}
