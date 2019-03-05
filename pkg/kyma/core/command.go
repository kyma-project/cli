package core

import (
	"github.com/kyma-incubator/kyma-cli/internal/kubectl"
	"github.com/kyma-incubator/kyma-cli/internal/step"
)

type Command struct {
	*Options
	CurrentStep step.Step
	kubectl *kubectl.Wrapper
}

func (c *Command) NewStep(msg string) step.Step {
	s := c.Factory.NewStep(msg)
	c.CurrentStep = s
	return s
}

func (c *Command) Kubectl() *kubectl.Wrapper {
	if c.kubectl == nil {
		c.kubectl = kubectl.NewWrapper(c.Verbose)
	}
	return c.kubectl
}
