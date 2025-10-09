package sourced

import (
	"errors"
	"strings"
)

type Env struct {
	// Env var Name
	Name string

	// Contains Configmap or Secret name or file path
	Location string

	// Single key extracted from Location. if empty it gets all keys
	LocationKey string

	// Envs prefix for case when LocationKey is not specified
	LocationKeysPrefix string
}

func (e *Env) String() string {
	if e.Name != "" {
		fullLocation := strings.Join([]string{e.Location, e.LocationKey}, ":")
		return e.Name + "=" + fullLocation
	}

	return strings.Join([]string{e.Location, e.LocationKeysPrefix}, ":")
}

func ParseEnv(value string) (Env, error) {
	values := strings.Split(value, ",")
	if len(values) > 0 {
		// strict format (like 'name=<NAME>,path=<PATH>')
		return parseStrictEnv(values)
	}

	// shorthand format (like '<PATH>' or '<PATH>:<PREFIX>' or '<NAME>=<PATH>' or '<NAME>=<PATH>:<KEY>')
	leftValue, rightValue := splitOnce(value, "=")
	if rightValue == "" {
		// no '=' found, assume left value is a filepath or resource (configmap/secret) name
		location, prefix := splitOnce(value, ":")
		return Env{
			Location:           location,
			LocationKeysPrefix: prefix,
		}, nil
	}

	// '=' found, assume left value is env var name and right value is a filepath or resource (configmap/secret) name
	location, key := splitOnce(rightValue, ":")
	return Env{
		Name:        leftValue,
		Location:    location,
		LocationKey: key,
	}, nil
}

// parseStrictEnv parses env in the strict format 'name=<NAME>,path=<PATH>'
func parseStrictEnv(values []string) (Env, error) {
	env := Env{}
	for _, v := range values {
		leftV, rightV := splitOnce(v, "=")
		if rightV == "" {
			// invalid format, '=' not found
			return Env{}, errors.New("invalid env format, should be in format '<FIELD_NAME>=<VALUE>,<FIELD_NAME>=<VALUE>,...'")
		}

		switch leftV {
		case "name":
			env.Name = rightV
		case "path", "resource":
			env.Location = rightV
		case "prefix":
			env.LocationKeysPrefix = rightV
		case "key":
			env.LocationKey = rightV
		default:
			return Env{}, errors.New("invalid env format, should be in format '<FIELD_NAME>=<VALUE>,<FIELD_NAME>=<VALUE>,...'")
		}
	}

	return env, nil
}

func splitOnce(value, sep string) (string, string) {
	parts := strings.Split(value, sep)
	if len(parts) == 1 {
		// no separator found
		return parts[0], ""
	}

	// join the rest back together
	afterSep := strings.Join(parts[1:], sep)

	return parts[0], afterSep
}
