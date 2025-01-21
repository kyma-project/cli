package parameters

import (
	"os"
	"strconv"

	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/types"
	"github.com/spf13/pflag"
)

type Value interface {
	pflag.Value
	GetValue() interface{}
	GetPath() string
}

func NewTyped(paramType types.CreateCustomFlagType, resourcepath string, defaultValue interface{}) Value {
	switch paramType {
	case types.StringCustomFlagType:
		return newStringValue(resourcepath, defaultValue)
	case types.PathCustomFlagType:
		return newPathValue(resourcepath, defaultValue)
	case types.IntCustomFlagType:
		return newInt64Value(resourcepath, defaultValue)
	default:
		// flag type is not supported
		return nil
	}
}

type int64Value struct {
	path  string
	value *int64
}

func newInt64Value(path string, defaultValue interface{}) *int64Value {
	var value *int64
	defaultInt64, ok := sanitizeDefaultValue(defaultValue).(int64)
	if ok {
		value = &defaultInt64
	}
	return &int64Value{
		value: value,
		path:  path,
	}
}

func (iv *int64Value) GetValue() interface{} {
	if iv.value == nil {
		return nil
	}

	return *iv.value
}

func (iv *int64Value) GetPath() string {
	return iv.path
}

func (iv *int64Value) String() string {
	if iv.value == nil {
		return ""
	}

	return strconv.FormatInt(*iv.value, 10)
}

func (iv *int64Value) Set(value string) error {
	if value == "" {
		return nil
	}

	v, err := strconv.ParseInt(value, 10, 0)
	if err != nil {
		return err
	}

	iv.value = &v
	return nil
}

func (iv *int64Value) Type() string {
	return "int64"
}

type stringValue struct {
	path  string
	value *string
}

func newStringValue(path string, defaultValue interface{}) *stringValue {
	var value *string
	defaultString, ok := sanitizeDefaultValue(defaultValue).(string)
	if ok {
		value = &defaultString
	}
	return &stringValue{
		value: value,
		path:  path,
	}
}

func (sv *stringValue) GetValue() interface{} {
	if sv.value == nil {
		return nil
	}

	return *sv.value
}

func (sv *stringValue) GetPath() string {
	return sv.path
}

func (sv *stringValue) String() string {
	if sv.value == nil {
		return ""
	}

	return *sv.value
}

func (sv *stringValue) Set(value string) error {
	if value != "" {
		sv.value = &value
	}

	return nil
}

func (sv *stringValue) Type() string {
	return "string"
}

type pathValue struct {
	*stringValue
}

func newPathValue(path string, defaultValue interface{}) *pathValue {
	return &pathValue{
		stringValue: newStringValue(path, defaultValue),
	}
}

func (pv *pathValue) Set(value string) error {
	bytes, err := os.ReadFile(value)
	if err != nil {
		return err
	}

	return pv.stringValue.Set(string(bytes))
}
