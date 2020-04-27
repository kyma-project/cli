// package files provides all functionality to manage kyma CLI local files.
package files

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/pkg/errors"
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

// Save saves the given content to the relative file path inside the kyma CLI local folder
func Save(filePath string, content []byte) error {
	kh, err := KymaHome()
	if err != nil {
		return errors.Wrap(err, "Could not save file")
	}

	filePath = filepath.Join(kh, filePath)

	err = os.MkdirAll(filepath.Dir(filePath), 0700)
	if err != nil {
		return errors.Wrap(err, "Could not save file")
	}

	if err = ioutil.WriteFile(filePath, content, 0644); err != nil {
		return errors.Wrap(err, "Could not save file")
	}
	return nil
}

// Load loads the content of the relative file path inside the kyma CLI local folder
func Load(filePath string) ([]byte, error) {
	kh, err := KymaHome()
	if err != nil {
		return nil, errors.Wrap(err, "Could not load file")
	}

	filePath = filepath.Join(kh, filePath)
	return ioutil.ReadFile(filePath)
}
