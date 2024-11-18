package types_test

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/stretchr/testify/require"
)

var (
	trueConst  = true
	falseConst = false
)

func TestNullableBool_Set(t *testing.T) {
	tests := []struct {
		name          string
		value         string
		expected      *bool
		expectedError bool
	}{
		{
			name:          "empty",
			value:         "",
			expected:      nil,
			expectedError: false,
		},
		{
			name:          "true",
			value:         "true",
			expected:      &trueConst,
			expectedError: false,
		},
		{
			name:          "false",
			value:         "false",
			expected:      &falseConst,
			expectedError: false,
		},
		{
			name:          "incorrect",
			value:         "incorrect",
			expected:      nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nb := types.NullableBool{}
			err := nb.Set(tt.value)
			if tt.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, nb.Value)
		})
	}

}
