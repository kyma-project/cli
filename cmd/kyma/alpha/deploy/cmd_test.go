package deploy

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOverrides(t *testing.T) {

	t.Run("Test long override with equal separator", func(t *testing.T) {
		command := command{
			opts: &Options{
				Overrides: []string{"this.is.a.pretty.long.happy.path.test=successful"},
			},
		}
		assertValidOverride(t, command, `{"this":{"is":{"a":{"pretty":{"long":{"happy":{"path":{"test":"successful"}}}}}}}}`)
	})

	t.Run("Test override with whitespace separator", func(t *testing.T) {
		command := command{
			opts: &Options{
				Overrides: []string{"happy.path.test successful"},
			},
		}
		assertValidOverride(t, command, `{"happy":{"path":{"test":"successful"}}}`)
	})

	t.Run("Short override", func(t *testing.T) {
		command := command{
			opts: &Options{
				Overrides: []string{"happy.path=successful"},
			},
		}
		assertValidOverride(t, command, `{"happy":{"path":"successful"}}`)
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
func assertValidOverride(t *testing.T, command command, expectedJSON string) {
	overrides, err := command.overrides()
	assert.NoError(t, err)

	mergedOverrides, err := overrides.Merge()
	assert.NoError(t, err)

	var expected interface{}
	json.Unmarshal([]byte(expectedJSON), &expected)

	assert.Equal(t, expected, mergedOverrides)
}
