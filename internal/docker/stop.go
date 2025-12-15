package docker

import (
	"context"

	"github.com/docker/docker/api/types/container"
)

func (c *Client) Stop(ctx context.Context, containerID string) error {
	return c.client.ContainerStop(ctx, containerID, container.StopOptions{})
}
