package common

import (
	"bytes"
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

func (c *TemplateConfigurator[T]) Configure(cfgTmpl types.ActionConfigTmpl, overwrites types.ActionConfigOverwrites) clierror.Error {
	configTmpl, err := template.New("config").Parse(cfgTmpl)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to parse config template"))
	}

	templatedConfig := bytes.NewBuffer([]byte{})
	err = configTmpl.Option("missingkey=zero").Execute(templatedConfig, overwrites)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to template config"))
	}

	err = yaml.Unmarshal(templatedConfig.Bytes(), &c.Cfg)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to unmarshal config"))
	}

	return nil
}
