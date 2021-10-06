package values

import (
	"github.com/pkg/errors"
	"io/fs"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func Test_MergeOverrides(t *testing.T) {
	builder := builder{}
	err := builder.addValuesFile("testdata/deployment-values1.yaml")
	require.NoError(t, err)
	err = builder.addValuesFile("testdata/deployment-values2.json")
	require.NoError(t, err)

	override1 := map[string]interface{}{
		"chart": map[string]interface{}{
			"key4": "value4override1",
		},
	}
	err = builder.addValues(override1)
	require.NoError(t, err)

	override2 := map[string]interface{}{
		"chart": map[string]interface{}{
			"key5": "value5override2",
		},
	}
	err = builder.addValues(override2)
	require.NoError(t, err)

	// read expected result
	data, err := ioutil.ReadFile("testdata/deployment-values-result.yaml")
	require.NoError(t, err)
	var expected map[string]interface{}
	err = yaml.Unmarshal(data, &expected)
	require.NoError(t, err)

	// verify merge result with expected data
	result, err := builder.build()
	require.NoError(t, err)
	require.Equal(t, expected, result.toMap())
}

func Test_AddFile(t *testing.T) {
	builder := builder{}
	err := builder.addValuesFile("testdata/deployment-values1.yaml")
	require.NoError(t, err)
	err = builder.addValuesFile("testdata/deployment-values2.json")
	require.NoError(t, err)
	err = builder.addValuesFile("testdata/values.xml") // unsupported format
	require.Error(t, err)

	t.Run("detect missing file", func(t *testing.T) {
		err = builder.addValuesFile("testdata/nofile.yaml")
		require.Equal(t, true, errors.Is(err, fs.ErrNotExist))
		require.Error(t, err)
	})
}

func Test_AddOverrides(t *testing.T) {
	builder := builder{}
	data := make(map[string]interface{})

	//invalid
	err := builder.addValues(data)
	require.Error(t, err)

	//valid
	data["xyz"] = "abc"
	err = builder.addValues(data)
	require.NoError(t, err)
}
