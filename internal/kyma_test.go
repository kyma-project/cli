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
			description:    "Checks if a command is correct.",
			cmd:            []string{"echo", "Hello!"},
			expectedOutput: "Hello!\n",
		},
		{
			name:           "Incorrect command",
			description:    "Checks if a command is correct. If not, exists and returns an error.",
			cmd:            []string{"ehco", "This is spelled incorrectly"},
			expectedOutput: "",
			expectedErr:    errors.New("Executing command 'ehco [This is spelled incorrectly]' failed with output '' and error message 'exec: \"ehco\": executable file not found in $PATH'"),
		},
		{
			name:           "Strip ' character",
			description:    "Checks if the ' character is stripped from output",
			cmd:            []string{"echo", "This is a 'single-quoted output'"},
			expectedOutput: "This is a single-quoted output\n",
		},
		{
			name:           "No args",
			description:    "Tests a command without arguments.",
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
