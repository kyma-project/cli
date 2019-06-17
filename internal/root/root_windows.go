// +build windows

package root

import (
	"fmt"
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

func PromptUser() bool {
	for {
		fmt.Print("Type [y/n]: ")
		var res string
		if _, err := fmt.Scanf("%s", &res); err != nil {
			return false
		}
		switch res {
		case "yes", "y":
			return true
		case "no", "n":
			return false
		default:
			continue
		}
	}
}
