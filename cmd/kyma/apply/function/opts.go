package function

import (
	"github.com/kyma-project/cli/internal/cli"
)

//Options defines available options for the command
type Options struct {
	*cli.Options

	OnError  OnError
	Output   Output
	Filename string
	DryRun   bool
}

type OnError = string

const (
	NothingOnError OnError = "nothing"
	PurgeOnError   OnError = "purge"
)

var validOnError = []string{
	NothingOnError,
	PurgeOnError,
}

type Output = string

const (
	NoneOutput Output = "none"
	JSONOutput Output = "json"
	YAMLOutput Output = "yaml"
	TextOutput Output = "text"
)

var validOutput = []string{
	NoneOutput,
	JSONOutput,
	YAMLOutput,
	TextOutput,
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
