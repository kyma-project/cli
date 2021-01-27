package deploy

import (
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/pkg/asyncui"
	"github.com/kyma-project/cli/pkg/step"
)

// UIStepFactory is an implementation of installation.StepFactroy
type UIStepFactory struct {
	verbose bool
	asyncUI asyncui.AsyncUI
}

//NewUIStepFactory creates a UI step factory instance
func NewUIStepFactory(verbose bool, ui asyncui.AsyncUI) *UIStepFactory {
	return &UIStepFactory{
		verbose: verbose,
		asyncUI: ui,
	}
}

//AddStep defined in installation.StepFactroy interface
func (dsf *UIStepFactory) AddStep(stepName string) step.Step {
	step, err := dsf.asyncUI.AddStep(stepName)
	if err == nil {
		return step
	}
	step = NewUIStep(stepName, cli.LogFunc(dsf.verbose)) //use step which logs to console as fallback
	step.Start()
	return step
}
