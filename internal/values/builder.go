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
	interceptors map[string]interceptor
}

func (b *builder) addValuesFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.Wrap(err, "invalid override file: not exists")
	}

	for _, ext := range supportedFileExt {
		if strings.HasSuffix(filePath, fmt.Sprintf(".%s", ext)) {
			b.files = append(b.files, filePath)
			return nil
		}
	}
	return fmt.Errorf("Unsupported override file extension. Supported extensions are: %s", strings.Join(supportedFileExt, ", "))
}

func (b *builder) addValues(values map[string]interface{}) error {
	if len(values) < 1 {
		return fmt.Errorf("invalid values: empty")
	}

	b.values = append(b.values, values)
	return nil
}

func (b *builder) addInterceptor(overrideKeys []string, i interceptor) {
	if b.interceptors == nil {
		b.interceptors = make(map[string]interceptor)
	}
	for _, overrideKey := range overrideKeys {
		b.interceptors[overrideKey] = i
	}
}

func (b *builder) build() (buildResult, error) {
	o, err := b.raw()
	if err != nil {
		return buildResult{}, err
	}

	// assign intercepted values back to the original object to not lose the values
	o.values, err = o.intercept(interceptorOpsIntercept)
	return o, err
}

func (b *builder) raw() (buildResult, error) {
	merged, err := b.mergeSources()
	if err != nil {
		return buildResult{}, err
	}

	return buildResult{
		values:       merged,
		interceptors: b.interceptors,
	}, nil
}

func (b *builder) mergeSources() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// merge files
	var fileOverrides map[string]interface{}
	for _, file := range b.files {
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
	for _, override := range b.values {
		if err := mergo.Map(&result, override, mergo.WithOverride); err != nil {
			return nil, err
		}
	}

	return result, nil
}

type buildResult struct {
	values       map[string]interface{}
	interceptors map[string]interceptor
}

func (r buildResult) toMap() map[string]interface{} {
	return copyMap(r.values)
}

// toFlattenedMap returns a copy of the values in flattened map form (nested keys are merged separated by dots)
func (r buildResult) toFlattenedMap() map[string]interface{} {
	return flattenValues(r.toMap())
}

func flattenValues(overrides map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for outerKey, outerValue := range overrides {
		if valueAsMap, ok := outerValue.(map[string]interface{}); ok {
			mapWithIncompleteKeys := flattenValues(valueAsMap)
			for innerKey, innerValue := range mapWithIncompleteKeys {
				result[outerKey+"."+innerKey] = innerValue
			}
		} else {
			result[outerKey] = outerValue
		}
	}

	return result
}

func (r buildResult) find(key string) (interface{}, bool) {
	return deepFind(r.values, strings.Split(key, "."))
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

func (r buildResult) intercept(ops interceptorOps) (map[string]interface{}, error) {
	result := copyMap(r.values)

	for k, interceptor := range r.interceptors {
		if v, exists := r.find(k); exists {
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
