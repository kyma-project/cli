package module

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/kyma-project/cli/pkg/module/kubebuilder"
)

func GetCrdFromModuleDef(kubebuilderProject bool, modDef *Definition) ([]byte, error) {
	if len(modDef.DefaultCR) == 0 {
		return nil, nil
	}

	crMap, err := parseYamlToMap(modDef.DefaultCR)
	if err != nil {
		return nil, fmt.Errorf("error parsing default CR: %w", err)
	}

	group, kind, err := readGroupKind(crMap)
	if err != nil {
		return nil, err
	}

	var crd []byte
	if kubebuilderProject {
		crdSearchDir := filepath.Join(modDef.Source, kubebuilder.OutputPath)
		crd, err = findCRDFileFor(group, kind, crdSearchDir)
		if err != nil {
			return nil, fmt.Errorf("error finding CRD file in the %q directory: %w", crdSearchDir, err)
		}
		if crd == nil {
			return nil, fmt.Errorf("can't find the CRD for (group: %q, kind %q)", group, kind)
		}
	} else {
		// extract CRD matching group and kind from the multi-document YAML manifest
		crd, err = getCRDFromFile(group, kind, modDef.SingleManifestPath)
		if err != nil {
			return nil, fmt.Errorf("error finding CRD file in the %q file: %w", modDef.SingleManifestPath, err)
		}
		if crd == nil {
			return nil, fmt.Errorf("can't find the CRD for (group: %q, kind %q)", group, kind)
		}
	}

	return crd, nil
}

func parseYamlToMap(crData []byte) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})
	err := yaml.Unmarshal(crData, &modelMap)
	if err != nil {
		return nil, fmt.Errorf("error parsing default CR: %w", err)
	}

	return modelMap, nil
}

func readGroupKind(crMap map[string]interface{}) (group, kind string, retErr error) {
	apiVersion, err := mustReadString(crMap, "apiVersion")
	if err != nil {
		retErr = fmt.Errorf("can't parse default CR data: %w", err)
		return
	}

	group = strings.Split(apiVersion, "/")[0] // e.g: apiVersion: example.org/v1
	kind, err = mustReadString(crMap, "kind")
	if err != nil {
		retErr = fmt.Errorf("can't parse default CR data: %w", err)
		return
	}

	return
}

// mustReadString reads a value from the given map using the given key. The value must be a string. An error is returned if the key is not in the map, or the value is not a string.
func mustReadString(input map[string]interface{}, key string) (string, error) {
	attrVal, ok := input[key]
	if !ok {
		return "", fmt.Errorf("attribute %q not found", key)
	}

	asString, ok := attrVal.(string)
	if !ok {
		return "", fmt.Errorf("attribute %q is not a string", key)
	}

	return asString, nil
}

// findCRDFileFor returns path to the file with a CRD definition for the given group and kind, if exists.
// It looks in the dirPath directory and all of its subdirectories, recursively.
func findCRDFileFor(group, kind, dirPath string) ([]byte, error) {
	// list all files in the dirPath and all it's subdirectories, recursively
	files, err := listFiles(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error listing files in %q directory: %w", dirPath, err)
	}

	var crd []byte
	for _, f := range files {
		crd, err = getCRDFileFor(group, kind, f)
		if err != nil {
			// Error is expected. Either the file is not YAML, or it's not a CRD, or it's a CRD but not the one we're looking for.
			continue
		}
		if crd != nil {
			break
		}
	}

	return crd, nil
}

// getCRDFileFor returns the crd if the given file is a CRD for given group and kind.
func getCRDFileFor(group, kind, filePath string) ([]byte, error) {
	res, err := getCRDFromFile(group, kind, filePath)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// getCRDFromFile tries to find a CRD for given group and kind in the given multi-document YAML file. Returns a generic map representation of the CRD
func getCRDFromFile(group, kind, filePath string) ([]byte, error) {
	{
		f, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("error reading \"%q\": %w", filePath, err)
		}
		defer f.Close()

		yd := yaml.NewDecoder(f)

		isNotEOF := func(err error) bool {
			return !errors.Is(err, io.EOF)
		}

		// Iteration is necessary, because the file may contain several documents and the CRD we're looking for may not be the first one.
		err = nil
		for err == nil || isNotEOF(err) {
			modelMap := make(map[string]interface{})
			err = yd.Decode(modelMap)
			if err != nil {
				// fail fast if it's not a proper YAML
				return nil, err
			}

			// Ensure it's a CRD
			rootKindVal, err := mustReadString(modelMap, "kind")
			if err != nil {
				continue
			}
			if rootKindVal != "CustomResourceDefinition" {
				continue
			}

			// Find the Group/Kind of this CRD
			specMap, err := mustReadMap(modelMap, "spec")
			if err != nil {
				continue
			}
			groupVal, err := mustReadString(specMap, "group")
			if err != nil {
				continue
			}
			namesMap, err := mustReadMap(specMap, "names")
			if err != nil {
				continue
			}
			kindVal, err := mustReadString(namesMap, "kind")
			if err != nil {
				continue
			}

			// Check if this CRD is the one we're looking for
			if groupVal == group && kindVal == kind {
				res, err := yaml.Marshal(modelMap)
				if err != nil {
					return nil, err
				}
				return res, nil
			}
		}

		if errors.Is(err, io.EOF) {
			// Failure: No document in the file matches our search criteria.
			return nil, nil
		}

		// We should never get here, because Decode() should return EOF once there's no more data to read.
		return nil, err
	}
}

// listFiles returns an unordered slice of all the files within the given directory and all it's subdirectories, recursively.
func listFiles(dirPath string) ([]string, error) {
	res := []string{}

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			res = append(res, path)
		}
		return nil
	}

	err := filepath.Walk(dirPath, walkFunc)
	if err != nil {
		return nil, fmt.Errorf("error reading directory \"%q\": %w", dirPath, err)
	}

	return res, nil
}

// mustReadMap reads a value from the given map using the given key. The value must be a Map. An error is returned if the key is not in the input map, or the value is not a Map.
func mustReadMap(input map[string]interface{}, key string) (map[string]interface{}, error) {
	attrVal, ok := input[key]
	if !ok {
		return nil, fmt.Errorf("attribute %q not found", key)
	}

	asMap, ok := attrVal.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("attribute %q is not a Map", key)
	}

	return asMap, nil
}
