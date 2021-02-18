package deploy

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kyma-project/cli/internal/cli"
)

var (
	localSource           = "local"
	defaultSource         = "master"
	kymaProfiles          = []string{"evaluation", "production"}
	defaultWorkspacePath  = filepath.Join(".", "workspace")
	defaultComponentsFile = filepath.Join(defaultWorkspacePath, "installation", "resources", "components.yaml")
)

//Options defines available options for the command
type Options struct {
	*cli.Options
	WorkspacePath  string
	ComponentsFile string
	OverridesFile  string
	Overrides      []string
	CancelTimeout  time.Duration
	QuitTimeout    time.Duration
	HelmTimeout    time.Duration
	WorkersCount   int
	Domain         string
	TLSCrtFile     string
	TLSKeyFile     string
	Source         string
	Profile        string
	Atomic         bool
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

func (o *Options) supportedProfile(profile string) bool {
	for _, supportedProfile := range kymaProfiles {
		if supportedProfile == profile {
			return true
		}
	}
	return false
}

//tlsCrtEnc returns the base64 encoded TLS certificate
func (o *Options) tlsCrtEnc() (string, error) {
	return o.readFileAndEncode(o.TLSCrtFile)
}

//tlsKeyEnc returns the base64 encoded TLS key
func (o *Options) tlsKeyEnc() (string, error) {
	return o.readFileAndEncode(o.TLSKeyFile)
}

func (o *Options) readFileAndEncode(file string) (string, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(content), nil
}

// ResolveComponentsFile resolves the components file path related to the configured workspace path
func (o *Options) ResolveComponentsFile() string {
	if (o.ComponentsFile == "") || (o.WorkspacePath != defaultWorkspacePath && o.ComponentsFile == defaultComponentsFile) {
		return filepath.Join(o.WorkspacePath, "installation", "resources", "components.yaml")
	}
	return o.ComponentsFile
}

// validateFlags applies a sanity check on provided options
func (o *Options) validateFlags() error {
	// Overrides file is optional, but if provided it has to exist
	if o.OverridesFile != "" {
		if err := o.pathExists(o.OverridesFile, "Overrides file"); err != nil {
			return err
		}
	}
	if o.QuitTimeout < o.CancelTimeout {
		return fmt.Errorf("Quit timeout (%v) cannot be smaller than cancel timeout (%v)", o.QuitTimeout, o.CancelTimeout)
	}
	if o.Profile != "" && !o.supportedProfile(o.Profile) {
		return fmt.Errorf("Profile unknown or not supported. Supported profiles are: %s", strings.Join(kymaProfiles, ", "))
	}
	certsProvided, err := o.tlsCertAndKeyProvided()
	if err != nil {
		return err
	}
	if o.Domain != "" && !certsProvided {
		return fmt.Errorf("To use a custom domain name also a custom TLS certificate and TLS key has to be provided")
	}
	return nil
}

//tlsCertAndKeyProvided verify that always both cert parameters are provided and pointing to files
func (o *Options) tlsCertAndKeyProvided() (bool, error) {
	if o.TLSKeyFile == "" && o.TLSCrtFile == "" {
		return false, nil
	}
	if err := o.pathExists(o.TLSKeyFile, "TLS key"); err != nil {
		return false, err
	}
	if err := o.pathExists(o.TLSCrtFile, "TLS certificate"); err != nil {
		return false, err
	}
	return true, nil
}

func (o *Options) pathExists(path string, description string) error {
	if path == "" {
		return fmt.Errorf("%s is empty", description)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("%s '%s' not found", description, path)
	}
	return nil
}
