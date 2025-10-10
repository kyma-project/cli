package types

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidEnvFormat = errors.New("invalid env format, should be in format 'name=<NAME>,value=<VALUE>' or <NAME>=<VALUE>")
	ErrUnknownEnvField  = errors.New("unknown env field, supported fields are 'name' or 'value'")
)

type EnvMap struct {
	*Map
}

// SetValue sets the value of the NullableString from a string
func (e *EnvMap) SetValue(value *string) error {
	if value == nil {
		return nil
	}

	if e.Map == nil || e.Values == nil {
		e.Map = &Map{Values: map[string]interface{}{}}
	}

	if *value == "" {
		// input is empty, do nothing
		return nil
	}

	fields := strings.SplitN(*value, ",", 2)
	if len(fields) == 2 {
		// input is in the format name=NAME,key=KEY
		envName, envValue, err := parseStrictEnv(fields)
		if err != nil {
			return fmt.Errorf("failed to parse value '%s', %w", *value, err)
		}
		e.Values[envName] = envValue
		return nil
	}

	// input is in the format NAME=VALUE
	envElems := strings.Split(fields[0], "=")
	if len(envElems) != 2 {
		return fmt.Errorf("failed to parse value '%s': %w", *value, ErrInvalidEnvFormat)
	}
	e.Values[envElems[0]] = envElems[1]
	return nil
}

// parseStrictEnv parses env in strict format (like 'name=<NAME>,path=<PATH>')
// returns env name on first elem and its value on second elem
func parseStrictEnv(fields []string) (string, string, error) {
	envName := ""
	envValue := ""
	for _, v := range fields {
		elems := strings.SplitN(v, "=", 2)
		if len(elems) != 2 {
			// invalid format, '=' not found
			return "", "", ErrInvalidEnvFormat
		}

		switch elems[0] {
		case "name":
			envName = elems[1]
		case "value":
			envValue = elems[1]
		default:
			return "", "", ErrUnknownEnvField
		}
	}

	return envName, envValue, nil
}

// Set implements the flag.Value interface
func (e *EnvMap) Set(value string) error {
	if value == "" {
		return nil
	}

	return e.SetValue(&value)
}
