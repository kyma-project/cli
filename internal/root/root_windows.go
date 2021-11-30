// +build windows

package root

import (
	"os"
	"github.com/pkg/errors"
)

// IsWithSudo tells if a command is runnig with root privileges on
func IsWithSudo() (bool,error){
	// On windows checking admin rights by trying to open a file that is only available to admins.
	f, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false, errors.New("Elevated permissions are required to make entries to host file. Make sure you are running the command as an administrator.")
	}
	f.Close()
	return true, nil
}
