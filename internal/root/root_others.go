// +build !windows

package root

import (
	"os"
	"github.com/pkg/errors"
)

// IsWithSudo tells if a command is runnig with root privileges on
func IsWithSudo() (error) {
	if os.Getenv("SUDO_UID") != "" {
		return nil
	}
	return errors.New("Elevated permissions are required to make entries to host file. Make sure you are using sudo")
}
