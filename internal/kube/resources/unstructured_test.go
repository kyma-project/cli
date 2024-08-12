package resources

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	testResourcesBytes = `
apiVersion: v1
kind: Pod
metadata:
  name: test
  namespace: testtest
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: test-crd
  namespace: testtest-crd
`

	testPodUnstructured = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "test",
				"namespace": "testtest",
			},
		},
	}

	testCRDUnstructured = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name":      "test-crd",
				"namespace": "testtest-crd",
			},
		},
	}
)

func TestReadFromFiles(t *testing.T) {
	t.Run("read objs from file", func(t *testing.T) {
		testResourcesFile, err := os.CreateTemp(t.TempDir(), "test-*.yaml")
		require.NoError(t, err)
		defer testResourcesFile.Close()

		_, err = testResourcesFile.WriteString(testResourcesBytes)
		require.NoError(t, err)

		objs, err := ReadFromFiles(testResourcesFile.Name())
		require.NoError(t, err)
		require.Len(t, objs, 2)

		require.Contains(t, objs, testPodUnstructured)
		require.Contains(t, objs, testCRDUnstructured)
	})

	t.Run("file does not exist error", func(t *testing.T) {
		objs, err := ReadFromFiles("file/does/not/exist")
		require.ErrorContains(t, err, "no such file or directory")
		require.Len(t, objs, 0)
	})

	t.Run("decode error", func(t *testing.T) {
		testResourcesFile, err := os.CreateTemp(t.TempDir(), "test-*.yaml")
		require.NoError(t, err)
		defer testResourcesFile.Close()

		_, err = testResourcesFile.WriteString("apiVersion: v1  \t\t\t  name: test")
		require.NoError(t, err)

		objs, err := ReadFromFiles(testResourcesFile.Name())
		require.ErrorContains(t, err, "yaml: mapping values are not allowed in this context")
		require.Len(t, objs, 0)
	})
}

func Test_decodeYaml(t *testing.T) {
	t.Run("Decode YAML", func(t *testing.T) {
		yaml := []byte("apiVersion: v1\nkind: Pod\nmetadata:\n  name: test\nspec:\n  containers:\n  - name: test\n    image: test")
		unstructured, err := DecodeYaml(bytes.NewReader(yaml))
		if unstructured[0].GetKind() != "Pod" {
			t.Errorf("decodeYaml() got = %v, want %v", unstructured[0].GetKind(), "Pod")
		}
		if err != nil {
			t.Errorf("decodeYaml() got = %v, want %v", err, nil)
		}
	})
}
