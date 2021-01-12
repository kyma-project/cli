package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kyma-project/cli/internal/cli"
)

var (
	defaultDomain         = "local.kyma.dev"
	defaultVersion        = "latest"
	kymaProfiles          = []string{"evaluation", "production"}
	defaultWorkspacePath  = filepath.Join(".", "workspace")
	defaultResourcePath   = filepath.Join(defaultWorkspacePath, "kyma", "resources")
	defaultComponentsFile = filepath.Join(defaultWorkspacePath, "kyma", "installation", "resources", "components.yaml")
)

//Options defines available options for the command
type Options struct {
	*cli.Options
	WorkspacePath      string
	ResourcePath       string
	OverridesFile      string
	ComponentsListFile string
	CancelTimeout      time.Duration
	QuitTimeout        time.Duration
	HelmTimeout        time.Duration
	WorkersCount       int
	Domain             string
	TLSCert            string
	TLSKey             string
	Version            string
	Profile            string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

//GetProfiles return the currently supported profiles
func (o *Options) getProfiles() []string {
	return kymaProfiles
}

//GetDefaultDomain return the default domain
func (o *Options) getDefaultDomain() string {
	return defaultDomain
}

//getDefaultVersion return the default Kyma version
func (o *Options) getDefaultVersion() string {
	return defaultVersion
}

//getDefaultResourcesPath return the default path to the Kyma resources directory
func (o *Options) getDefaultResourcePath() string {
	return defaultResourcePath
}

//getDefaultWorkspacePath return the default path to the CLI workspace directory
func (o *Options) getDefaultWorkspacePath() string {
	return defaultWorkspacePath
}

//getDefaultComponentsFile return the default path to the Kyma components file
func (o *Options) getDefaultComponentsListFile() string {
	return defaultComponentsFile
}

func (o *Options) isSupportedProfile(profile string) bool {
	for _, supportedProfile := range kymaProfiles {
		if supportedProfile == profile {
			return true
		}
	}
	return false
}

// ValidateFlags applies a sanity check on provided options
func (o *Options) validateFlags() error {
	if err := o.pathExists(o.ResourcePath, "Resource path"); err != nil {
		return err
	}
	// Overrides file is optional, but if provided it has to exist
	if o.OverridesFile != "" {
		if err := o.pathExists(o.OverridesFile, "Overrides file"); err != nil {
			return err
		}
	}
	if o.QuitTimeout < o.CancelTimeout {
		return fmt.Errorf("Quit timeout (%v) cannot be smaller than cancel timeout (%v)", o.QuitTimeout, o.CancelTimeout)
	}
	if o.Profile != "" && !o.isSupportedProfile(o.Profile) {
		return fmt.Errorf("Profile unknown or not supported. Supported profiles are: %s", strings.Join(o.getProfiles(), ", "))
	}
	if o.Domain != defaultDomain && !o.tlsCertAndKeyProvided() {
		return fmt.Errorf("To use a custom domain name also a custom TLS certificate and TLS key has to be provided")
	}
	if (o.TLSKey != "" || o.TLSCert != "") && !o.tlsCertAndKeyProvided() {
		return fmt.Errorf("To use a custom TLS certificate the TLS certificate and TLS key has to be provided")
	}
	return nil
}

func (o *Options) tlsCertAndKeyProvided() bool {
	return o.TLSCert != "" && o.TLSKey != ""
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
