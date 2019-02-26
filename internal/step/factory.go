package step

import "runtime"

type Factory struct {
	NonInteractive bool
}

func (f *Factory) NewStep(msg string) Step {
	if f.NonInteractive || runtime.GOOS != "darwin" {
		return NewSimpleStep(msg)
	}
	return NewStepWithSpinner(msg)
}
