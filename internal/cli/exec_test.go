package cli

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	t.Parallel()
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

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := RunCmd(tc.cmd[0], tc.cmd[1:]...)
			require.Equal(t, tc.expectedOutput, out, tc.description)
			require.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestPipe(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name           string
		description    string
		src            *exec.Cmd
		dst            *exec.Cmd
		expectedOutput string
		expectedErr    error
	}{
		{
			name:           "Correct pipe",
			description:    "Checks if a pipe is correct.",
			src:            exec.Command("echo", "Hello! I am gonna count the words"),
			dst:            exec.Command("wc", "-w"),
			expectedOutput: "7",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := Pipe(tc.src, tc.dst)
			require.Equal(t, tc.expectedOutput, strings.TrimSpace(out), tc.description) // trim spaces to avoid different results depending on OS
			require.Equal(t, tc.expectedErr, err)
		})
	}
}
