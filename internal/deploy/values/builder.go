package values

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
)

var (
	supportedFileExt = []string{"yaml", "yml", "json"}
)

type builder struct {
	files  []string
	values []map[string]interface{}
}

func (b *builder) addValuesFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.Wrap(err, "invalid override file: not exists")
	}

	for _, ext := range supportedFileExt {
		if strings.HasSuffix(filePath, fmt.Sprintf(".%s", ext)) {
			b.files = append(b.files, filePath)
			return nil
		}
	}
	return fmt.Errorf("Unsupported override file extension. Supported extensions are: %s", strings.Join(supportedFileExt, ", "))
}

func (b *builder) addValues(values map[string]interface{}) error {
	if len(values) < 1 {
		return fmt.Errorf("invalid values: empty")
	}

	b.values = append(b.values, values)
	return nil
}

func (b *builder) addGlobalDomainName(domainName string) error {
	return b.addValues(map[string]interface{}{
		"global": map[string]interface{}{
			"domainName": domainName,
		},
	})
}

func (b *builder) addGlobalTLSCrtAndKey(tlsCrt, tlsKey string) error {
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

func (b *builder) addServerlessRegistryConfig(config serverlessRegistryConfig) error {
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

	var fileOverrides map[string]interface{}
	for _, file := range b.files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(file, ".json") {
			err = json.Unmarshal(data, &fileOverrides)
		} else {
			err = yaml.Unmarshal(data, &fileOverrides)
		}
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Failed to process configuration values defined in file '%s'", file))
		}
		if err := mergo.Map(&result, fileOverrides, mergo.WithOverride); err != nil {
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
