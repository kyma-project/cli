package common

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"gopkg.in/yaml.v3"
)

type TemplateConfigurator[T any] struct {
	Cfg T
}

func (c *TemplateConfigurator[T]) Configure(in map[string]interface{}) clierror.Error {
	if in == nil {
		return clierror.New("empty config object")
	}

	configBytes, err := yaml.Marshal(in)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to marshal config"))
	}

	err = yaml.Unmarshal(configBytes, &c.Cfg)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to unmarshal config"))
	}

	return nil
}
