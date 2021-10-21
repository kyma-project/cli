package values

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
)

type builder struct {
	files  []string
	values []map[string]interface{}
}

func (b *builder) addValuesFile(filePath string) *builder {
	b.files = append(b.files, filePath)
	return b
}

func (b *builder) addValues(values map[string]interface{}) *builder {
	b.values = append(b.values, values)
	return b
}

func (b *builder) addGlobalDomainName(domainName string) *builder {
	return b.addValues(map[string]interface{}{
		"global": map[string]interface{}{
			"domainName": domainName,
		},
	})
}

func (b *builder) addGlobalTLSCrtAndKey(tlsCrt, tlsKey string) *builder {
	return b.addValues(map[string]interface{}{
		"global": map[string]interface{}{
			"tlsCrt": tlsCrt,
			"tlsKey": tlsKey,
		},
	})
}

type serverlessRegistryConfig struct {
	enable                bool
	serverAddress         string
	internalServerAddress string
	registryAddress       string
}

func (b *builder) addServerlessRegistryConfig(config serverlessRegistryConfig) *builder {
	return b.addValues(map[string]interface{}{
		"serverless": map[string]interface{}{
			"dockerRegistry": map[string]interface{}{
				"enableInternal":        config.enable,
				"internalServerAddress": config.internalServerAddress,
				"serverAddress":         config.serverAddress,
				"registryAddress":       config.registryAddress,
			},
		},
	})
}

func (b *builder) build() (map[string]interface{}, error) {
	merged, err := b.mergeSources()
	if err != nil {
		return nil, err
	}

	return merged, nil
}

func (b *builder) mergeSources() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, file := range b.files {
		vals, err := loadValuesFile(file)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to process values defined in file '%s'", file))
		}

		if err := mergo.Map(&result, vals, mergo.WithOverride); err != nil {
			return nil, err
		}
	}

	for _, override := range b.values {
		if err := mergo.Map(&result, override, mergo.WithOverride); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func loadValuesFile(filePath string) (map[string]interface{}, error) {
	var vals map[string]interface{}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(filePath, ".json") {
		if err := json.Unmarshal(data, &vals); err != nil {
			return nil, err
		}
		return vals, nil
	} else if strings.HasSuffix(filePath, "yaml") || strings.HasSuffix(filePath, "yml") {
		if err := yaml.Unmarshal(data, &vals); err != nil {
			return nil, err
		}
		return vals, nil
	}

	return nil, errors.New("Only JSON and YAML files are supported")
}
