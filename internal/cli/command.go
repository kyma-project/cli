package cli

import (
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/step"
)

type Command struct {
	*Options
	CurrentStep step.Step
	K8s         kube.KymaKube
}

func (c *Command) NewStep(msg string) step.Step {
	s := c.Factory.NewStep(msg)
	c.CurrentStep = s
	return s
}
