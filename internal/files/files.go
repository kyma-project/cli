// package files provides all functionality to manage kyma CLI local files.
package files

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	fp "path/filepath"

	"github.com/pkg/errors"
)

const kyma_folder = ".kyma"

func KymaHome() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	p := filepath.Join(u.HomeDir, kyma_folder)

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

	filePath = fp.Join(kh, filePath)

	err = os.MkdirAll(fp.Dir(filePath), 0700)
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

	filePath = fp.Join(kh, filePath)
	return ioutil.ReadFile(filePath)
}
