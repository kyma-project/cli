package docker

import (
	"io"
	"testing"

	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/stretchr/testify/require"
)

func TestContainerFollowRun(t *testing.T) {
	t.Run("forwardOutput true writes to out.Default writers", func(t *testing.T) {
		cli := &mockClient{
			outputMessage: "test-output",
		}

		err := cli.ContainerFollowRun(true)
		require.NoError(t, err)
	})

	t.Run("forwardOutput false discards output", func(t *testing.T) {
		cli := &mockClient{
			outputMessage: "test-output",
		}

		err := cli.ContainerFollowRun(false)
		require.NoError(t, err)
	})
}

type mockClient struct {
	outputMessage string
}

func (m *mockClient) ContainerFollowRun(forwardOutput bool) error {
	var dstout, dsterr = io.Discard, io.Discard
	if forwardOutput {
		dstout = out.Default.MsgWriter()
		dsterr = out.Default.ErrWriter()
	}

	_, _ = dstout.Write([]byte(m.outputMessage))
	_, _ = dsterr.Write([]byte(m.outputMessage))

	return nil
}
