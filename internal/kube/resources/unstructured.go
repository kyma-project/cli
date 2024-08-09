package resources

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ReadFromFiles reads and decodes objects from given paths
func ReadFromFiles(path ...string) ([]unstructured.Unstructured, error) {
	var result []unstructured.Unstructured
	for _, path := range path {
		file, err := os.Open(path)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open file %s", path)
		}

		decoded, err := DecodeYaml(file)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode objects from file %s", path)
		}

		result = append(result, decoded...)
	}

	return result, nil
}

// DecodeYaml decodes unstructured objects from given reader
// reader can be *os.File, *bytes.Reader or any reader
// func positions CRDs at the beginning of the list
func DecodeYaml(r io.Reader) ([]unstructured.Unstructured, error) {
	results := make([]unstructured.Unstructured, 0)
	decoder := yaml.NewDecoder(r)

	for {
		var obj map[string]interface{}
		err := decoder.Decode(&obj)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		u := unstructured.Unstructured{Object: obj}
		if u.GetObjectKind().GroupVersionKind().Kind == "CustomResourceDefinition" {
			results = append([]unstructured.Unstructured{u}, results...)
			continue
		}
		results = append(results, u)
	}

	return results, nil
}
