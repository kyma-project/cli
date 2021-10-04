package install

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/kyma-project/cli/internal/cli"
)

//Options defines available options for the command
type Options struct {
	*cli.Options
	NoWait           bool
	Domain           string
	TLSCert          string
	TLSKey           string
	LocalSrcPath     string
	Timeout          time.Duration
	Password         string
	OverrideConfigs  []string
	ComponentsConfig string
	Source           string
	FallbackLevel    int
	CustomImage      string
	Profile          string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

// validateFlags performs a sanity check of provided options
func (o *Options) validateFlags() error {
	// validate --source
	checkFirstDigit := regexp.MustCompile(`^[1-9]\.`)
	startsWithDigit := checkFirstDigit.MatchString(o.Source)

	if !startsWithDigit {
		return nil
	}

	checkSemanticVersion := regexp.MustCompile(`[1-9]\.[0-9]+\.[0-9]+`)
	isSemVer := checkSemanticVersion.MatchString(o.Source)
	if isSemVer {
		if strings.HasPrefix(o.Source, "2") {
			return fmt.Errorf("Kyma version 2.x can not be installed via 'install'. Please use the 'deploy' command, which supports Kyma 2 versions")
		}
		return nil
	}
	return fmt.Errorf("Provided version (%s) is not a valid semantic version. It should be of format X.Y.Z", o.Source)

}
