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

		expectedConfig := map[string]interface{}{
			"cmd1":  true,
			"cmd2":  true,
			"args1": true,
			"flag1": int64(20),
		}
		require.Equal(t, expectedConfig, actionMock.configureArg)
	})

	t.Run("action not found", func(t *testing.T) {
		cmd, err := buildCommand(fixTestExtension(), types.ActionsMap{
			// no actions defined
		})
		require.ErrorContains(t, err, "failed to build command 'cmd2': \n"+
			"  action 'action1' not found")
		require.NotNil(t, cmd)
	})

	t.Run("wrong flags and missing action", func(t *testing.T) {
		extension := fixTestExtension()
		extension.Action = "action2"
		extension.SubCommands[0].Flags = []types.Flag{
			{
				Name:         "flag1",
				Type:         parameters.BoolCustomType,
				DefaultValue: "WRONG VALUE",
			},
			{
				Name:         "flag2",
				Type:         parameters.IntCustomType,
				DefaultValue: "WRONG VALUE",
			},
		}

		cmd, err := buildCommand(extension, types.ActionsMap{
			// no actions defined
		})
		require.ErrorContains(t, err,
			"failed to build command 'cmd1': \n"+
				"  action 'action2' not found\n"+
				"failed to build command 'cmd2': \n"+
				"  flag 'flag1' error: strconv.ParseBool: parsing \"WRONG VALUE\": invalid syntax\n"+
				"  flag 'flag2' error: strconv.ParseInt: parsing \"WRONG VALUE\": invalid syntax\n"+
				"  action 'action1' not found")
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
						ConfigPath:   ".flag1",
						DefaultValue: "5",
						Required:     true,
					},
				},
				Args: &types.Args{
					Type:       parameters.BoolCustomType,
					Optional:   false,
					ConfigPath: ".args1",
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

	configureArg map[string]interface{}
	runArgs      []string
}

func (m *mockAction) Configure(arg map[string]interface{}) clierror.Error {
	m.configureArg = arg
	return m.configureError
}

func (m *mockAction) Run(_ *cobra.Command, args []string) clierror.Error {
	m.runArgs = args
	return m.runError
}
