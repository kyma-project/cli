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
	overrides, err := command.overrides()
	assert.NoError(t, err)

	mergedOverrides, err := overrides.Merge()
	assert.NoError(t, err)

	var expected interface{}
	if err := json.Unmarshal([]byte(expectedJSON), &expected); err != nil {
		return err
	}

	assert.Equal(t, expected, mergedOverrides["test"])

	return nil
}
