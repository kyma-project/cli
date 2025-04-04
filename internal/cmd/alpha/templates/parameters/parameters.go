package parameters

import (
	"os"

	cmdcommon_types "github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/spf13/pflag"
)

type Value interface {
	pflag.Value
	GetValue() interface{}
	GetPath() string
}

type boolValue struct {
	cmdcommon_types.NullableBool
	path string
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
	cmdcommon_types.NullableInt64
	path string
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
	cmdcommon_types.NullableString
	path string
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
	stringValue
}

func (pv *pathValue) Set(value string) error {
	bytes, err := os.ReadFile(value)
	if err != nil {
		return err
	}

	return pv.stringValue.Set(string(bytes))
}
