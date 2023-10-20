package function

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/hydroform/function/pkg/workspace"
)

// Options defines available options for the command
type Options struct {
	*cli.Options

	OnError  value
	Output   value
	Filename string
	DryRun   bool
	Watch    bool
	Timeout  time.Duration
}

// NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{
		Options: o,
		OnError: newValue(NothingOnError, validOnError),
		Output:  newValue(TextOutput, validOutput),
	}
}

var (
	validOnError = []string{
		NothingOnError,
		PurgeOnError,
	}
	validOutput = []string{
		NoneOutput,
		JSONOutput,
		YAMLOutput,
		TextOutput,
	}
)

const (
	NothingOnError = "nothing"
	PurgeOnError   = "purge"
)

const (
	NoneOutput = "none"
	JSONOutput = "json"
	YAMLOutput = "yaml"
	TextOutput = "text"
)

type value struct {
	value     string
	available []string
}

func newValue(defaultVal string, available []string) value {
	return value{
		value:     defaultVal,
		available: available,
	}
}

func (g *value) String() string {
	return g.value
}

func (g *value) Set(v string) error {
	if g == nil {
		return fmt.Errorf("nil pointer reference")
	}
	if v == "" {
		return nil
	} else if g.isAllowed(v) {
		g.value = v
		return nil
	}
	return fmt.Errorf("specified value: %s is not supported", v)
}

func (g value) isAllowed(item string) bool {
	for _, elem := range g.available {
		if elem == item {
			return true
		}
	}
	return false
}

func (g value) Type() string {
	return "value"
}

func defaultFilename() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(pwd, workspace.CfgFilename), nil
}
