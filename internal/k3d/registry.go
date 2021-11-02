package k3d

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/kyma-project/cli/internal/files"
	"gopkg.in/yaml.v3"
)

const (
	k3dDirectory  = "k3d"
	k3dConfigFile = "config.yaml"
)

//RegistryList containing Registry entities
type RegistryList struct {
	Registries []Registry
}

//Registry including list of nodes
type Registry struct {
	Name         string
	State        State
	PortMappings map[string]interface{}
}

//Unmarshal converts a JSON to nested structs
func (cl *RegistryList) Unmarshal(data []byte) error {
	var registries []Registry
	if err := json.Unmarshal(data, &registries); err != nil {
		return err
	}
	cl.Registries = registries
	return nil
}

type Config struct {
	MappedRegistryPort string `yaml:"mapped_registry_port"`
}

// SaveRegistryPort writes the given port to the KYMA_HOME directory
// This function is called in the `provision` step and ensures that the correct k3d registry port is used during the `deploy`step.
func SaveRegistryPort(port string) error {
	kymaHomePath, err := files.KymaHome()
	if err != nil {
		return err
	}
	p := filepath.Join(kymaHomePath, k3dDirectory)

	if _, err := os.Stat(p); os.IsNotExist(err) {
		err = os.MkdirAll(p, 0777)
		if err != nil {
			return err
		}
	}

	config := &Config{MappedRegistryPort: port}
	yamlConfig, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	configFilePath := filepath.Join(p, k3dConfigFile)

	if err := os.WriteFile(configFilePath, yamlConfig, 0666); err != nil {
		return err
	}
	return nil
}

// ReadRegistryPort reads the registry port from the KYMA_HOME directory
// This function is called in the `deploy` step and ensures that the correct k3d registry port is used.
func ReadRegistryPort() (string, error) {
	kymaHomePath, err := files.KymaHome()
	if err != nil {
		return "", err
	}
	p := filepath.Join(kymaHomePath, k3dDirectory)

	if _, err := os.Stat(p); os.IsNotExist(err) {
		err = os.MkdirAll(p, 0777)
		if err != nil {
			return "", err
		}
	}
	configFilePath := filepath.Join(p, k3dConfigFile)
	yamlConfig, err := os.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}

	var config Config
	if err := yaml.Unmarshal(yamlConfig, &config); err != nil {
		return "", err
	}

	return config.MappedRegistryPort, nil
}
