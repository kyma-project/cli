package common

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"gopkg.in/yaml.v3"
)

// TemplateConfigurator is a struct that implements the Configure for the types.Action interface.
// It is used to configure an action using go templates.
type TemplateConfigurator[T any] struct {
	Cfg T
}

func (c *TemplateConfigurator[T]) Configure(cfgTmpl types.ActionConfig, overwrites types.ActionConfigOverwrites) clierror.Error {
	clierr := c.configure(cfgTmpl, overwrites)
	if clierr != nil {
		return clierror.WrapE(clierr, clierror.New("failed to configure action",
			"make sure the cli version is compatible with the extension"))
	}

	return nil
}

func (c *TemplateConfigurator[T]) configure(cfgTmpl types.ActionConfig, overwrites types.ActionConfigOverwrites) clierror.Error {
	tmplBytes, err := yaml.Marshal(cfgTmpl)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to marshal config template"))
	}

	configBytes, clierr := templateConfig(tmplBytes, overwrites)
	if clierr != nil {
		return clierror.WrapE(clierr, clierror.New("failed to template config"))
	}

	err = yaml.Unmarshal(configBytes, &c.Cfg)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to configure action"))
	}

	return nil
}
