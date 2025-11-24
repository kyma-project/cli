package os

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func IsLinux() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	// check if the operating system is Windows Subsystem for Linux
	if kernelRelease, err := RunCmd("uname", "-r"); err == nil {
		return !strings.Contains(kernelRelease, "microsoft")
	}
	return true
}

func RunCmd(c string, args ...string) (string, error) {
	cmd := exec.Command(c, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Executing command '%s %s' failed with output '%s' and error message '%s'", c, args, out, err)
	}
	return strings.Replace(string(out), "'", "", -1), nil
}
