// package files provides all functionality to manage kyma CLI local files.
package files

import (
	"os"
	"os/user"
	"path/filepath"
)

const kymaFolder = ".kyma"

func KymaHome() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	p := filepath.Join(u.HomeDir, kymaFolder)

	if _, err := os.Stat(p); os.IsNotExist(err) {
		err = os.MkdirAll(p, 0700)
		if err != nil {
			return "", err
		}
	}
	return p, nil
}
