package actions

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"gopkg.in/yaml.v3"
)

type configurator[T any] struct {
	cfg T
}

func (c *configurator[T]) Configure(in map[string]interface{}) clierror.Error {
	if in == nil {
		return clierror.New("empty config object")
	}

	configBytes, err := yaml.Marshal(in)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to marshal config"))
	}

	err = yaml.Unmarshal(configBytes, &c.cfg)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to unmarshal config"))
	}

	return nil
}
