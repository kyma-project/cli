package common

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
)

var funcMap = template.FuncMap{
	"newLineIndent": newLineIndent,
	"toEnvs":        toEnvs,
}

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

func newLineIndent(n int, s string) string {
	return strings.ReplaceAll(s, "\n", fmt.Sprintf("\n%s", strings.Repeat(" ", n)))
}

func toEnvs(val interface{}) string {

	return ""
}
