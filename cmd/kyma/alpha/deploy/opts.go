package deploy

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/download"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/files"
)

const (
	quitTimeoutFactor = 1.25
)

var (
	localSource           = "local"
	defaultSource         = "main"
	kymaProfiles          = []string{"evaluation", "production"}
	defaultWorkspacePath  = getDefaultWorkspacePath()
	defaultComponentsFile = filepath.Join(defaultWorkspacePath, "installation", "resources", "components.yaml")
)

//Options defines available options for the command
type Options struct {
	*cli.Options
	WorkspacePath    string
	ComponentsFile   string
	Components       []string
	OverridesFiles   []string
	Overrides        []string
	Timeout          time.Duration
	TimeoutComponent time.Duration
	Concurrency      int
	Domain           string
	TLSCrtFile       string
	TLSKeyFile       string
	Source           string
	Profile          string
	Atomic           bool
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

//QuitTimeout returns the calculated duration of the installation quit timeout
func (o *Options) QuitTimeout() time.Duration {
	return time.Duration((o.Timeout.Seconds() * quitTimeoutFactor)) * time.Second
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

//ResolveWorkspacePath tries to resolve the Kyma source folder if --source=local is defined,
//otherwise the defined workspace path will be returned
func (o *Options) ResolveLocalWorkspacePath() string {
	//resolve local Kyma source directory only if user has not defined a custom workspace directory
	if o.Source == localSource && o.WorkspacePath == defaultWorkspacePath {
		//use Kyma sources stored in GOPATH (if they exist)
		goPath := os.Getenv("GOPATH")
		if goPath != "" {
			kymaPath := filepath.Join(goPath, "src", "github.com", "kyma-project", "kyma")
			if o.pathExists(kymaPath, "Local Kyma source directory") == nil {
				return kymaPath
			}
		}
	}
	//no Kyma sources found in GOPATH
	return o.WorkspacePath
}

//ResolveComponentsFile resolves the components file path relative to the workspace path or makes a remote file locally available
func (o *Options) ResolveComponentsFile() (string, error) {
	workspacePath := o.ResolveLocalWorkspacePath()
	if (o.ComponentsFile == "") || (workspacePath != defaultWorkspacePath && o.ComponentsFile == defaultComponentsFile) {
		return filepath.Join(workspacePath, "installation", "resources", "components.yaml"), nil
	}
	file, err := download.GetFile(o.ComponentsFile, o.workspaceTmpDir())
	logger := cli.NewLogger(o.Verbose)
	logger.Debug(fmt.Sprintf("Using component list file '%s'", file))
	return file, err
}

//ResolveOverridesFiles makes overrides files locally available
func (o *Options) ResolveOverridesFiles() ([]string, error) {
	files, err := download.GetFiles(o.OverridesFiles, o.workspaceTmpDir())
	if len(files) > 0 {
		logger := cli.NewLogger(o.Verbose)
		logger.Debug(fmt.Sprintf("Using override files '%s'", strings.Join(files, "', '")))
	}
	return files, err
}

// validateFlags applies a sanity check on provided options
func (o *Options) validateFlags() error {
	if o.Timeout < o.TimeoutComponent {
		return fmt.Errorf("Timeout (%v) cannot be smaller than component timeout (%v)", o.Timeout, o.TimeoutComponent)
	}
	if o.Profile != "" && !o.supportedProfile(o.Profile) {
		return fmt.Errorf("Profile unknown or not supported. Supported profiles are: %s", strings.Join(kymaProfiles, ", "))
	}
	if _, err := o.tlsCertAndKeyProvided(); err != nil {
		return err
	}
	if o.ComponentsFile != defaultComponentsFile && len(o.Components) > 0 {
		return fmt.Errorf(`Provide either "components-file" or "component" flag`)
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

func (o *Options) workspaceTmpDir() string {
	return filepath.Join(o.WorkspacePath, "tmp")
}

func getDefaultWorkspacePath() string {
	kymaHome, err := files.KymaHome()
	if err != nil {
		return filepath.Join(".kyma-sources")
	}
	return filepath.Join(kymaHome, "sources")
}
