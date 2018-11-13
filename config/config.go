package config

import (
	"fmt"
	"os"
	"os/user"
)

const (
	KYMA_FOLDER = ".kyma"
)

func Dir() string {
	u, err := user.Current()
	if err != nil {
		fmt.Printf("Could not get user: %s", err)
		return ""
	}
	configFolder := u.HomeDir + "/" + KYMA_FOLDER
	if _, err := os.Stat(configFolder); os.IsNotExist(err) {
		if err = os.Mkdir(configFolder, 0700); err != nil {
			fmt.Printf("Could not create Kyma folder: %s", err)
		}
	}
	return configFolder
}
