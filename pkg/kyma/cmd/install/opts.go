package install

import (
	"strings"
	"time"

	"github.com/kyma-project/cli/pkg/kyma/core"
)

//DefaultKymaVersion contains the default Kyma version to be installed in case another version is not specified
var DefaultKymaVersion string

type overrideFileList []string

func (ovf *overrideFileList) String() string {
	return "[" + strings.Join(*ovf, " ,") + "]"
}

func (ovf *overrideFileList) Set(value string) error {
	*ovf = append(*ovf, value)
	return nil
}

func (ovf *overrideFileList) Type() string {
	return "[]string"
}

func (ovf *overrideFileList) Len() int {
	return len(*ovf)
}

//Options defines available options for the command
type Options struct {
	*core.Options
	ReleaseVersion        string
	ReleaseConfig         string
	NoWait                bool
	Domain                string
	Local                 bool
	LocalSrcPath          string
	LocalInstallerVersion string
	LocalInstallerDir     string
	Timeout               time.Duration
	Password              string
	OverrideConfigs       overrideFileList
}

//NewOptions creates options with default values
func NewOptions(o *core.Options) *Options {
	return &Options{Options: o}
}
