package step

import (
	"runtime"
)

// FactoryInterface is an abstraction for step factory
type FactoryInterface interface {
	NewStep(msg string) Step
}

// Factory contains the option to determine the interactivity of a Step.
type Factory struct {
	NonInteractive bool
	UseLogger      bool
	MuteLogger     bool
}

// NewStep creates a new Step to print out the current status with or without a spinner.
func (f *Factory) NewStep(msg string) Step {
	if f.UseLogger {
		return newLogStep(msg)
	}
	if f.NonInteractive || runtime.GOOS != "darwin" {
		return newSimpleStep(msg)
	}
	if f.MuteLogger {
		return NewMutedStep()
	}
	return newStepWithSpinner(msg)
}
