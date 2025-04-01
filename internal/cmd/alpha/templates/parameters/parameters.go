package parameters

import (
	"os"

	templates_types "github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/types"
	cmdcommon_types "github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/spf13/pflag"
)

type Value interface {
	pflag.Value
	GetValue() interface{}
	GetPath() string
}

func NewTyped(paramType templates_types.CreateCustomFlagType, resourcepath string, defaultValue interface{}) Value {
	switch paramType {
	case templates_types.PathCustomFlagType:
		return newPathValue(resourcepath, defaultValue)
	case templates_types.IntCustomFlagType:
		return newInt64Value(resourcepath, defaultValue)
	case templates_types.BoolCustomFlagType:
		return newBoolValue(resourcepath, defaultValue)
	default:
		return newStringValue(resourcepath, defaultValue)
	}
}

type boolValue struct {
	*cmdcommon_types.NullableBool
	path string
}

func newBoolValue(path string, defaultValue interface{}) *boolValue {
	value := cmdcommon_types.NullableBool{}
	defaultBool, ok := sanitizeDefaultValue(defaultValue).(bool)
	if ok {
		value.Value = &defaultBool
	}
	return &boolValue{
		NullableBool: &value,
		path:         path,
	}
}

func (v *boolValue) GetValue() interface{} {
	if v.Value == nil {
		return nil
	}

	return *v.Value
}

func (sv *boolValue) GetPath() string {
	return sv.path
}

type int64Value struct {
	*cmdcommon_types.NullableInt64
	path string
}

func newInt64Value(path string, defaultValue interface{}) *int64Value {
	value := cmdcommon_types.NullableInt64{}
	defaultInt64, ok := sanitizeDefaultValue(defaultValue).(int64)
	if ok {
		value.Value = &defaultInt64
	}

	return &int64Value{
		NullableInt64: &value,
		path:          path,
	}
}

func (v *int64Value) GetValue() interface{} {
	if v.Value == nil {
		return nil
	}

	return *v.Value
}

func (v *int64Value) GetPath() string {
	return v.path
}

type stringValue struct {
	*cmdcommon_types.NullableString
	path string
}

func newStringValue(path string, defaultValue interface{}) *stringValue {
	value := cmdcommon_types.NullableString{}
	defaultString, ok := sanitizeDefaultValue(defaultValue).(string)
	if ok {
		value.Value = &defaultString
	}
	return &stringValue{
		NullableString: &value,
		path:           path,
	}
}

func (v *stringValue) GetValue() interface{} {
	if v.Value == nil {
		return nil
	}

	return *v.Value
}

func (sv *stringValue) GetPath() string {
	return sv.path
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
