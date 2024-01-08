package scaffold

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
)

type Generator struct {
	ModuleName         string
	ModuleVersion      string
	ModuleChannel      string
	ModuleConfigFile   string
	ManifestFile       string
	SecurityConfigFile string
	DefaultCRFile      string
}

func (g *Generator) fileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil

	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil

	} else {
		return false, err
	}
}

func (g *Generator) generateYamlFileFromObject(obj interface{}, filePath string) error {
	reflectValue := reflect.ValueOf(obj)
	var yamlBuilder strings.Builder
	generateYaml(&yamlBuilder, reflectValue, 0, "")
	yamlString := yamlBuilder.String()

	err := os.WriteFile(filePath, []byte(yamlString), 0600)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}
