package values

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
)

type builder struct {
	files         []string
	values        []map[string]interface{}
	defaultValues []map[string]interface{}
}

func (b *builder) addValuesFile(filePath string) *builder {
	b.files = append(b.files, filePath)
	return b
}

func (b *builder) addValues(values map[string]interface{}) *builder {
	b.values = append(b.values, values)
	return b
}

func (b *builder) addDefaultValues(values map[string]interface{}) *builder {
	b.defaultValues = append(b.defaultValues, values)
	return b
}

func (b *builder) addDefaultGlobalDomainName(domainName string) *builder {
	return b.addDefaultValues(map[string]interface{}{
		"global": map[string]interface{}{
			"domainName": domainName,
		},
	})
}

func (b *builder) addDefaultGlobalTLSCrtAndKey(tlsCrt, tlsKey string) *builder {
	return b.addDefaultValues(map[string]interface{}{
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

func (b *builder) addDefaultServerlessRegistryConfig(config serverlessRegistryConfig) *builder {
	return b.addDefaultValues(map[string]interface{}{
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

// https://github.com/GoogleContainerTools/kaniko/issues/1592
// https://github.com/kyma-project/kyma/issues/13051
func (b *builder) addDefaultServerlessKanikoForce() *builder {
	return b.addDefaultValues(map[string]interface{}{
		"serverless": map[string]interface{}{
			"containers": map[string]interface{}{
				"manager": map[string]interface{}{
					"envs": map[string]interface{}{
						"functionBuildExecutorArgs": map[string]interface{}{
							"value": "--insecure,--skip-tls-verify,--skip-unused-stages,--log-format=text,--cache=true,--force",
						},
					},
				},
			},
		},
	})
}

func (b *builder) addDefaultk3dValuesForIstio() *builder {
	return b.addDefaultValues(map[string]interface{}{
		"istio": map[string]interface{}{
			"helmValues": map[string]interface{}{
				"cni": map[string]string{
					"cniConfDir": "/var/lib/rancher/k3s/agent/etc/cni/net.d",
					"cniBinDir":  "/bin",
				},
			},
		},
	})
}

func (b *builder) addDefaultGkeValuesForIstio() *builder {
	return b.addDefaultValues(map[string]interface{}{
		"istio": map[string]interface{}{
			"helmValues": map[string]interface{}{
				"cni": map[string]interface{}{
					"cniBinDir": "/home/kubernetes/bin",
					"resourceQuotas": map[string]bool{
						"enabled": true,
					},
				},
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

	for _, defaults := range b.defaultValues {
		if err := mergo.Map(&result, defaults, mergo.WithOverride); err != nil {
			return nil, errors.Wrap(err, "while merging values from files")
		}
	}

	for _, file := range b.files {
		vals, err := loadValuesFile(file)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to process values defined in file '%s'", file))
		}

		if err := mergo.Map(&result, vals, mergo.WithOverride); err != nil {
			return nil, errors.Wrap(err, "while merging values from files")
		}
	}

	for _, override := range b.values {
		if err := mergo.Map(&result, override, mergo.WithOverride); err != nil {
			return nil, errors.Wrap(err, "while merging overrides values")
		}
	}

	return result, nil
}

func loadValuesFile(filePath string) (map[string]interface{}, error) {
	var vals map[string]interface{}

	data, err := os.ReadFile(filePath)
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
