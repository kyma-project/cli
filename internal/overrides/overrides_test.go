package overrides

import (
	"github.com/pkg/errors"
	"io/fs"
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

	override1 := map[string]interface{}{
		"chart": map[string]interface{}{
			"key4": "value4override1",
		},
	}
	err = builder.AddOverrides(override1)
	require.NoError(t, err)

	override2 := map[string]interface{}{
		"chart": map[string]interface{}{
			"key5": "value5override2",
		},
	}
	err = builder.AddOverrides(override2)
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
	err = builder.AddFile("testdata/deployment-overrides2.json")
	require.NoError(t, err)
	err = builder.AddFile("testdata/overrides.xml") // unsupported format
	require.Error(t, err)

	t.Run("detect missing file", func(t *testing.T) {
		err = builder.AddFile("testdata/nofile.yaml")
		require.Equal(t, true, errors.Is(err, fs.ErrNotExist))
		require.Error(t, err)
	})
}

func Test_AddOverrides(t *testing.T) {
	builder := Builder{}
	data := make(map[string]interface{})

	//invalid
	err := builder.AddOverrides(data)
	require.Error(t, err)

	//valid
	data["xyz"] = "abc"
	err = builder.AddOverrides(data)
	require.NoError(t, err)
}

func Test_FlattenedMap(t *testing.T) {
	testCases := []struct {
		summary        string
		givenChart     string
		givenOverrides map[string]interface{}
		expected       map[string]interface{}
	}{
		{
			summary:    "leaf key",
			givenChart: "xyz",
			givenOverrides: map[string]interface{}{
				"key": "value",
			},
			expected: map[string]interface{}{
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
			expected: map[string]interface{}{
				"xyz.key.nested": "value",
			},
		},
		{
			summary:    "multiple nested keys",
			givenChart: "xyz",
			givenOverrides: map[string]interface{}{
				"global": map[string]interface{}{
					"domainName": "local.kyma.dev",
					"installCRDs": false,
				},
				"cluster-users": map[string]interface{}{"users": map[string]interface{}{"bindStaticUsers": "false"}},
			},
			expected: map[string]interface{}{
				"xyz.global.domainName":                   "local.kyma.dev",
				"xyz.global.installCRDs":                  false,
				"xyz.cluster-users.users.bindStaticUsers": "false",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.summary, func(t *testing.T) {
			builder := Builder{}
			err := builder.AddOverrides(map[string]interface{}{
				tc.givenChart: tc.givenOverrides,
			})
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
