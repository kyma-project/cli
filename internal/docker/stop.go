package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
)

func (c *Client) Stop(ctx context.Context, containerID string, log func(...interface{})) func() {
	return func() {
		log(fmt.Sprintf("\r- Removing container %s...\n", containerID))
		err := c.client.ContainerStop(ctx, containerID, container.StopOptions{})
		if err != nil {
			log(err)
		}
	}
}
