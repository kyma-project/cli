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

func Test_FlattenedMap(t *testing.T) {
	testCases := []struct {
		summary        string
		givenChart     string
		givenOverrides map[string]interface{}
		expected       map[string]string
	}{
		{
			summary:    "leave key",
			givenChart: "xyz",
			givenOverrides: map[string]interface{}{
				"key": "value",
			},
			expected: map[string]string{
				"xyz.key": "value",
			},
		},
		{
			summary:    "single nested key",
			givenChart: "xyz",
			givenOverrides: map[string]interface{}{
				"key": map[string]interface{}{
					"nested": "value",
				},
			},
			expected: map[string]string{
				"xyz.key.nested": "value",
			},
		},
		{
			summary:    "multiple nested keys",
			givenChart: "xyz",
			givenOverrides: map[string]interface{}{
				"global": map[string]interface{}{
					"domainName": "local.kyma.dev",
					"ingress": map[string]interface{}{
						"domainName": "local.kyma.dev",
					},
					"installCRDs": false,
				},
				"cluster-users": map[string]interface{}{"users": map[string]interface{}{"bindStaticUsers": false}},
			},
			expected: map[string]string{
				"xyz.global.domainName":                   "local.kyma.dev",
				"xyz.global.ingress.domainName":           "local.kyma.dev",
				"xyz.global.installCRDs":                  "false",
				"xyz.cluster-users.users.bindStaticUsers": "false",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.summary, func(t *testing.T) {
			builder := Builder{}
			err := builder.AddOverrides(tc.givenChart, tc.givenOverrides)
			require.NoError(t, err)

			ovs, err := builder.Build()
			require.NoError(t, err)
			flat := ovs.FlattenedMap()

			require.Len(t, flat, len(tc.expected))
			for key, expected := range tc.expected {
				actual, exists := flat[key]
				require.Truef(t, exists, "not found %s", key)
				require.Equal(t, expected, actual)
			}
		})
	}
}
