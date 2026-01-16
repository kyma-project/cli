package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/kyma-project/cli.v3/internal/out"
)

type followRunUtils struct {
	containerAttach func(
		ctx context.Context,
		containerID string,
		opts container.AttachOptions,
	) (types.HijackedResponse, error)

	stopOnSigInt func(
		containerID string,
		dstout io.Writer,
		dsterr io.Writer,
		reader io.Reader,
	)

	msgWriter func() io.Writer
	errWriter func() io.Writer
}

func (c *Client) ContainerFollowRun(containerID string, forwardOutput bool) error {
	return c.containerFollowRun(containerID, forwardOutput, followRunUtils{
		containerAttach: c.ContainerAttach,
		stopOnSigInt:    c.StopContainerOnSigInt,
		msgWriter:       out.Default.MsgWriter,
		errWriter:       out.Default.ErrWriter,
	})
}

func (c *Client) containerFollowRun(
	containerID string,
	forwardOutput bool,
	utils followRunUtils,
) error {
	buf, err := utils.containerAttach(
		context.Background(),
		containerID,
		container.AttachOptions{
			Stdout: true,
			Stderr: true,
			Stream: true,
		},
	)
	if err != nil {
		return err
	}

	defer buf.Close()

	dstout, dsterr := io.Discard, io.Discard
	if forwardOutput {
		dstout = utils.msgWriter()
		dsterr = utils.errWriter()
	}

	utils.stopOnSigInt(
		containerID,
		dstout,
		dsterr,
		buf.Reader,
	)

	return nil
}
