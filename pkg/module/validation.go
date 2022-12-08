package module

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	setup "github.com/kyma-project/cli/internal/cli/setup/envtest"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/step"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	amv "k8s.io/apimachinery/pkg/util/validation"
)

var ErrEmptyCR = errors.New("provided CR is empty")

type DefaultCRValidator struct {
	modulePath string
	crData     []byte
}

func NewDefaultCRValidator(cr []byte, modulePath string) (*DefaultCRValidator, error) {
	return &DefaultCRValidator{
		modulePath: modulePath,
		crData:     cr,
	}, nil
}

func (v *DefaultCRValidator) Run(s step.Step, verbose bool, log *zap.SugaredLogger) error {
	// skip validation if no CR detected
	if len(v.crData) == 0 {
		return nil
	}

	// setup test env
	runner, err := setup.EnvTest(s, verbose)
	if err != nil {
		return err
	}

	crMap, err := parseYamlToMap(v.crData)
	if err != nil {
		return fmt.Errorf("Error parsing default CR: %w", err)
	}

	group, kind, err := readGroupKind(crMap)
	if err != nil {
		return err
	}

	searchDirPath := filepath.Join(v.modulePath, "charts")
	crdFound, crdFilePath, err := findCRDFileFor(group, kind, searchDirPath)
	if err != nil {
		return fmt.Errorf("Error finding CRD file in the %q directory: %w", searchDirPath, err)
	}
	if !crdFound {
		return fmt.Errorf("Can't find the CRD for (group: %q, kind %q)", group, kind)
	}

	err = ensureDefaultNamespace(crMap)
	if err != nil {
		return err
	}

	properCR, err := renderYamlFromMap(crMap)
	if err != nil {
		return err
	}

	if err := runner.Start(crdFilePath, log); err != nil {
		return err
	}
	defer func() {
		if err := runner.Stop(); err != nil {
			//TODO: This doesn't seem to print anything...
			log.Error(fmt.Errorf("Error stopping envTest: %w", err))
			//THIS does: fmt.Println(fmt.Errorf("Error stopping envTest: %w", stopErr))
		}
	}()

	kc, err := kube.NewFromRestConfigWithTimeout(runner.RestClient(), 30*time.Second)
	if err != nil {
		return err
	}

	if err := kc.Apply(properCR); err != nil {
		return fmt.Errorf("Error applying the default CR: %w", err)
	}
	return nil
}

// ensureDefaultNamespace ensures that the metadata.namespace attribute exists and it's value is "default". This is because of how we use the envtest to validate the CR.
func ensureDefaultNamespace(modelMap map[string]interface{}) error {

	//Traverse the Map to look for "metadata.namespace"
	metadataMap, err := mustReadMap(modelMap, "metadata")
	if err != nil {
		return fmt.Errorf("Error parsing default CR: %w", err)
	}

	namespaceVal, ok := metadataMap["namespace"]
	if !ok {
		//Add the "metadata.namespace" attribute if missing
		metadataMap["namespace"] = "default"
	} else {
		//Set the "metadata.namespace" if different than "default"
		existing, ok := namespaceVal.(string)
		if !ok {
			return errors.New("Error parsing default CR: Attribute \"metadata.namespace\" is not a string")
		}
		if existing != "default" {
			metadataMap["namespace"] = "default"
		}
	}

	return nil
}

func readGroupKind(crMap map[string]interface{}) (group, kind string, retErr error) {
	apiVersion, err := mustReadString(crMap, "apiVersion")
	if err != nil {
		retErr = fmt.Errorf("Can't parse default CR data: %w", err)
		return
	}

	group = strings.Split(apiVersion, "/")[0] //e.g: apiVersion: example.org/v1
	kind, err = mustReadString(crMap, "kind")
	if err != nil {
		retErr = fmt.Errorf("Can't parse default CR data: %w", err)
		return
	}

	return
}

// mustReadMap reads a value from the given map using the given key. The value must be a Map. An error is returned if the key is not in the input map, or the value is not a Map.
func mustReadMap(input map[string]interface{}, key string) (map[string]interface{}, error) {
	attrVal, ok := input[key]
	if !ok {
		return nil, fmt.Errorf("Attribute %q not found", key)
	}

	asMap, ok := attrVal.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Attribute %q is not a Map", key)
	}

	return asMap, nil
}

// mustReadString reads a value from the given map using the given key. The value must be a string. An error is returned if the key is not in the map, or the value is not a string.
func mustReadString(input map[string]interface{}, key string) (string, error) {
	attrVal, ok := input[key]
	if !ok {
		return "", fmt.Errorf("Attribute %q not found", key)
	}

	asString, ok := attrVal.(string)
	if !ok {
		return "", fmt.Errorf("Attribute %q is not a string", key)
	}

	return asString, nil
}

func parseYamlToMap(crData []byte) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(crData), &modelMap)
	if err != nil {
		return nil, fmt.Errorf("Error parsing default CR: %w", err)
	}

	return modelMap, nil
}

func renderYamlFromMap(modelMap map[string]interface{}) ([]byte, error) {

	output, err := yaml.Marshal(modelMap)
	if err != nil {
		return nil, fmt.Errorf("Error processing default CR data: %w", err)
	}

	return output, nil

}

// findCRDFileFor returns path to the file with a CRD definition for the given group and kind, if exists. It looks in the dirPath directory and all of it's subdirectories, recursively. The first parameter is true if the file is found, it's false otherwise.
func findCRDFileFor(group, kind, dirPath string) (bool, string, error) {

	//list all files in the dirPath and all it's subdirectories, recursively
	files, err := listFiles(dirPath)
	if err != nil {
		return false, "", fmt.Errorf("Error listing files in %q directory: %w", dirPath, err)
	}

	var found string
	for _, f := range files {
		//fmt.Printf("- Checking file: %q\n", f)
		ok, err := isCRDFileFor(group, kind, f)
		if err != nil {
			//Error is expected. Either the file is not YAML, or it's not a CRD, or it's a CRD but not the one we're looking for.
			continue
		}
		if ok {
			found = f
			break
		}
	}

	if found != "" {
		return true, found, nil
	}

	return false, "", nil
}

// isCRDFileFor checks if the given file is a CRD for given group and kind.
func isCRDFileFor(group, kind, filePath string) (bool, error) {
	{

		f, err := os.Open(filePath)
		if err != nil {
			return false, fmt.Errorf("Error reading \"%q\": %w", filePath, err)
		}
		defer f.Close()

		yd := yaml.NewDecoder(f)

		isNotEOF := func(err error) bool {
			return !errors.Is(err, io.EOF)
		}

		//Iteration is necessary, because the file may contain several documents and the CRD we're looking for may not be the first one.
		err = nil
		for err == nil || isNotEOF(err) {
			modelMap := make(map[string]interface{})
			err = yd.Decode(modelMap)
			if err != nil {
				//fail fast if it's not a proper YAML
				return false, err
			}

			//Ensure it's a CRD
			rootKindVal, err := mustReadString(modelMap, "kind")
			if err != nil {
				continue
			}
			if rootKindVal != "CustomResourceDefinition" {
				continue
			}

			//Find the Group/Kind of this CRD
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

			//Check if this CRD is the one we're looking for
			if groupVal == group && kindVal == kind {
				return true, nil
			}
		}

		if errors.Is(err, io.EOF) {
			//Failure: No document in the file matches our search criteria.
			return false, nil
		}

		//We should never get here, because Decode() should return EOF once there's no more data to read.
		return false, err
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
		return nil, fmt.Errorf("Error reading directory \"%q\": %w", dirPath, err)
	}

	return res, nil
}

// ValidateName checks if the name is at least three characters long and if it conforms to the "RFC 1035 Label Names" specification (K8s compatibility requirement)
func ValidateName(name string) error {
	if len(name) < 3 {
		return errors.New("Invalid module name: name must be at least three characters long")
	}

	violations := amv.IsDNS1035Label(name)
	if len(violations) == 1 {
		return fmt.Errorf("Invalid module name: %s", violations[0])
	}
	if len(violations) > 1 {
		vl := "\n - " + strings.Join(violations, "\n - ")
		return fmt.Errorf("Invalid module name: %s", vl)
	}

	return nil
}
