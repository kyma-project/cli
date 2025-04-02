package flags

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	t.Run("validate", func(t *testing.T) {
		t.Run("zero", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var1", "value"))
			require.NoError(t, flagSet.Set("var2", "value"))

			clierr := Validate(flagSet)
			require.Nil(t, clierr)
		})

		t.Run("many", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var1", "value"))
			require.NoError(t, flagSet.Set("var2", "value"))

			clierr := Validate(flagSet,
				MarkRequired("var1", "var2"),
				MarkRequiredTogether("var1", "var2"),
				MarkMutuallyExclusive("var2", "var3"),
				MarkPrerequisites("var1", "var2"),
				MarkExclusive("var2", "var3"),
			)
			require.Nil(t, clierr)
		})

		t.Run("error on many rules", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var2", "value"))
			require.NoError(t, flagSet.Set("var3", "value"))

			expectedCliErr := fixValidationErr(
				"all flags in group [var1 var2] must be set, missing [var1]",
				"all flags in group [var1 var2] must be set if any is used, missing [var1]",
				"only one flag from groud [var2 var3] can be used at the same time, used [var2 var3]",
				"all flags in group [var1] must be set when [var2] flag is used, missing [var1]",
				"flags in group [var3] can't be used together with [var2], used [var3]",
			)

			clierr := Validate(flagSet,
				MarkRequired("var1", "var2"),
				MarkRequiredTogether("var1", "var2"),
				MarkMutuallyExclusive("var2", "var3"),
				MarkPrerequisites("var2", "var1"),
				MarkExclusive("var2", "var3"),
			)
			require.Equal(t, expectedCliErr, clierr)
		})
	})

	t.Run("validate required", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var1", "value"))
			require.NoError(t, flagSet.Set("var2", "value"))

			clierr := Validate(flagSet, MarkRequired("var1", "var2"))
			require.Nil(t, clierr)
		})

		t.Run("one missing", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var1", "value"))

			expectedCliErr := fixValidationErr("all flags in group [var1 var2] must be set, missing [var2]")

			clierr := Validate(flagSet, MarkRequired("var1", "var2"))
			require.Equal(t, expectedCliErr, clierr)
		})

		t.Run("two missing", func(t *testing.T) {
			flagSet := fixTestFlagSet()

			expectedCliErr := fixValidationErr("all flags in group [var1 var2] must be set, missing [var1 var2]")

			clierr := Validate(flagSet, MarkRequired("var1", "var2"))
			require.Equal(t, expectedCliErr, clierr)
		})
	})

	t.Run("validate together required", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var1", "value"))
			require.NoError(t, flagSet.Set("var2", "value"))

			clierr := Validate(flagSet, MarkRequiredTogether("var1", "var2"))
			require.Nil(t, clierr)
		})

		t.Run("ok - both missing", func(t *testing.T) {
			flagSet := fixTestFlagSet()

			clierr := Validate(flagSet, MarkRequiredTogether("var1", "var2"))
			require.Nil(t, clierr)
		})

		t.Run("one missing", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var2", "value"))

			expectedCliErr := fixValidationErr("all flags in group [var1 var2] must be set if any is used, missing [var1]")

			clierr := Validate(flagSet, MarkRequiredTogether("var1", "var2"))
			require.Equal(t, expectedCliErr, clierr)
		})
	})

	t.Run("validate one required", func(t *testing.T) {
		t.Run("ok - first set", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var1", "value"))

			clierr := Validate(flagSet, MarkOneRequired("var1", "var2"))
			require.Nil(t, clierr)
		})

		t.Run("ok - second set", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var2", "value"))

			clierr := Validate(flagSet, MarkOneRequired("var1", "var2"))
			require.Nil(t, clierr)
		})

		t.Run("missing flags", func(t *testing.T) {
			flagSet := fixTestFlagSet()

			expectedCliErr := fixValidationErr("at least one of the flags from the group [var1 var2] must be used")

			clierr := Validate(flagSet, MarkOneRequired("var1", "var2"))
			require.Equal(t, expectedCliErr, clierr)
		})
	})

	t.Run("validate exactly one required", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var3", "value"))

			clierr := Validate(flagSet, MarkExactlyOneRequired("var1", "var2", "var3"))
			require.Nil(t, clierr)
		})

		t.Run("missing flags", func(t *testing.T) {
			flagSet := fixTestFlagSet()

			expectedCliErr := fixValidationErr("exactly one from group [var1 var2 var3] must be set")

			clierr := Validate(flagSet, MarkExactlyOneRequired("var1", "var2", "var3"))
			require.Equal(t, expectedCliErr, clierr)
		})

		t.Run("too many used", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var1", "value"))
			require.NoError(t, flagSet.Set("var2", "value"))
			require.NoError(t, flagSet.Set("var3", "value"))

			expectedCliErr := fixValidationErr("exactly one from group [var1 var2 var3] must be set, used [var1 var2 var3]")

			clierr := Validate(flagSet, MarkExactlyOneRequired("var1", "var2", "var3"))
			require.Equal(t, expectedCliErr, clierr)
		})
	})

	t.Run("validate prerequisites", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var1", "value"))
			require.NoError(t, flagSet.Set("var2", "value"))
			require.NoError(t, flagSet.Set("var3", "value"))

			clierr := Validate(flagSet, MarkPrerequisites("var1", "var2", "var3"))
			require.Nil(t, clierr)
		})

		t.Run("missing one", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var1", "value"))
			require.NoError(t, flagSet.Set("var3", "value"))

			expectedCliErr := fixValidationErr("all flags in group [var2 var3] must be set when [var1] flag is used, missing [var2]")

			clierr := Validate(flagSet, MarkPrerequisites("var1", "var2", "var3"))
			require.Equal(t, expectedCliErr, clierr)
		})

		t.Run("missing all", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var1", "value"))

			expectedCliErr := fixValidationErr("all flags in group [var2 var3] must be set when [var1] flag is used, missing [var2 var3]")

			clierr := Validate(flagSet, MarkPrerequisites("var1", "var2", "var3"))
			require.Equal(t, expectedCliErr, clierr)
		})

		t.Run("skip validation when flag is not set", func(t *testing.T) {
			flagSet := fixTestFlagSet()

			clierr := Validate(flagSet, MarkPrerequisites("var1", "var2", "var3"))
			require.Nil(t, clierr)
		})
	})

	t.Run("validate exclusive", func(t *testing.T) {
		t.Run("ok - exclusive flags missing", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var1", "value"))
			require.NoError(t, flagSet.Set("var4", "value"))
			require.NoError(t, flagSet.Set("var5", "value"))

			clierr := Validate(flagSet, MarkExclusive("var1", "var2", "var3"))
			require.Nil(t, clierr)
		})

		t.Run("ok - main flag missing", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var2", "value"))
			require.NoError(t, flagSet.Set("var3", "value"))
			require.NoError(t, flagSet.Set("var4", "value"))
			require.NoError(t, flagSet.Set("var5", "value"))

			clierr := Validate(flagSet, MarkExclusive("var1", "var2", "var3"))
			require.Nil(t, clierr)
		})

		t.Run("used flag with exclusive flags", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var1", "value"))
			require.NoError(t, flagSet.Set("var2", "value"))
			require.NoError(t, flagSet.Set("var3", "value"))

			expectedCliErr := fixValidationErr("flags in group [var2 var3] can't be used together with [var1], used [var2 var3]")

			clierr := Validate(flagSet, MarkExclusive("var1", "var2", "var3"))
			require.Equal(t, expectedCliErr, clierr)
		})
	})

	t.Run("validate mutually exclusive", func(t *testing.T) {
		t.Run("ok - none from mutually exclusive used", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var4", "value"))
			require.NoError(t, flagSet.Set("var5", "value"))

			clierr := Validate(flagSet, MarkMutuallyExclusive("var1", "var2", "var3"))
			require.Nil(t, clierr)
		})

		t.Run("ok - one from mutually exclusive used", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var1", "value"))
			require.NoError(t, flagSet.Set("var4", "value"))
			require.NoError(t, flagSet.Set("var5", "value"))

			clierr := Validate(flagSet, MarkMutuallyExclusive("var1", "var2", "var3"))
			require.Nil(t, clierr)
		})

		t.Run("two mutually exclusive flags used", func(t *testing.T) {
			flagSet := fixTestFlagSet()
			require.NoError(t, flagSet.Set("var2", "value"))
			require.NoError(t, flagSet.Set("var3", "value"))
			require.NoError(t, flagSet.Set("var4", "value"))
			require.NoError(t, flagSet.Set("var5", "value"))

			expectedCliErr := fixValidationErr("only one flag from groud [var1 var2 var3] can be used at the same time, used [var2 var3]")

			clierr := Validate(flagSet, MarkMutuallyExclusive("var1", "var2", "var3"))
			require.Equal(t, expectedCliErr, clierr)
		})
	})
}

func fixTestFlagSet() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("test", pflag.ExitOnError)
	flagSet.String("var1", "", "")
	flagSet.String("var2", "", "")
	flagSet.String("var3", "", "")
	flagSet.String("var4", "", "")
	flagSet.String("var5", "", "")
	return flagSet
}

func fixValidationErr(hints ...string) clierror.Error {
	return clierror.New("failed to validate given flags", hints...)
}
