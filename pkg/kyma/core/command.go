package core

import (
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/kubectl"
	"github.com/kyma-project/cli/internal/step"
)

type Command struct {
	*Options
	CurrentStep step.Step
	kubectl     *kubectl.Wrapper
	K8s         kube.KymaKube
}

func (c *Command) NewStep(msg string) step.Step {
	s := c.Factory.NewStep(msg)
	c.CurrentStep = s
	return s
}

func (c *Command) Kubectl() *kubectl.Wrapper {
	if c.kubectl == nil {
		c.kubectl = kubectl.NewWrapper(c.Verbose, c.Options.KubeconfigPath)
	}
	return c.kubectl
}
