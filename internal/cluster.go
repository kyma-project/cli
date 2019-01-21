package internal

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	sleep = 5 * time.Second
)

// RunCmd executes a command with given arguments
func RunCmd(c string, args []string) (string, error) {
	cmd := exec.Command(c, args[0:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed executing command '%s %s' with output '%s' and error message '%s'", c, args, out, err)
	}
	return strings.Replace(string(out), "'", "", -1), nil
}
