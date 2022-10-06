package errs

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergeErrors(t *testing.T) {
	t.Parallel()

	// No errors passed -> nil
	res := MergeErrors()
	require.Nil(t, res, "MergeErrors should return nil when no parameters are received.")

	// Nil errors passed -> nil
	res = MergeErrors(nil, nil)
	require.Nil(t, res, "MergeErrors should return nil when all errors passed are nil.")

	// happy path
	res = MergeErrors(errors.New("Error 1"), errors.New("Error 2"), nil)
	require.NotNil(t, res, "MergeErrors should not return nil when at least one non-nil error is passed.")
	require.Len(t, res, 2, "MergeErrors should return a length 2 Multierror, ignoring nil errors.")
	require.Equal(t, res.Error(), "Error 1\nError 2\n", "Multierror format not as expected")
}
