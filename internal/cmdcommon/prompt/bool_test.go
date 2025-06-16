package prompt

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/require"
)

func TestBoolPrompt_Table(t *testing.T) {
	tests := []struct {
		name         string
		inputReader  io.Reader
		defaultValue bool
		expectResult bool
		expectErr    string
	}{
		{
			name:         "Yes short",
			inputReader:  bytes.NewBufferString("y\n"),
			defaultValue: true,
			expectResult: true,
		},
		{
			name:         "Yes long",
			inputReader:  bytes.NewBufferString("yes\n"),
			defaultValue: false,
			expectResult: true,
		},
		{
			name:         "No short",
			inputReader:  bytes.NewBufferString("n\n"),
			defaultValue: true,
			expectResult: false,
		},
		{
			name:         "No long",
			inputReader:  bytes.NewBufferString("no\n"),
			defaultValue: true,
			expectResult: false,
		},
		{
			name:         "Default true with empty input",
			inputReader:  bytes.NewBufferString(""),
			defaultValue: true,
			expectResult: true,
		},
		{
			name:         "Default true with whitespaces",
			inputReader:  bytes.NewBufferString("  \t\n"),
			defaultValue: true,
			expectResult: false,
			expectErr:    "unexpected newline",
		},
		{
			name:         "Default false with empty input",
			inputReader:  bytes.NewBufferString(""),
			defaultValue: false,
			expectResult: false,
		},
		{
			name:         "Invalid input",
			inputReader:  bytes.NewBufferString("maybe\n"),
			defaultValue: false,
			expectResult: false,
			expectErr:    "invalid input, please enter 'y' or 'n'",
		},
		{
			name:         "Erroneous input",
			inputReader:  iotest.ErrReader(errors.New("test error")),
			defaultValue: true,
			expectResult: false,
			expectErr:    "test error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := bytes.NewBuffer([]byte{})
			b := Bool{
				reader:       tc.inputReader,
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
