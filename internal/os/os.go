package os

import (
	"runtime"
	"strings"

	"github.com/kyma-project/cli/internal/cli"
)

func IsLinux() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	// check if the operating system is Windows Subsystem for Linux
	if kernelRelease, err := cli.RunCmd("uname", "-r"); err == nil {
		return !strings.Contains(kernelRelease, "microsoft")
	}
	return true
}
