package component

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const defaultNamespace = "kyma-system"

//List collects component definitions
type List struct {
	DefaultNamespace string       `yaml:"defaultNamespace" json:"defaultNamespace"`
	Prerequisites    []Definition `yaml:"prerequisites" json:"prerequisites"`
	Components       []Definition `yaml:"components" json:"components"`
}

//Definition defines a component in component list
type Definition struct {
	Name      string `yaml:"name" json:"name"`
	Namespace string `yaml:"namespace" json:"namespace"`
}

func FromStrings(components []string) List {
	list := List{
		DefaultNamespace: defaultNamespace,
	}

	for _, item := range components {
		namespace := defaultNamespace

		tokens := strings.Split(item, "@")
		if len(tokens) == 2 {
			namespace = tokens[1]
		}

		definition := Definition{Name: tokens[0], Namespace: namespace}
		list.Components = append(list.Components, definition)
	}
	return list
}

// FromFile creates a new component list
func FromFile(filePath string) (List, error) {
	if filePath == "" {
		return List{}, errors.New("Path to component list file is required")
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return List{}, err
	}

	list := List{
		DefaultNamespace: defaultNamespace,
	}

	if isJSON(filePath) {
		if err := json.Unmarshal(data, &list); err != nil {
			return List{}, errors.Wrapf(err, "Failed to process component file '%s'", filePath)
		}
	} else if isYAML(filePath) {
		if err := yaml.Unmarshal(data, &list); err != nil {
			return List{}, errors.Wrapf(err, "Failed to process component file '%s'", filePath)
		}
	} else {
		return List{}, errors.New("Only JSON and YAML files are supported")
	}

	ensureNamespaces(&list)
	return list, nil
}

func isJSON(filePath string) bool {
	return filepath.Ext(filePath) == ".json"
}

func isYAML(filePath string) bool {
	ext := filepath.Ext(filePath)
	return ext == ".yaml" || ext == ".yml"
}

func ensureNamespaces(list *List) {
	for i := range list.Prerequisites {
		if list.Prerequisites[i].Namespace == "" {
			list.Prerequisites[i].Namespace = list.DefaultNamespace
		}
	}

	for i := range list.Components {
		if list.Components[i].Namespace == "" {
			list.Components[i].Namespace = list.DefaultNamespace
		}
	}
}
