// +build windows

package root

import (
	"github.com/pkg/errors"
	"os"
)

// IsWithSudo tells if a command is running with root privileges on
func IsWithSudo() error {
	// On Windows checking admin rights by trying to open a file that is only available to admins.
	f, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return errors.New("Elevated permissions are required to make entries to host file. Make sure you are running the command as an administrator")
	}
	f.Close()
	return nil
}
