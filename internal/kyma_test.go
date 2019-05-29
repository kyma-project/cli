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
			cmd:            []string{"mkdir", "/no-permission"},
			expectedOutput: "",
			expectedErr:    errors.New("Failed executing command 'mkdir [/no-permission]' with output 'mkdir: /no-permission: Permission denied\n' and error message 'exit status 1'"),
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
