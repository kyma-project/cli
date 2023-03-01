package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/cli/internal/clusterinfo"
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

func (c *Command) EnsureClusterAccess(ctx context.Context, timeout time.Duration) (clusterinfo.Info, error) {
	if c.K8s == nil {
		var err error
		if c.K8s, err = kube.NewFromConfigWithTimeout("", c.KubeconfigPath, timeout); err != nil {
			return nil, fmt.Errorf("failed to initialize the Kubernetes client from given kubeconfig: %w", err)
		}
	}

	info, err := clusterinfo.Discover(ctx, c.K8s.Static())

	if err != nil {
		return nil, fmt.Errorf("failed to discover clusterinfo: %w", err)
	}

	return info, nil
}
