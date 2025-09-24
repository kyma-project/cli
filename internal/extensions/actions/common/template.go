package common

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/extensions/errors"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
)

var funcMap = template.FuncMap{
	"newLineIndent": newLineIndent,
	"toEnvs":        toEnvs,
	"toArray":       toArray,
	"toYaml":        toYaml,
	"wasUsed":       wasUsed,
}

// templateConfig parses the given template and executes it with the provided overwrites
func templateConfig(tmpl []byte, overwrites types.ActionConfigOverwrites) ([]byte, clierror.Error) {
	configTmpl, err := template.
		New("config").
		Option("missingkey=zero").
		Delims("${{", "}}").Funcs(funcMap).
		Parse(string(tmpl))
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("failed to parse config template"))
	}

	templatedConfig := bytes.NewBuffer([]byte{})
	err = configTmpl.
		Execute(templatedConfig, overwrites)
	if err != nil {
		return nil, clierror.New(err.Error())
	}

	return templatedConfig.Bytes(), nil
}

// adds indentation to the beginning of each new line
func newLineIndent(n int, s string) string {
	return strings.ReplaceAll(s, "\n", fmt.Sprintf("\n%s", strings.Repeat(" ", n)))
}

// toEnvs converts a map of environment variables to a YAML array string
func toEnvs(val map[string]interface{}) string {
	envs := []string{}
	for k, v := range val {
		envs = append(envs, fmt.Sprintf(`{"name":"%s","value":"%s"}`, k, v))
	}

	return fmt.Sprintf(`[%s]`, strings.Join(envs, ","))
}

// toArray converts a map to a YAML array string using the provided format
func toArray(format string, val map[string]interface{}) (string, error) {
	fields := []string{}
	for k, v := range val {
		tmpl, err := template.New("array").Parse(format)
		if err != nil {
			return "", errors.Wrap(err, "failed to parse array template")
		}

		templatedConfig := bytes.NewBuffer([]byte{})
		err = tmpl.Execute(templatedConfig, map[string]interface{}{"key": k, "value": v})
		if err != nil {
			return "", errors.Wrap(err, "failed to execute array template")
		}
		fields = append(fields, templatedConfig.String())
	}

	return fmt.Sprintf(`[%s]`, strings.Join(fields, ",")), nil
}

// toYaml converts a map to a YAML object string
func toYaml(val map[string]interface{}) string {
	fields := []string{}
	for k, v := range val {
		fields = append(fields, fmt.Sprintf(`"%s":"%s"`, k, v))
	}
	return fmt.Sprintf("{%s}", strings.Join(fields, ","))
}

// wasUsed checks if the last argument (flag) is nil (was used) and returns appropriate value.
func wasUsed(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return "", errors.New("ifNil requires at least two arguments")
	}
	// last argument is the flag used and semi-last is the value to return if nil
	if args[len(args)-1] == nil {
		return args[len(args)-2], nil
	}

	flagValue := args[len(args)-1]
	// if last argument(flag) is not nil
	switch v := flagValue.(type) {
	case bool:
		if len(args) == 3 { // notNil, nil, flag
			return args[0], nil // nil handled by ifNilBool func
		} else if len(args) == 4 { // true, false, nil, flag
			if v {
				return args[0], nil
			}
			return args[1], nil
		}
	default: // covers string, int, map and path flag types
		if len(args) != 2 {
			return "", errors.New(fmt.Sprintf("ifNil requires exactly two arguments for type %T", v))
		}
		return args[0], nil
	}
	return "", errors.New("ifNil requires at least three arguments for type bool")
}
