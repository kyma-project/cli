// +build windows

package root

import (
	"os"
)

// IsWithSudo tells if a command is runnig with root privileges on
func IsWithSudo() bool {
	// On windows checking admin rights by trying to open a file that is only available to admins.
	f, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false
	}
	f.Close()
	return true
}
