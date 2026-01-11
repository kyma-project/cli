package docker

import (
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kyma-project/cli.v3/internal/out"
)

func (c *Client) stopContainerOnSigInt(
	containerID string,
	dstout io.Writer,
	dsterr io.Writer,
	reader io.Reader,
) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	done := make(chan struct{})

	go func() {
		_, _ = stdcopy.StdCopy(dstout, dsterr, reader)
		close(done)
	}()

	select {
	case <-sigCh:
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := c.ContainerStop(stopCtx, containerID, container.StopOptions{}); err != nil {
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

func (c *Client) StopContainerOnSigInt(sigCh chan os.Signal) error {
	select {
	case <-sigCh:
		_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return nil
	default:
		return errors.New("no interrupt signal received")
	}
}
