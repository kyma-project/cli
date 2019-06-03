package internal

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	cases := []struct {
		name           string
		description    string
		cmd            []string // cmd[0] = command, cmd[1:] = args
		expectedOutput string
		expectedErr    error
	}{
		{
			name:           "Correct command",
			description:    "Simple test checking that a correct command runs",
			cmd:            []string{"echo", "Hello!"},
			expectedOutput: "Hello!\n",
		},
		{
			name:           "Incorrect command",
			description:    "test a command that exits with error",
			cmd:            []string{"ehco", "this is wrongly spelled"},
			expectedOutput: "",
			expectedErr:    errors.New("Failed executing command 'ehco [this is wrongly spelled]' with output '' and error message 'exec: \"ehco\": executable file not found in $PATH'"),
		},
		{
			name:           "Strip ' character",
			description:    "Check that ' character is stripped from output",
			cmd:            []string{"echo", "This is a 'single-quoted output'"},
			expectedOutput: "This is a single-quoted output\n",
		},
		{
			name:           "No args",
			description:    "test a command without args",
			cmd:            []string{"echo"},
			expectedOutput: "\n",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			out, err := RunCmd(test.cmd[0], test.cmd[1:]...)
			require.Equal(t, test.expectedOutput, out, test.description)
			require.Equal(t, test.expectedErr, err)
		})
	}
}
