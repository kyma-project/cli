package docker

import (
	"context"
	"io"
	"os"
	"os/signal"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
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

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	done := make(chan struct{})

	go func() {
		_, _ = stdcopy.StdCopy(dstout, dsterr, buf.Reader)
		close(done)
	}()

	c.stopContainerOnSigInt(
		containerID,
		dstout,
		dsterr,
		buf.Reader,
	)

	return nil
}
