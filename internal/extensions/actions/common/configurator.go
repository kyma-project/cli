package common

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"gopkg.in/yaml.v3"
)

// TemplateConfigurator is a struct that implements the Configure for the types.Action interface.
// It is used to configure an action using go tempaltes.
type TemplateConfigurator[T any] struct {
	Cfg T
}

func (c *TemplateConfigurator[T]) Configure(cfgTmpl types.ActionConfig, overwrites types.ActionConfigOverwrites) clierror.Error {
	tmplBytes, err := yaml.Marshal(cfgTmpl)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to marshal config template"))
	}

	tmplBytes = bytes.ReplaceAll(tmplBytes, []byte("${{"), []byte("{{"))
	configTmpl, err := template.New("config").Parse(string(tmplBytes))
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to parse config template"))
	}

	templatedConfig := bytes.NewBuffer([]byte{})
	err = configTmpl.Option("missingkey=zero").Execute(templatedConfig, overwrites)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to template config"))
	}

	fmt.Println("templated config:", templatedConfig.String())
	err = yaml.Unmarshal(templatedConfig.Bytes(), &c.Cfg)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to unmarshal config"))
	}

	return nil
}
