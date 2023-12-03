package scaffold

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type CustomResourceDefinition struct {
	Spec struct {
		Group string `yaml:"group"`
		Names struct {
			Kind         string `yaml:"kind"`
			SingularKind string `yaml:"singular"`
		} `yaml:"names"`
		Versions []struct {
			Name   string `yaml:"name"`
			Schema struct {
				OpenAPIV3Schema struct {
					Properties map[string]Property `yaml:"properties"`
				} `yaml:"openAPIV3Schema"`
			} `yaml:"schema"`
		} `yaml:"versions"`
	} `yaml:"spec"`
}

type Property struct {
	Type       string              `yaml:"type"`
	Properties map[string]Property `yaml:"properties"`
}

func (cmd *command) generateDefaultCR() error {
	crdDir := path.Join(cmd.opts.Directory, "config", "crd", "bases")
	crds, err := os.ReadDir(crdDir)
	if err != nil {
		return fmt.Errorf("error while reading directory: %w", err)
	}

	for _, crdFile := range crds {
		if !strings.HasSuffix(crdFile.Name(), ".yaml") {
			continue
		}

		cmd.CurrentStep.LogInfof("Generating default CR for %s", crdFile)
		fullPath := filepath.Join(crdDir, crdFile.Name())

		data, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("error while opening crd file: %w", err)
		}
		var crd CustomResourceDefinition
		err = yaml.Unmarshal(data, &crd)
		if err != nil {
			return fmt.Errorf("error while unmarshalling crd file: %w", err)
		}

		group := crd.Spec.Group
		kind := crd.Spec.Names.Kind

		for _, versionSchema := range crd.Spec.Versions {
			version := versionSchema.Name
			sampleCR := generateCRFields(group+"/"+version,
				kind, versionSchema.Schema.OpenAPIV3Schema.Properties)

			crYaml, err := yaml.Marshal(sampleCR)
			if err != nil {
				return fmt.Errorf("error while marshalling default cr file: %w", err)
			}

			samplesDir, err := ensureDirExists(path.Join(cmd.opts.Directory, "config", "samples"))
			if err != nil {
				return err
			}
			filePath := path.Join(samplesDir, group+"_"+version+"_"+crd.Spec.Names.SingularKind+".yaml")
			err = os.WriteFile(filePath, crYaml, 0600)
			if err != nil {
				return fmt.Errorf("error while saving yaml: %w", err)
			}
			relativeFilePath, _ := filepath.Rel(cmd.opts.Directory, filePath)
			generatedDefaultCRFiles = append(generatedDefaultCRFiles, relativeFilePath)
		}
	}

	return nil
}

func generateCRFields(apiVersion, kind string, properties map[string]Property) interface{} {
	spec := make(map[string]interface{})
	for propName, prop := range properties {
		switch propName {
		case "apiVersion":
			spec[propName] = apiVersion
			continue
		case "kind":
			spec[propName] = kind
			continue
		case "metadata":
			spec[propName] = map[string]interface{}{
				"name": "sample-" + kind,
			}
			continue
		case "status":
			continue
		default:
		}

		if prop.Type == "object" && prop.Properties != nil {
			spec[propName] = generateCRFields("", "", prop.Properties)
		} else {
			spec[propName] = provideEmptyDataForType(prop.Type)
		}
	}
	return spec
}

func provideEmptyDataForType(t string) interface{} {
	switch t {
	case "string":
		return ""
	case "integer":
		return 0
	case "boolean":
		return false
	case "array":
		return make([]interface{}, 0)
	default:
		return nil
	}
}
