package extensions

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/extensions/parameters"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func Test_buildCommand(t *testing.T) {
	t.Run("build command", func(t *testing.T) {
		actionMock := &mockAction{}

		cmd, err := buildCommand(fixTestExtension(), types.ActionsMap{
			"action1": actionMock,
		})
		require.NoError(t, err)

		cmd.SetArgs([]string{"cmd2", "true", "--flag1", "20"})
		err = cmd.Execute()
		require.NoError(t, err)

		require.Equal(t, map[string]interface{}{
			"cmd2": true,
		}, actionMock.configureArg)
		require.Equal(t, types.ActionConfigOverwrites{
			"args": map[string]interface{}{
				"optional": false,
				"type":     parameters.BoolCustomType,
				"value":    true,
			},
			"flags": map[string]interface{}{
				"flag1": map[string]interface{}{
					"default":     "5",
					"description": "flag-desc1",
					"name":        "flag1",
					"shorthand":   "f",
					"type":        parameters.IntCustomType,
					"value":       int64(20),
				},
			},
		}, actionMock.configureOverwritesArg)
	})

	t.Run("action not found", func(t *testing.T) {
		cmd, err := buildCommand(fixTestExtension(), types.ActionsMap{
			// no actions defined
		})
		require.NoError(t, err)
		require.NotNil(t, cmd)
	})

	t.Run("error with wrong flags when action not defined", func(t *testing.T) {
		extension := fixTestExtension()
		extension.Action = "action2"
		extension.SubCommands[0].Flags = []types.Flag{
			{
				Name:         "flag1",
				Type:         parameters.BoolCustomType,
				DefaultValue: toPtr("WRONG VALUE"),
			},
			{
				Name:         "flag2",
				Type:         parameters.IntCustomType,
				DefaultValue: toPtr("WRONG VALUE"),
			},
		}

		cmd, err := buildCommand(extension, types.ActionsMap{
			// no actions defined
		})
		require.EqualError(t, err, "failed to build command 'cmd2':\n"+
			"  flag 'flag1' error: strconv.ParseBool: parsing \"WRONG VALUE\": invalid syntax\n"+
			"  flag 'flag2' error: strconv.ParseInt: parsing \"WRONG VALUE\": invalid syntax")
		require.NotNil(t, cmd)
	})
}

func fixTestExtension() types.Extension {
	return types.Extension{
		Metadata: types.Metadata{
			Name:            "cmd1",
			Description:     "desc1",
			DescriptionLong: "desc-long1",
		},
		Config: map[string]interface{}{
			"cmd1": true,
		},
		SubCommands: []types.Extension{
			{
				Metadata: types.Metadata{
					Name:            "cmd2",
					Description:     "desc2",
					DescriptionLong: "desc-long2",
				},
				Action: "action1",
				Flags: []types.Flag{
					{
						Type:         parameters.IntCustomType,
						Name:         "flag1",
						Description:  "flag-desc1",
						Shorthand:    "f",
						DefaultValue: toPtr("5"),
						Required:     true,
					},
				},
				Args: &types.Args{
					Type:     parameters.BoolCustomType,
					Optional: false,
				},
				Config: map[string]interface{}{
					"cmd2": true,
				},
			},
		},
	}
}

type mockAction struct {
	configureError clierror.Error
	runError       clierror.Error

	configureArg           types.ActionConfig
	configureOverwritesArg types.ActionConfigOverwrites
	runArgs                []string
}

func (m *mockAction) Configure(arg types.ActionConfig, overwritesArg types.ActionConfigOverwrites) clierror.Error {
	m.configureArg = arg
	m.configureOverwritesArg = overwritesArg
	return m.configureError
}

func (m *mockAction) Run(_ *cobra.Command, args []string) clierror.Error {
	m.runArgs = args
	return m.runError
}
