package extensions

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/extensions/parameters"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func Test_buildFlag(t *testing.T) {
	t.Run("build string flag", func(t *testing.T) {
		overwrites := fixEmptyOverwrites()
		expectedValue := parameters.NewTyped(parameters.StringCustomType, ".flags.testname.value")
		_ = expectedValue.SetValue(toPtr("defval"))
		expectedFlag := flag{
			value:   expectedValue,
			warning: nil,
			pflag: &pflag.Flag{
				Name:      "test-name",
				Shorthand: "t",
				Usage:     "test description",
				Value:     expectedValue,
				DefValue:  expectedValue.String(),
			},
		}

		givenFlag := buildFlag(types.Flag{
			Type:         parameters.StringCustomType,
			Name:         "test-name",
			Description:  "test description",
			Shorthand:    "t",
			DefaultValue: toPtr("defval"),
			Required:     true,
		}, overwrites)
		require.Equal(t, expectedFlag, givenFlag)
		require.Equal(t, map[string]interface{}{
			"type":        parameters.StringCustomType,
			"name":        "test-name",
			"shorthand":   "t",
			"description": "test description",
			"default":     "defval",
			"value":       "defval",
		}, overwrites["flags"].(map[string]interface{})["testname"])
	})

	t.Run("build bool flag", func(t *testing.T) {
		overwrites := fixEmptyOverwrites()
		expectedValue := parameters.NewTyped(parameters.BoolCustomType, ".flags.testname.value")
		_ = expectedValue.SetValue(toPtr("false"))
		expectedFlag := flag{
			value:   expectedValue,
			warning: nil,
			pflag: &pflag.Flag{
				Name:        "test-name",
				Shorthand:   "t",
				Usage:       "test description",
				Value:       expectedValue,
				DefValue:    expectedValue.String(),
				NoOptDefVal: "true",
			},
		}

		givenFlag := buildFlag(types.Flag{
			Type:         parameters.BoolCustomType,
			Name:         "test-name",
			Description:  "test description",
			Shorthand:    "t",
			DefaultValue: toPtr("false"),
			Required:     true,
		}, overwrites)
		require.Equal(t, expectedFlag, givenFlag)
		require.Equal(t, map[string]interface{}{
			"type":        parameters.BoolCustomType,
			"name":        "test-name",
			"shorthand":   "t",
			"description": "test description",
			"default":     "false",
			"value":       false,
		}, overwrites["flags"].(map[string]interface{})["testname"])
	})

	t.Run("build flag without default", func(t *testing.T) {
		overwrites := fixEmptyOverwrites()
		expectedValue := parameters.NewTyped(parameters.StringCustomType, ".flags.testname.value")
		_ = expectedValue.SetValue(nil)
		expectedFlag := flag{
			value:   expectedValue,
			warning: nil,
			pflag: &pflag.Flag{
				Name:      "test-name",
				Shorthand: "t",
				Usage:     "test description",
				Value:     expectedValue,
				DefValue:  expectedValue.String(),
			},
		}

		givenFlag := buildFlag(types.Flag{
			Type:         parameters.StringCustomType,
			Name:         "test-name",
			Description:  "test description",
			Shorthand:    "t",
			DefaultValue: nil,
			Required:     true,
		}, overwrites)
		require.Equal(t, expectedFlag, givenFlag)
		require.Equal(t, map[string]interface{}{
			"type":        parameters.StringCustomType,
			"name":        "test-name",
			"shorthand":   "t",
			"description": "test description",
			"default":     "",
			"value":       nil,
		}, overwrites["flags"].(map[string]interface{})["testname"])
	})

	t.Run("build flag with empty default", func(t *testing.T) {
		overwrites := fixEmptyOverwrites()
		expectedValue := parameters.NewTyped(parameters.StringCustomType, ".flags.testname.value")
		_ = expectedValue.SetValue(toPtr(""))
		expectedFlag := flag{
			value:   expectedValue,
			warning: nil,
			pflag: &pflag.Flag{
				Name:      "test-name",
				Shorthand: "t",
				Usage:     "test description",
				Value:     expectedValue,
				DefValue:  expectedValue.String(),
			},
		}

		givenFlag := buildFlag(types.Flag{
			Type:         parameters.StringCustomType,
			Name:         "test-name",
			Description:  "test description",
			Shorthand:    "t",
			DefaultValue: toPtr(""),
			Required:     true,
		}, overwrites)
		require.Equal(t, expectedFlag, givenFlag)
		require.Equal(t, map[string]interface{}{
			"type":        parameters.StringCustomType,
			"name":        "test-name",
			"shorthand":   "t",
			"description": "test description",
			"default":     "",
			"value":       "",
		}, overwrites["flags"].(map[string]interface{})["testname"])
	})

	t.Run("build warning", func(t *testing.T) {
		givenFlag := buildFlag(types.Flag{
			Type:         parameters.IntCustomType,
			DefaultValue: toPtr("WRONG VALUE"),
		}, fixEmptyOverwrites())
		require.ErrorContains(t, givenFlag.warning, "parsing \"WRONG VALUE\": invalid syntax")
	})
}

func fixEmptyOverwrites() map[string]interface{} {
	return map[string]interface{}{
		"flags": map[string]interface{}{},
	}
}
