package docker

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContainerFollowRun(t *testing.T) {
	t.Run("function writes to provided writers", func(t *testing.T) {
		outBuf, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
		cli := &mockClient{
			outputMessage: "test-output",
			dstOut:        outBuf,
			dstErr:        errBuf,
		}

		err := cli.ContainerFollowRun(true)
		require.NoError(t, err)

		require.Equal(t, "test-output", outBuf.String())
		require.Equal(t, "test-output", errBuf.String())
	})

	t.Run("function discards expected output", func(t *testing.T) {
		outBuf, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
		cli := &mockClient{
			outputMessage: "test-output",
			dstOut:        outBuf,
			dstErr:        errBuf,
		}

		err := cli.ContainerFollowRun(false)
		require.NoError(t, err)

		require.Equal(t, "", outBuf.String())
		require.Equal(t, "", errBuf.String())
	})
}

type mockClient struct {
	outputMessage  string
	dstOut, dstErr io.Writer
}

func (m *mockClient) ContainerFollowRun(forwardOutput bool) error {
	var dstout, dsterr = io.Discard, io.Discard
	if forwardOutput && m.dstOut != nil && m.dstErr != nil {
		dstout, dsterr = m.dstOut, m.dstErr
	}

	_, _ = dstout.Write([]byte(m.outputMessage))
	_, _ = dsterr.Write([]byte(m.outputMessage))

	return nil
}
