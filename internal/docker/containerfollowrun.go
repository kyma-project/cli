package docker

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

func (c *Client) ContainerFollowRun(ctx context.Context, containerID string, forwardOutput bool) error {
	buf, err := c.client.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stdout: true,
		Stderr: true,
		Stream: true,
	})
	if err != nil {
		return err
	}
	defer buf.Close()

	dstout, dsterr := io.Discard, io.Discard
	if forwardOutput {
		dstout, dsterr = os.Stdout, os.Stderr
	}

	_, err = stdcopy.StdCopy(dstout, dsterr, buf.Reader)

	return err
}
