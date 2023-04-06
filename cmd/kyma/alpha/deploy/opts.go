package deploy

import (
	"fmt"
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kustomize"
	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

const (
	targetControlPlane = "control-plane"
	targetRemote       = "remote"
)

// Options defines available options for the command
type Options struct {
	*cli.Options
	LifecycleManager    string
	WildcardPermissions bool
	DryRun              bool
	OpenDashboard       bool
	Force               bool
	CertManagerVersion  string
	Namespace           string
	Channel             string
	KymaCR              string
	Target              string
	Modules             []string
	Kustomizations      []string
	AdditionalTemplates []string
	Timeout             time.Duration
	Filters             []kio.Filter
}

// NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

// validateFlags performs a sanity check of provided options
func (o *Options) validateFlags() error {
	if err := o.validateTimeout(); err != nil {
		return err
	}

	if err := o.validateFilters(); err != nil {
		return err
	}

	return o.validateTarget()
}

func (o *Options) validateTimeout() error {
	if o.Timeout <= 0 {
		return errors.New("timeout must be a positive duration")
	}
	return nil
}

// validateFilters sets up all filters that will be used by kustomize when running the command
func (o *Options) validateFilters() error {
	var filters []kio.Filter
	modifier, err := kustomize.LifecycleManagerImageModifier(
		o.LifecycleManager,
		func(image string) {
			o.NewStep(fmt.Sprintf("Used Lifecycle-Manager: %s", image)).Success()
		},
	)
	if err != nil {
		return err
	}
	filters = append(filters, modifier)
	o.Filters = filters
	return nil
}

func (o *Options) validateTarget() error {
	if o.Target == targetControlPlane || o.Target == targetRemote {
		return nil
	}

	return fmt.Errorf("target must be either '%s' or '%s'", targetControlPlane, targetRemote)
}
