// +build !windows

package root

import (
	"os"
)

// IsWithSudo tells if a command is runnig with root privileges on
func IsWithSudo() bool {
	return os.Getenv("SUDO_UID") != ""
}
