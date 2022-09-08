package cli

import (
	"fmt"
	"os/exec"
	"strings"
)

// RunCmd executes a command with given arguments
func RunCmd(c string, args ...string) (string, error) {
	cmd := exec.Command(c, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Executing command '%s %s' failed with output '%s' and error message '%s'", c, args, out, err)
	}
	return strings.Replace(string(out), "'", "", -1), nil
}

// Pipe runs the src command and pipes its output to the dst commands' input.
// The output and stderr of dst are returned.
func Pipe(src, dst *exec.Cmd) (string, error) {
	var err error

	if dst.Stdin, err = src.StdoutPipe(); err != nil {
		return "", fmt.Errorf("could not pipe %s to %s: %w", src.Path, dst.Path, err)
	}

	if err := src.Start(); err != nil {
		return "", fmt.Errorf("error running %s: %w", src.Path, err)
	}

	out, err := dst.CombinedOutput()

	if waitErr := src.Wait(); waitErr != nil {
		return "", fmt.Errorf("%s: %w", err.Error(), waitErr)
	}
	return string(out), err
}
