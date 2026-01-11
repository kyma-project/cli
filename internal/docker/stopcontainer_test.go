package docker

import (
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStopContainerOnSigInt(t *testing.T) {
	t.Run("container stops on interrupt", func(t *testing.T) {
		mock := &dockerClientMock{}
		cli := NewTestClient(mock)

		sigCh := make(chan os.Signal, 1)
		sigCh <- syscall.SIGINT

		err := cli.StopContainerOnSigInt(sigCh)

		require.NoError(t, err)
	})

	t.Run("function finishes without interrupt", func(t *testing.T) {
		mock := &dockerClientMock{}
		cli := NewTestClient(mock)

		sigCh := make(chan os.Signal)

		err := cli.StopContainerOnSigInt(sigCh)

		require.Error(t, err)
	})
}
