package reflect

import (
	"fmt"
	"reflect"
)

// isSet - returns false if value is nil or zero-value otherwise returns true
func isSet(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.UnsafePointer, reflect.Func, reflect.Slice,
		reflect.Map:
		return !v.IsNil()
	default:
		return !v.IsZero()
	}
}

func NoneOf(v interface{}, names []string) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("%s", v)
		}
	}()

	val := reflect.ValueOf(v)
	for _, name := range names {
		field := val.FieldByName(name)
		if isSet(field) {
			return fmt.Errorf("field: '%s' is set", name)
		}
	}
	return
}
