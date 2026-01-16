package docker

import (
	"bytes"
	"context"
	"io"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_stopContainerOnSigInt(t *testing.T) {
	t.Run("container stops when signal is received", func(t *testing.T) {
		c := &Client{}

		utils := stopUtils{
			waitForSignal: func() <-chan os.Signal {
				ch := make(chan os.Signal, 1)
				ch <- syscall.SIGINT
				return ch
			},
			stopContainer: func(ctx context.Context, containerID string) error {
				require.Equal(t, "test-container", containerID)
				return nil
			},
			stdCopy: func(dstout, dsterr io.Writer, src io.Reader) (int64, error) {
				return 0, nil
			},
		}

		c.stopContainerOnSigInt(
			"test-container",
			&bytes.Buffer{},
			&bytes.Buffer{},
			bytes.NewBuffer(nil),
			utils,
		)
	})

	t.Run("function finishes without receiving signal", func(t *testing.T) {
		c := &Client{}

		utils := stopUtils{
			waitForSignal: func() <-chan os.Signal {
				return make(chan os.Signal)
			},
			stopContainer: func(ctx context.Context, containerID string) error {
				t.Fatalf("stopContainer must not be called")
				return nil
			},
			stdCopy: func(dstout, dsterr io.Writer, src io.Reader) (int64, error) {
				return 0, nil
			},
		}

		c.stopContainerOnSigInt(
			"test-container",
			&bytes.Buffer{},
			&bytes.Buffer{},
			bytes.NewBuffer(nil),
			utils,
		)
	})
}
