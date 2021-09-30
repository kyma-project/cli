package deploy

import (
	"fmt"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/files"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"regexp"
)

var (
	defaultWorkspacePath = getDefaultWorkspacePath()
)

const (
	VersionLocal      = "local"
	profileEvaluation = "evaluation"
	profileProduction = "production"
)

//Options defines available options for the command
type Options struct {
	*cli.Options
	WorkspacePath  string
	Source         string
	Components     []string
	ComponentsFile string
	Domain         string
	Values         []string
	ValueFiles     []string
	Profile        string
	TLSCrtFile     string
	TLSKeyFile     string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

func (o *Options) ResolveLocalWorkspacePath() (string, error) {

	if o.WorkspacePath == "" {
		o.WorkspacePath = defaultWorkspacePath
	}
	//resolve local Kyma source directory only if user has not defined a custom workspace directory
	if o.Source == VersionLocal && o.WorkspacePath == defaultWorkspacePath {
		//use Kyma sources stored in GOPATH (if they exist)
		goPath := os.Getenv("GOPATH")
		if goPath != "" {
			kymaPath := filepath.Join(goPath, "src", "github.com", "kyma-project", "kyma")
			if o.pathExists(kymaPath, "Local Kyma source directory") == nil {
				return kymaPath, nil
			}
		}
	}

	if o.Source != VersionLocal {
		if err := os.RemoveAll(o.WorkspacePath); err != nil {
			return "", errors.Wrapf(err, "Could not delete old kyma source files in (%s)", o.WorkspacePath)
		}
	}

	//no Kyma sources found in GOPATH
	return o.WorkspacePath, nil
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

func getDefaultWorkspacePath() string {
	kymaHome, err := files.KymaHome()
	if err != nil {
		return ".kyma-sources"
	}
	return filepath.Join(kymaHome, "sources")
}

// validateFlags performs a sanity check of provided options
func (o *Options) validateFlags() error {
	if err := o.validateProfile(); err != nil {
		return err
	}
	if err := o.validateSource(); err != nil {
		return err
	}
	if err := o.validateTLSCertAndKey(); err != nil {
		return err
	}

	return nil
}

func (o *Options) validateSource() error {
	checkFirstDigit := regexp.MustCompile(`^[1-9]\.`)
	startsWithNo := checkFirstDigit.MatchString(o.Source)
	if startsWithNo {
		checkCompleteSource := regexp.MustCompile(`[1-9]\.[0-9]+\.[0-9]+`)
		isSemVer := checkCompleteSource.MatchString(o.Source)
		if isSemVer {
			return nil
		}
		return fmt.Errorf("provided version (%s) is not semver should be of format X.Y.Z", o.Source)
	}
	return nil
}

func (o *Options) validateProfile() error {
	if o.Profile == "" || o.Profile == profileEvaluation || o.Profile == profileProduction {
		return nil
	}

	return fmt.Errorf("unknown profile: %s", o.Profile)
}

func (o *Options) validateTLSCertAndKey() error {
	if o.TLSKeyFile == "" && o.TLSCrtFile == "" {
		return nil
	}
	if _, err := os.Stat(o.TLSKeyFile); os.IsNotExist(err) {
		return errors.New("tls key not found")
	}
	if _, err := os.Stat(o.TLSCrtFile); os.IsNotExist(err) {
		return errors.New("tls cert not found")
	}

	return nil
}
