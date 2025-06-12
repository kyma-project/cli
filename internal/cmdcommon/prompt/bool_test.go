package prompt

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBoolPrompt_Table(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue bool
		expectResult bool
		expectErr    string
	}{
		{
			name:         "Yes short",
			input:        "y\n",
			defaultValue: true,
			expectResult: true,
		},
		{
			name:         "Yes long",
			input:        "yes\n",
			defaultValue: false,
			expectResult: true,
		},
		{
			name:         "No short",
			input:        "n\n",
			defaultValue: true,
			expectResult: false,
		},
		{
			name:         "No long",
			input:        "no\n",
			defaultValue: true,
			expectResult: false,
		},
		{
			name:         "Default true with empty input",
			input:        "\n",
			defaultValue: true,
			expectResult: true,
		},
		{
			name:         "Default true with whitespaces",
			input:        "  \t\n",
			defaultValue: true,
			expectResult: true,
		},
		{
			name:         "Default false with empty input",
			input:        "\n",
			defaultValue: false,
			expectResult: false,
		},
		{
			name:         "Invalid input",
			input:        "maybe\n",
			defaultValue: false,
			expectResult: false,
			expectErr:    "invalid input, please enter 'y' or 'n'",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := bytes.NewBufferString(tc.input)
			output := bytes.NewBuffer([]byte{})
			b := Bool{
				reader:       input,
				writer:       output,
				message:      "Proceed?",
				defaultValue: tc.defaultValue,
			}

			result, err := b.Prompt()

			if tc.expectErr != "" {
				require.Error(t, err)
				require.Equal(t, tc.expectErr, err.Error())
				require.False(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectResult, result)
			}
		})
	}
}
