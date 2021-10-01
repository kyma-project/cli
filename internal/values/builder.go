package values

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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

type builder struct {
	files        []string
	values       []map[string]interface{}
	interceptors map[string]valueInterceptor
}

func (ob *builder) addValuesFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.Wrap(err, "invalid override file: not exists")
	}

	for _, ext := range supportedFileExt {
		if strings.HasSuffix(filePath, fmt.Sprintf(".%s", ext)) {
			ob.files = append(ob.files, filePath)
			return nil
		}
	}
	return fmt.Errorf("Unsupported override file extension. Supported extensions are: %s", strings.Join(supportedFileExt, ", "))
}

func (ob *builder) addValues(values map[string]interface{}) error {
	if len(values) < 1 {
		return fmt.Errorf("invalid values: empty")
	}

	ob.values = append(ob.values, values)
	return nil
}

func (ob *builder) addInterceptor(overrideKeys []string, interceptor valueInterceptor) {
	if ob.interceptors == nil {
		ob.interceptors = make(map[string]valueInterceptor)
	}
	for _, overrideKey := range overrideKeys {
		ob.interceptors[overrideKey] = interceptor
	}
}

func (ob *builder) build() (Overrides, error) {
	o, err := ob.raw()
	if err != nil {
		return Overrides{}, err
	}

	// assign intercepted values back to the original object to not loose the values
	o.overrides, err = o.intercept(interceptorOpsIntercept)
	return o, err
}

func (ob *builder) raw() (Overrides, error) {
	merged, err := ob.mergeSources()
	if err != nil {
		return Overrides{}, err
	}

	return Overrides{
		overrides:    merged,
		interceptors: ob.interceptors,
	}, nil
}

func (ob *builder) mergeSources() (map[string]interface{}, error) {
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

	//merge values
	for _, override := range ob.values {
		if err := mergo.Map(&result, override, mergo.WithOverride); err != nil {
			return nil, err
		}
	}

	return result, nil
}

type Overrides struct {
	overrides    map[string]interface{}
	interceptors map[string]valueInterceptor
}

// Map returns a copy of the values in map form
func (o Overrides) Map() map[string]interface{} {
	return copyMap(o.overrides)
}

// FlattenedMap returns a copy of the values in flattened map form (nested keys are merged separated by dots)
func (o Overrides) FlattenedMap() map[string]interface{} {
	return flattenOverrides(o.Map())
}

func flattenOverrides(overrides map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for outerKey, outerValue := range overrides {
		if valueAsMap, ok := outerValue.(map[string]interface{}); ok {
			mapWithIncompleteKeys := flattenOverrides(valueAsMap)
			for innerKey, innerValue := range mapWithIncompleteKeys {
				result[outerKey+"."+innerKey] = innerValue
			}
		} else {
			result[outerKey] = outerValue
		}
	}

	return result
}

// String all provided values
func (o Overrides) String() string {
	in, err := o.intercept(interceptorOpsString)
	if err != nil {
		return fmt.Sprint(err)
	}
	return fmt.Sprintf("%v", in)
}

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

func setValue(m map[string]interface{}, path []string, value interface{}) error {
	// if reached the end of the path it means the user is setting a map or path is wrong
	if len(path) == 0 {
		return errors.New("Error setting value, givenOverrides key is a map")
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
