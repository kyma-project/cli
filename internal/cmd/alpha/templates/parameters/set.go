package parameters

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Set(obj map[string]interface{}, values []Value) clierror.Error {
	for _, extraValue := range values {
		value := extraValue.GetValue()
		if value == nil {
			// value is not set and has no default value
			continue
		}

		fields := splitPath(extraValue.GetPath())
		subObj, err := buildExtraValuesObject(value, fields...)
		if err != nil {
			return clierror.Wrap(err, clierror.New(
				fmt.Sprintf("failed to build value %v for path %s", value, extraValue.GetPath()),
			))
		}

		err = mergeObjects(subObj, obj)
		if err != nil {
			return clierror.Wrap(err, clierror.New(
				fmt.Sprintf("failed to set value %v for path %s", value, extraValue.GetPath()),
			))
		}
	}

	return nil
}

func buildExtraValuesObject(value interface{}, fields ...string) (map[string]interface{}, error) {
	obj := map[string]interface{}{}
	for i := range fields {
		if len(fields) == i+1 {
			// is last elem
			return obj, unstructured.SetNestedField(obj, value, fields...)
		}

		if strings.HasSuffix(fields[i], "]") && strings.Contains(fields[i], "[") {
			// is slice
			sliceIter, err := getSliceFieldIterator(fields[i])
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get slice element number from field %s", fields[i])
			}

			fields[i] = trimSliceFieldSuffix(fields[i])
			subObj, err := buildExtraValuesObject(value, fields[i+1:]...)
			if err != nil {
				return nil, err
			}

			slice := make([]interface{}, sliceIter+1)
			slice[sliceIter] = subObj
			err = unstructured.SetNestedSlice(obj, slice, fields[:i+1]...)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to set slice value for path .%s", strings.Join(fields[:i], "."))
			}
			return obj, err
		}

		err := unstructured.SetNestedMap(obj, map[string]interface{}{}, fields[:i+1]...)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to set value for path .%s", strings.Join(fields[:i], "."))
		}
	}

	return obj, nil
}

func getSliceFieldIterator(field string) (int, error) {
	stringIter := field[strings.Index(field, "[")+1 : len(field)-1]
	if stringIter == "" {
		// is []
		return 0, nil
	}

	iter, err := strconv.ParseInt(stringIter, 10, 0)
	return int(iter), err
}

func trimSliceFieldSuffix(field string) string {
	return field[:strings.Index(field, "[")]
}

func mergeObjects(from map[string]interface{}, to map[string]interface{}) error {
	for key, val := range from {
		toVal, ok := to[key]
		if !ok {
			// key not found
			// insert whole map
			to[key] = val
			return nil
		}

		err := hasSameTypes(val, toVal)
		if err != nil {
			// existing field in other type than expeced one
			return errors.Wrapf(err, "fields have different types for key %s", key)
		}

		switch val.(type) {
		case map[string]interface{}:
			return mergeObjects(val.(map[string]interface{}), to[key].(map[string]interface{}))
		case []interface{}:
			var err error
			to[key], err = mergeSlices(val.([]interface{}), to[key].([]interface{}))
			return err
		default:
			// is simple type like int64, string, bool...
			to[key] = val

			return nil
		}
	}
	return nil
}

func mergeSlices(from, to []interface{}) ([]interface{}, error) {
	dest := to
	for i := range from {
		if len(dest) < i+1 {
			// desired slice is smaller so I add elem
			dest = append(dest, from[i])
			continue
		}

		var err error
		switch from[i].(type) {
		case map[string]interface{}:
			err = mergeObjects(from[i].(map[string]interface{}), dest[i].(map[string]interface{}))
		case []interface{}:
			dest[i], err = mergeSlices(from[i].([]interface{}), dest[i].([]interface{}))
		case nil:
			// in this case we want keep value from dest slice
			continue
		default:
			// is simple type like int64, string, bool...
			dest[i] = from[i]
		}

		if err != nil {
			return nil, err
		}
	}

	return dest, nil
}

func hasSameTypes(a, b interface{}) error {
	aKind := reflect.TypeOf(a).Kind()
	bKind := reflect.TypeOf(b).Kind()
	if aKind != bKind {
		// existing field in other type than expeced one
		return fmt.Errorf("type %s other than expected %s", aKind.String(), bKind.String())
	}

	return nil
}

func splitPath(path string) []string {
	return strings.Split(
		// remove optional dot at the beginning of the path
		strings.TrimPrefix(path, "."),
		".",
	)
}
