package step

type Factory struct {
	NonInteractive bool
}

func (f *Factory) NewStep(msg string) Step {
	if f.NonInteractive {
		return NewSimpleStep(msg)
	}
	return NewStepWithSpinner(msg)
}
