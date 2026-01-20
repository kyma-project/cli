package docker

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
)

func Test_containerFollowRun(t *testing.T) {
	t.Run("forwards output when enabled", func(t *testing.T) {
		outBuf := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}

		utils := followRunUtils{
			containerAttach: func(
				ctx context.Context,
				containerID string,
				opts container.AttachOptions,
			) (types.HijackedResponse, error) {
				require.Equal(t, context.Background(), ctx)
				require.Equal(t, "container-id", containerID)
				require.True(t, opts.Stdout)
				require.True(t, opts.Stderr)
				require.True(t, opts.Stream)
				return types.HijackedResponse{
					Reader: bufio.NewReader(bytes.NewBufferString("test-output")),
					Conn:   fixTestConn(),
				}, nil
			},
			stopOnSigInt: func(containerID string, dstout, dsterr io.Writer, reader io.Reader) {
				_, _ = io.Copy(dstout, reader)
				_, _ = dsterr.Write([]byte("test-output"))
			},
			msgWriter: func() io.Writer { return outBuf },
			errWriter: func() io.Writer { return errBuf },
		}

		c := &Client{}
		err := c.containerFollowRun("container-id", true, utils)

		require.NoError(t, err)
		require.Equal(t, "test-output", outBuf.String())
		require.Equal(t, "test-output", errBuf.String())
	})

	t.Run("returns error when container attach fails", func(t *testing.T) {
		outBuf := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}

		utils := followRunUtils{
			containerAttach: func(
				ctx context.Context,
				containerID string,
				opts container.AttachOptions,
			) (types.HijackedResponse, error) {
				require.Equal(t, context.Background(), ctx)
				require.Equal(t, "container-id", containerID)
				require.True(t, opts.Stdout)
				require.True(t, opts.Stderr)
				require.True(t, opts.Stream)
				return types.HijackedResponse{}, io.ErrUnexpectedEOF
			},
			msgWriter: func() io.Writer { return outBuf },
			errWriter: func() io.Writer { return errBuf },
		}

		c := &Client{}
		err := c.containerFollowRun("container-id", true, utils)

		require.ErrorIs(t, err, io.ErrUnexpectedEOF)
		require.Empty(t, outBuf.String())
		require.Empty(t, errBuf.String())
	})

	t.Run("fails when one of the parameters is missing", func(t *testing.T) {
		utils := followRunUtils{
			containerAttach: func(
				ctx context.Context,
				containerID string,
				opts container.AttachOptions,
			) (types.HijackedResponse, error) {

				if containerID == "" ||
					ctx == nil ||
					!opts.Stdout ||
					!opts.Stderr ||
					!opts.Stream {
					return types.HijackedResponse{}, errors.New("missing parameter")
				}

				t.Fatal("containerAttach should not be called with missing parameters")
				return types.HijackedResponse{}, nil
			},
			msgWriter: func() io.Writer { return io.Discard },
			errWriter: func() io.Writer { return io.Discard },
		}

		c := &Client{}
		err := c.containerFollowRun("", true, utils)

		require.Error(t, err)
		require.Contains(t, err.Error(), "missing parameter")
	})

}

func fixTestConn() net.Conn {
	srv, cl := net.Pipe()
	defer srv.Close()
	return cl
}
