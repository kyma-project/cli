package docker

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kyma-project/cli.v3/internal/out"
)

func (c *Client) ContainerFollowRun(ctx context.Context, containerID string, forwardOutput bool) error {
	// Attach to the container
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

	done := make(chan struct{})

	go func() {
		_, _ = stdcopy.StdCopy(dstout, dsterr, buf.Reader)
		close(done)
	}()

	select {
	case <-sigCh:
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = c.ContainerStop(stopCtx, containerID, container.StopOptions{})

		<-done
	}

	return nil
}
