package deploy

import (
	"fmt"
	"strings"
	"time"

	"github.com/kyma-project/cli/internal/cli"
)

var (
	defaultDomain  = "kyma.local"
	defaultVersion = "latest"
	kymaProfiles   = []string{"production", "evaluation"}
)

//Options defines available options for the command
type Options struct {
	*cli.Options
	OverridesYaml  string
	ComponentsYaml string
	ResourcesPath  string
	CancelTimeout  time.Duration
	QuitTimeout    time.Duration
	HelmTimeout    time.Duration
	WorkersCount   int
	Domain         string
	TLSCert        string
	TLSKey         string
	Source         string
	Profile        string
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
	if o.ResourcesPath == "" {
		return fmt.Errorf("Resources path cannot be empty")
	}
	if o.ComponentsYaml == "" {
		return fmt.Errorf("Components YAML cannot be empty")
	}
	if o.QuitTimeout < o.CancelTimeout {
		return fmt.Errorf("Quit timeout (%v) cannot be smaller than cancel timeout (%v)", o.QuitTimeout, o.CancelTimeout)
	}
	if !o.isSupportedProfile(o.Profile) {
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
