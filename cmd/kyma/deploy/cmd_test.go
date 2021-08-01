package deploy

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOverrides(t *testing.T) {

	t.Run("Test long override with equal separator", func(t *testing.T) {
		command := command{
			opts: &Options{
				Overrides: []string{"test.is.a.pretty.long.happy.path.test=successful"},
			},
		}
		err := assertValidOverride(t, command, `{"is":{"a":{"pretty":{"long":{"happy":{"path":{"test":"successful"}}}}}}}`)
		require.NoError(t, err)
	})

	t.Run("Test override with whitespace separator", func(t *testing.T) {
		command := command{
			opts: &Options{
				Overrides: []string{"test.happypath.test successful"},
			},
		}
		err := assertValidOverride(t, command, `{"happypath":{"test":"successful"}}`)
		require.NoError(t, err)
	})

	t.Run("Short override", func(t *testing.T) {
		command := command{
			opts: &Options{
				Overrides: []string{"test.happypath=successful"},
			},
		}
		err := assertValidOverride(t, command, `{"happypath":"successful"}`)
		require.NoError(t, err)
	})

	t.Run("Comma separated overrides", func(t *testing.T) {
		command := command{
			opts: &Options{
				Overrides: []string{"test.happypath=successful,test.secondpath=also_successful"},
			},
		}
		err := assertValidOverride(t, command, `{"happypath":"successful", "secondpath":"also_successful"}`)
		require.NoError(t, err)
	})

	t.Run("Comma separated overrides with space", func(t *testing.T) {
		command := command{
			opts: &Options{
				Overrides: []string{"test.happypath=successful, test.secondpath=also_successful"},
			},
		}
		err := assertValidOverride(t, command, `{"happypath":"successful", "secondpath":"also_successful"}`)
		require.NoError(t, err)
	})

	t.Run("No value - invalid", func(t *testing.T) {
		command := command{
			opts: &Options{
				Overrides: []string{""},
			},
		}
		_, err := command.overrides()
		assert.Error(t, err)
	})

	t.Run("One value - invalid", func(t *testing.T) {
		command := command{
			opts: &Options{
				Overrides: []string{"one"},
			},
		}
		_, err := command.overrides()
		assert.Error(t, err)
	})

	t.Run("Two values - invalid", func(t *testing.T) {
		command := command{
			opts: &Options{
				Overrides: []string{"one two"},
			},
		}
		_, err := command.overrides()
		assert.Error(t, err)
	})

	t.Run("Key without value - invalide", func(t *testing.T) {
		command := command{
			opts: &Options{
				Overrides: []string{"chart.key1="},
			},
		}
		_, err := command.overrides()
		assert.Error(t, err)
	})
}

// assert the generated override map
// use expectedJSON to define the expected result as JSON string
// (this is just for convenience as we have to compare a nested map which is painful to generate by hand)
func assertValidOverride(t *testing.T, command command, expectedJSON string) error {
	builder, err := command.overrides()
	assert.NoError(t, err)

	overrides, err := builder.Build()
	assert.NoError(t, err)

	var expected interface{}
	if err := json.Unmarshal([]byte(expectedJSON), &expected); err != nil {
		return err
	}

	v, ok := overrides.Find("test")
	assert.Equal(t, expected, v)
	assert.True(t, ok)

	return nil
}

func TestCreateCompList(t *testing.T) {
	t.Run("Create component list using --component flag", func(t *testing.T) {
		command := command{
			opts: &Options{
				Components: []string{"comp1", "comp2@test-namespace"},
			},
		}
		compList, _ := command.createCompList()

		require.Equal(t, "comp1", compList.Components[0].Name)
		// comp1 will have the default namespace which is specified in parallel-install library
		require.NotEmpty(t, compList.Components[0].Namespace)

		require.Equal(t, "comp2", compList.Components[1].Name)
		require.Equal(t, "test-namespace", compList.Components[1].Namespace)
	})
}
