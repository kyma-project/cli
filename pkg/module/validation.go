package module

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kyma-project/cli/internal/kube"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	amv "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const defaultCRName = "default.yaml"

//readDefaultCR reads the default CR file's contents. The first returned value is true if the file exists. It's false if the file does not exist or an error occurred.
func readDefaultCR(modulePath string) (bool, []byte, error) {
	//TODO: Do we need to override the name or it's always "default.yaml"?
	crPath := filepath.Join(modulePath, defaultCRName)

	fileInfo, err := os.Stat(crPath)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, fmt.Errorf("Error accessing the default CR file `%q`: %w", crPath, err)
	}

	if !fileInfo.Mode().IsRegular() {
		return false, nil, fmt.Errorf("Error reading the default CR file `%q`: Not a regular file", crPath)
	}

	crData, err := os.ReadFile(crPath)
	if err != nil {
		return false, nil, fmt.Errorf("Error reading the default CR file `%q`: %w", crPath, err)
	}

	return true, crData, nil
}

//ensureDefaultNamespace parses crData and ensures that the metadata.namespace attibute exists and it's value is "default". This is because of how we use the envtest to validate the CR.
func ensureDefaultNamespace(crData []byte) ([]byte, error) {

	//Parse input data into a generic Map
	modelMap := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(crData), &modelMap)
	if err != nil {
		return nil, fmt.Errorf("Error during parsing default CR: %w", err)
	}

	//Traverse the Map to look for "metadata.namespace"
	metadataVal, ok := modelMap["metadata"]
	if !ok {
		return nil, errors.New("Error during parsing default CR: no \"metadata\" attribute")
	}

	metadataMap, ok := metadataVal.(map[string]interface{})
	if !ok {
		return nil, errors.New("Error during parsing default CR: \"metadata\" attribute is not a Map")
	}

	namespaceVal, ok := metadataMap["namespace"]
	if !ok {
		//Add the "metadata.namespace" attribute if missing
		metadataMap["namespace"] = "default"
	} else {
		//Set the "metadata.namespace" if different than "default"
		existing, ok := namespaceVal.(string)
		if !ok {
			return nil, errors.New("Error during parsing default CR: \"metadata.namespace\" attribute is not a string")
		}
		if existing != "default" {
			metadataMap["namespace"] = "default"
		}
	}

	output, err := yaml.Marshal(modelMap)
	if err != nil {
		return nil, fmt.Errorf("Error during parsing default CR: %w", err)
	}

	return output, nil
}

func ValidateDefaultCR(modulePath string, log *zap.SugaredLogger) (skipped bool, err error) {

	exists, rawCRData, err := readDefaultCR(modulePath)
	if err != nil {
		return false, err
	}
	if !exists {
		//If there's no "default.yaml" file, skip the validation.
		return true, nil
	}

	crData, err := ensureDefaultNamespace(rawCRData)
	if err != nil {
		return false, err
	}

	//TODO: How to get this directory path programatically? What's the contract?
	crdDirectoryPath := os.Getenv("CRD_DIRECTORY")

	envTest, restCfg, err := startValidationEnv(crdDirectoryPath)
	if err != nil {
		return false, err
	}

	defer func() {
		stopErr := envTest.Stop()
		if stopErr != nil {
			//TODO: This doesn't seem to print anything...
			log.Error(fmt.Errorf("Error during stopping envTest: %w", stopErr))
			//THIS does: fmt.Println(fmt.Errorf("Error during stopping envTest: %w", stopErr))
		}
	}()

	kc, err := kube.NewFromRestConfigWithTimeout(restCfg, 30*time.Second)

	err = kc.Apply(crData)
	if err != nil {
		return false, fmt.Errorf("Error during applying the default CR: %w", err)
	}
	return false, nil
}

func startValidationEnv(crdDirectoryPath string) (*envtest.Environment, *rest.Config, error) {

	envTest := &envtest.Environment{
		CRDDirectoryPaths:     []string{crdDirectoryPath},
		ErrorIfCRDPathMissing: true,
	}

	restCfg, err := envTest.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("could not start the `envtest` envionment: %w", err)
	}
	if restCfg == nil {
		return nil, nil, fmt.Errorf("could not get the RestConfig for the `envtest` envionment: %w", err)
	}

	return envTest, restCfg, nil
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
