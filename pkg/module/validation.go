package module

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	setup "github.com/kyma-project/cli/internal/cli/setup/envtest"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/module/kubebuilder"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	amv "k8s.io/apimachinery/pkg/util/validation"
)

var ErrEmptyCR = errors.New("provided CR is empty")

type DefaultCRValidator struct {
	crdSearchDir string
	crData       []byte
	crd          []byte
}

func NewDefaultCRValidator(cr []byte, modulePath string) *DefaultCRValidator {
	crdSearchDir := filepath.Join(modulePath, kubebuilder.OutputPath)
	return &DefaultCRValidator{
		crdSearchDir: crdSearchDir,
		crData:       cr,
	}
}

func (v *DefaultCRValidator) Run(ctx context.Context, log *zap.SugaredLogger) error {
	// skip validation if no CR detected
	if len(v.crData) == 0 {
		return ErrEmptyCR
	}

	crMap, err := parseYamlToMap(v.crData)
	if err != nil {
		return fmt.Errorf("error parsing default CR: %w", err)
	}

	group, kind, err := readGroupKind(crMap)
	if err != nil {
		return err
	}

	// find the file containing the CRD for given group and kind
	crd, crdFilePath, err := findCRDFileFor(group, kind, v.crdSearchDir)
	if err != nil {
		return fmt.Errorf("error finding CRD file in the %q directory: %w", v.crdSearchDir, err)
	}
	if crd == nil {
		return fmt.Errorf("can't find the CRD for (group: %q, kind %q)", group, kind)
	}
	v.crd = crd

	return runTestEnv(ctx, log, crdFilePath, crMap)
}

func runTestEnv(ctx context.Context, log *zap.SugaredLogger, crdFilePath string, crMap map[string]interface{}) error {
	// setup test env
	runner, err := setup.EnvTest()
	if err != nil {
		return err
	}

	err = ensureDefaultNamespace(crMap)
	if err != nil {
		return fmt.Errorf("error parsing default CR: %w", err)
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
			log.Error(fmt.Errorf("error stopping envTest: %w", err))
		}
	}()

	kc, err := kube.NewFromRestConfigWithTimeout(runner.RestClient(), 30*time.Second)
	if err != nil {
		return err
	}

	objs, err := kc.ParseManifest(properCR)
	if err != nil {
		return err
	}

	if err := kc.Apply(ctx, false, objs...); err != nil {
		return fmt.Errorf("error applying the default CR: %w", err)
	}
	return nil
}

// ensureDefaultNamespace ensures that the metadata.namespace attribute exists, and its value is "default". This is because of how we use the envtest to validate the CR.
func ensureDefaultNamespace(modelMap map[string]interface{}) error {

	//Traverse the Map to look for "metadata.namespace"
	metadataMap, err := mustReadMap(modelMap, "metadata")
	if err != nil {
		return err
	}

	namespaceVal, ok := metadataMap["namespace"]
	if !ok {
		//Add the "metadata.namespace" attribute if missing
		metadataMap["namespace"] = "default"
	} else {
		//Set the "metadata.namespace" if different than "default"
		existing, ok := namespaceVal.(string)
		if !ok {
			return errors.New("attribute \"metadata.namespace\" is not a string")
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
		retErr = fmt.Errorf("can't parse default CR data: %w", err)
		return
	}

	group = strings.Split(apiVersion, "/")[0] //e.g: apiVersion: example.org/v1
	kind, err = mustReadString(crMap, "kind")
	if err != nil {
		retErr = fmt.Errorf("can't parse default CR data: %w", err)
		return
	}

	return
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

func parseYamlToMap(crData []byte) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})
	err := yaml.Unmarshal(crData, &modelMap)
	if err != nil {
		return nil, fmt.Errorf("error parsing default CR: %w", err)
	}

	return modelMap, nil
}

func renderYamlFromMap(modelMap map[string]interface{}) ([]byte, error) {

	output, err := yaml.Marshal(modelMap)
	if err != nil {
		return nil, fmt.Errorf("error processing default CR data: %w", err)
	}

	return output, nil

}

// findCRDFileFor returns path to the file with a CRD definition for the given group and kind, if exists.
// It looks in the dirPath directory and all of its subdirectories, recursively.
func findCRDFileFor(group, kind, dirPath string) ([]byte, string, error) {

	//list all files in the dirPath and all it's subdirectories, recursively
	files, err := listFiles(dirPath)
	if err != nil {
		return nil, "", fmt.Errorf("error listing files in %q directory: %w", dirPath, err)
	}

	var found string
	var crd []byte
	for _, f := range files {
		crd, err = getCRDFileFor(group, kind, f)
		if err != nil {
			//Error is expected. Either the file is not YAML, or it's not a CRD, or it's a CRD but not the one we're looking for.
			continue
		}
		if crd != nil {
			found = f
			break
		}
	}

	if found != "" {
		return crd, found, nil
	}

	return nil, "", nil
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

		//Iteration is necessary, because the file may contain several documents and the CRD we're looking for may not be the first one.
		err = nil
		for err == nil || isNotEOF(err) {
			modelMap := make(map[string]interface{})
			err = yd.Decode(modelMap)
			if err != nil {
				//fail fast if it's not a proper YAML
				return nil, err
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
				res, err := yaml.Marshal(modelMap)
				if err != nil {
					return nil, err
				}
				return res, nil
			}
		}

		if errors.Is(err, io.EOF) {
			//Failure: No document in the file matches our search criteria.
			return nil, nil
		}

		//We should never get here, because Decode() should return EOF once there's no more data to read.
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

// ValidateName checks if the name is at least three characters long and if it conforms to the "RFC 1035 Label Names" specification (K8s compatibility requirement)
func ValidateName(name string) error {
	if len(name) < 3 {
		return errors.New("invalid module name: name must be at least three characters long")
	}

	violations := amv.IsDNS1035Label(name)
	if len(violations) == 1 {
		return fmt.Errorf("invalid module name: %s", violations[0])
	}
	if len(violations) > 1 {
		vl := "\n - " + strings.Join(violations, "\n - ")
		return fmt.Errorf("invalid module name: %s", vl)
	}

	return nil
}

type SingleManifestFileCRValidator struct {
	manifestPath string
	crData       []byte
	Crd          []byte
}

func NewSingleManifestFileCRValidator(cr []byte, manifestPath string) *SingleManifestFileCRValidator {
	return &SingleManifestFileCRValidator{
		manifestPath: manifestPath,
		crData:       cr,
	}
}

func (v *SingleManifestFileCRValidator) Run(ctx context.Context, log *zap.SugaredLogger) error {
	// skip validation if no CR detected
	if len(v.crData) == 0 {
		return ErrEmptyCR
	}

	crMap, err := parseYamlToMap(v.crData)
	if err != nil {
		return fmt.Errorf("error parsing default CR: %w", err)
	}

	group, kind, err := readGroupKind(crMap)
	if err != nil {
		return err
	}

	// extract CRD matching group and kind from the multi-document YAML manifest
	crdBytes, err := getCRDFromFile(group, kind, v.manifestPath)
	if err != nil {
		return fmt.Errorf("error finding CRD file in the %q file: %w", v.manifestPath, err)
	}
	if crdBytes == nil {
		return fmt.Errorf("can't find the CRD for (group: %q, kind %q)", group, kind)
	}
	v.Crd = crdBytes

	// store extracted CRD in a temp file
	tempDir, err := os.MkdirTemp("", "temporary-crd")
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %w", err)
	}
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			log.Warn("Error removing temporary directory", err)
		}
	}()
	tempCRDFile := filepath.Join(tempDir, "crd.yaml")
	err = os.WriteFile(tempCRDFile, crdBytes, 0600)
	if err != nil {
		return fmt.Errorf("error writing temporary CRD file %q file: %w", tempCRDFile, err)
	}

	// run testEnv using the temporary file with extracted CRD
	return runTestEnv(ctx, log, tempCRDFile, crMap)
}

func (v *SingleManifestFileCRValidator) GetCrd() []byte {
	return v.Crd
}

func (v *DefaultCRValidator) GetCrd() []byte {
	return v.crd
}
