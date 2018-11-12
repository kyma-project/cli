package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"

	"gopkg.in/yaml.v2"
)

const (
	KYMA_FOLDER  = ".kyma"
	CONTEXT_FILE = "ctx.yaml"
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

type ContextConfig struct {
	CTX      string            `yaml:"CTX,omitempty"`
	Contexts map[string]string `yaml:"contexts"`
}

func Context() (*ContextConfig, error) {
	ctxFile := Dir() + "/" + CONTEXT_FILE

	if _, err := os.Stat(ctxFile); os.IsNotExist(err) {
		return &ContextConfig{
			Contexts: make(map[string]string),
		}, nil // not an error, config not set yet
	}

	b, err := ioutil.ReadFile(ctxFile)
	if err != nil {
		return nil, err
	}

	var ctx *ContextConfig
	if err := yaml.Unmarshal(b, &ctx); err != nil {
		return nil, err
	}
	return ctx, nil
}

func SaveContext(ctx *ContextConfig) error {
	ctxFile := Dir() + "/" + CONTEXT_FILE

	b, err := yaml.Marshal(ctx)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(ctxFile, b, 0600)
}
