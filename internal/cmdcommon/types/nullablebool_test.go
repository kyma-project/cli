package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNullableBool_Set(t *testing.T) {
	tests := []struct {
		name           string
		value          string
		expected       *bool
		expectedString string
		expectedError  bool
	}{
		{
			name:           "empty",
			value:          "",
			expected:       nil,
			expectedString: "",
			expectedError:  false,
		},
		{
			name:           "true",
			value:          "true",
			expected:       toptr(true),
			expectedString: "true",
			expectedError:  false,
		},
		{
			name:           "false",
			value:          "false",
			expected:       toptr(false),
			expectedString: "false",
			expectedError:  false,
		},
		{
			name:           "incorrect",
			value:          "incorrect",
			expected:       nil,
			expectedString: "",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nb := NullableBool{}
			err := nb.Set(tt.value)
			if tt.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, nb.Value)
			require.Equal(t, tt.expectedString, nb.String())
			require.Equal(t, "bool", nb.Type())
		})
	}

}
