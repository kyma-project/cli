package types

import (
	"strings"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/types/sourced"
)

type SourcedEnvArray struct {
	Values []sourced.Env
}

func (e *SourcedEnvArray) String() string {
	values := make([]string, len(e.Values))
	for i, v := range e.Values {
		values[i] = v.String()
	}

	return strings.Join(values, ";")
}

// SetValue sets the value of the NullableString from a string
func (e *SourcedEnvArray) SetValue(value *string) error {
	if value == nil {
		return nil
	}

	env, err := sourced.ParseEnv(*value)
	if err != nil {
		return err
	}

	e.Values = append(e.Values, env)
	return nil
}

// Set implements the flag.Value interface
func (e *SourcedEnvArray) Set(value string) error {
	if value == "" {
		return nil
	}

	return e.SetValue(&value)
}

func (e *SourcedEnvArray) Type() string {
	return "stringArray"
}
