package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/kyma-project/cli.v3/internal/out"
)

func (c *Client) ContainerFollowRun(containerID string, forwardOutput bool) error {
	buf, err := c.ContainerAttach(context.Background(), containerID, container.AttachOptions{
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
		dstout = out.Default.MsgWriter()
		dsterr = out.Default.ErrWriter()
	}

	c.stopContainerOnSigInt(
		containerID,
		dstout,
		dsterr,
		buf.Reader,
	)

	return nil
}
