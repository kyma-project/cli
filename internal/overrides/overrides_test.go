package overrides

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func Test_MergeOverrides(t *testing.T) {
	builder := Builder{}
	err := builder.AddFile("testdata/deployment-overrides1.yaml")
	require.NoError(t, err)
	err = builder.AddFile("testdata/deployment-overrides2.json")
	require.NoError(t, err)

	override1 := make(map[string]interface{})
	override1["key4"] = "value4override1"
	err = builder.AddOverrides("chart", override1)
	require.NoError(t, err)

	override2 := make(map[string]interface{})
	override2["key5"] = "value5override2"
	err = builder.AddOverrides("chart", override2)
	require.NoError(t, err)

	// read expected result
	data, err := ioutil.ReadFile("testdata/deployment-overrides-result.yaml")
	require.NoError(t, err)
	var expected map[string]interface{}
	err = yaml.Unmarshal(data, &expected)
	require.NoError(t, err)

	// verify merge result with expected data
	result, err := builder.Build()
	require.NoError(t, err)
	require.Equal(t, expected, result.Map())
}

func Test_AddFile(t *testing.T) {
	builder := Builder{}
	err := builder.AddFile("testdata/deployment-overrides1.yaml")
	require.NoError(t, err)
	err = builder.AddFile(".testdata/deployment-overrides2.json")
	require.NoError(t, err)
	err = builder.AddFile("testdata/overrides.xml") // unsupported format
	require.Error(t, err)
}

func Test_AddOverrides(t *testing.T) {
	builder := Builder{}
	data := make(map[string]interface{})

	// invalid
	err := builder.AddOverrides("", data)
	require.Error(t, err)

	//invalid
	err = builder.AddOverrides("xyz", data)
	require.Error(t, err)

	//valid
	data["test"] = "abc"
	err = builder.AddOverrides("xyz", data)
	require.NoError(t, err)
}
