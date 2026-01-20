package docker

import (
	"context"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kyma-project/cli.v3/internal/out"
)

type stopUtils struct {
	waitForSignal func() <-chan os.Signal
	stopContainer func(ctx context.Context, containerID string) error
	stdCopy       func(dstout, dsterr io.Writer, src io.Reader) (written int64, err error)
}

func (c *Client) StopContainerOnSigInt(
	containerID string,
	dstout io.Writer,
	dsterr io.Writer,
	reader io.Reader,
) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	c.stopContainerOnSigInt(containerID, dstout, dsterr, reader, stopUtils{
		waitForSignal: func() <-chan os.Signal {
			return sigCh
		},
		stopContainer: func(ctx context.Context, containerID string) error {
			return c.ContainerStop(ctx, containerID, container.StopOptions{})
		},
		stdCopy: stdcopy.StdCopy,
	})
}

func (c *Client) stopContainerOnSigInt(
	containerID string,
	dstout io.Writer,
	dsterr io.Writer,
	reader io.Reader,
	utils stopUtils,
) {
	done := make(chan struct{})

	go func() {
		_, _ = utils.stdCopy(dstout, dsterr, reader)
		close(done)
	}()

	select {
	case <-utils.waitForSignal():
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := utils.stopContainer(stopCtx, containerID); err != nil {
			out.Default.Errfln(
				"Failed to stop the running container. The container may still be running.\n"+
					"You can try stopping it again using Kyma Dashboard.\n"+
					"Error: %v",
				err,
			)
		}
		<-done

	case <-done:
	}
}
