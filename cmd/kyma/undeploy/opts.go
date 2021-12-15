package undeploy

import (
	"github.com/kyma-project/cli/cmd/kyma/deploy"
	"github.com/kyma-project/cli/internal/cli"
)

const (
	VersionLocal      = "local"
	profileEvaluation = "evaluation"
	profileProduction = "production"
)

//Options defines available options for the command
type Options struct {
	*deploy.Options
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: deploy.NewOptions(o)}
}
