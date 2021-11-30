// +build !windows

package root

import (
	"os"
	"github.com/pkg/errors"
)

// IsWithSudo tells if a command is runnig with root privileges on
func IsWithSudo() (bool,error) {
	if os.Getenv("SUDO_UID") != "" {
		return true, nil
	}
	return false, errors.New("Elevated permissions are required to make entries to host file. Make sure you are using sudo.")
}
