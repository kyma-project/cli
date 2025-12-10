package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kyma-project/cli.v3/internal/out"
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
		dstout = out.New()
		dsterr = out.New()
	}

	_, err = stdcopy.StdCopy(dstout, dsterr, buf.Reader)

	return err
}
