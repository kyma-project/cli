package extensions

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/extensions/parameters"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func Test_buildArgs(t *testing.T) {
	t.Run("skip nil args", func(t *testing.T) {
		require.Equal(t, args{}, buildArgs(nil, nil))
	})

	t.Run("set arg value", func(t *testing.T) {
		overwrites := fixEmptyOverwrites()
		testArgs := buildArgs(&types.Args{
			Type:     parameters.StringCustomType,
			Optional: false,
		}, overwrites)

		err := testArgs.run(&cobra.Command{}, []string{"test"})

		require.NoError(t, err)
		require.Equal(t, "test", testArgs.value.String())
		require.Equal(t, ".args.value", testArgs.value.GetPath())
		require.Equal(t, overwrites["args"].(map[string]interface{}), map[string]interface{}{
			"type":     parameters.StringCustomType,
			"optional": false,
			"value":    "",
		})
	})

	t.Run("too many given args", func(t *testing.T) {
		testArgs := buildArgs(&types.Args{
			Type:     parameters.StringCustomType,
			Optional: false,
		}, fixEmptyOverwrites())

		err := testArgs.run(&cobra.Command{}, []string{"test", "another, not expected arg"})

		require.ErrorContains(t, err, "requires exactly one argument, received 2")
	})

	t.Run("not enough args", func(t *testing.T) {
		testArgs := buildArgs(&types.Args{
			Type:     parameters.IntCustomType,
			Optional: false,
		}, fixEmptyOverwrites())

		err := testArgs.run(&cobra.Command{}, []string{})

		require.ErrorContains(t, err, "requires exactly one argument, received 0")
	})

	t.Run("wrong arg type", func(t *testing.T) {
		testArgs := buildArgs(&types.Args{
			Type:     parameters.IntCustomType,
			Optional: false,
		}, fixEmptyOverwrites())

		err := testArgs.run(&cobra.Command{}, []string{"WRONG TYPE"})

		require.ErrorContains(t, err, "parsing \"WRONG TYPE\": invalid syntax")
	})

	t.Run("optional args with no given values", func(t *testing.T) {
		testArgs := buildArgs(&types.Args{
			Type:     parameters.IntCustomType,
			Optional: true,
		}, fixEmptyOverwrites())

		err := testArgs.run(&cobra.Command{}, []string{})

		require.NoError(t, err)
		require.Empty(t, testArgs.value.String())
	})

	t.Run("optional args with one given values", func(t *testing.T) {
		testArgs := buildArgs(&types.Args{
			Type:     parameters.IntCustomType,
			Optional: true,
		}, fixEmptyOverwrites())

		err := testArgs.run(&cobra.Command{}, []string{"2"})

		require.NoError(t, err)
		require.Equal(t, "2", testArgs.value.String())
	})

	t.Run("optional args with too much args", func(t *testing.T) {
		testArgs := buildArgs(&types.Args{
			Type:     parameters.IntCustomType,
			Optional: true,
		}, fixEmptyOverwrites())

		err := testArgs.run(&cobra.Command{}, []string{"2", "3", "4", "6"})

		require.ErrorContains(t, err, "accepts at most one argument, received 4")
	})
}
