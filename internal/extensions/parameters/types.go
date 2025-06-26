package parameters

import (
	"os"

	cmdcommontypes "github.com/kyma-project/cli.v3/internal/cmdcommon/types"
)

type ConfigFieldType string

const (
	StringCustomType ConfigFieldType = "string"
	PathCustomType   ConfigFieldType = "path"
	IntCustomType    ConfigFieldType = "int"
	BoolCustomType   ConfigFieldType = "bool"
	MapCustomType    ConfigFieldType = "map"
)

var (
	ValidTypes = []ConfigFieldType{
		StringCustomType,
		PathCustomType,
		IntCustomType,
		BoolCustomType,
		MapCustomType,
	}
)

type mapValue struct {
	cmdcommontypes.Map
	path string
}

func (v *mapValue) GetValue() interface{} {
	return v.Values
}

func (v *mapValue) GetPath() string {
	return v.path
}

func (v *mapValue) SetValue(value string) error {
	return v.Set(value)
}

type boolValue struct {
	cmdcommontypes.NullableBool
	path string
}

func (v *boolValue) GetValue() interface{} {
	return getValue(v.Value)
}

func (v *boolValue) GetPath() string {
	return v.path
}

func (v *boolValue) SetValue(value string) error {
	return v.Set(value)
}

type int64Value struct {
	cmdcommontypes.NullableInt64
	path string
}

func (v *int64Value) GetValue() interface{} {
	return getValue(v.Value)
}

func (v *int64Value) GetPath() string {
	return v.path
}

func (v *int64Value) SetValue(value string) error {
	return v.Set(value)
}

type stringValue struct {
	cmdcommontypes.NullableString
	path string
}

func (v *stringValue) GetValue() interface{} {
	return getValue(v.Value)
}

func (sv *stringValue) GetPath() string {
	return sv.path
}

func (sv *stringValue) SetValue(value string) error {
	return sv.Set(value)
}

type pathValue struct {
	stringValue
}

func (pv *pathValue) Set(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// set value even when file is empty
	newValue := string(bytes)
	pv.Value = &newValue

	return nil
}

func (pv *pathValue) SetValue(value string) error {
	return pv.stringValue.Set(value)
}

func getValue[T any](value *T) interface{} {
	if value != nil {
		return *value
	}

	var emptyValue T
	return emptyValue
}
