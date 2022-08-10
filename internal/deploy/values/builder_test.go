package values

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestBuild(t *testing.T) {
	t.Run("merge values and files", func(t *testing.T) {
		override1 := map[string]interface{}{
			"chart": map[string]interface{}{
				"key4": "value4override1",
			},
		}

		override2 := map[string]interface{}{
			"chart": map[string]interface{}{
				"key5": "value5override2",
			},
		}

		builder := builder{}
		builder.
			addValuesFile("testdata/deployment-values1.yaml").
			addValuesFile("testdata/deployment-values2.json").
			addValues(override1).
			addValues(override2)

		// read expected result
		data, err := os.ReadFile("testdata/deployment-values-result.yaml")
		require.NoError(t, err)
		var expected map[string]interface{}
		err = yaml.Unmarshal(data, &expected)
		require.NoError(t, err)

		// verify merge result with expected data
		result, err := builder.build()
		require.NoError(t, err)
		require.Equal(t, expected, result)
	})

	t.Run("unsupported file format", func(t *testing.T) {
		builder := builder{}
		builder.addValuesFile("testdata/values.xml")
		_, err := builder.build()
		require.Error(t, err)
	})

	t.Run("missing file", func(t *testing.T) {
		builder := builder{}
		builder.addValuesFile("testdata/nofile.yaml")
		_, err := builder.build()
		require.Error(t, err)
	})
}
