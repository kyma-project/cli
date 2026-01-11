package docker

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/stretchr/testify/require"
)

func TestPullImageAndStartContainer(t *testing.T) {
	t.Run("successful pulling and starting the container", func(t *testing.T) {
		cli := &testClient{
			pullBody: validPullResponse(),
			createID: "test-container",
		}

		id, err := cli.PullImageAndStartContainer()

		require.NoError(t, err)
		require.Equal(t, "test-container", id)
	})

	t.Run("error while creating container", func(t *testing.T) {
		cli := &testClient{
			pullBody: validPullResponse(),
			createID: "",
		}

		_, err := cli.PullImageAndStartContainer()

		require.ErrorContains(t, err, "missing container ID")
	})

	t.Run("error while starting container", func(t *testing.T) {
		cli := &testClient{
			pullBody: validPullResponse(),
			createID: "test-container",
			startErr: errors.New("start error"),
		}

		_, err := cli.PullImageAndStartContainer()

		require.ErrorContains(t, err, "start error")
	})
}

type testClient struct {
	pullBody io.ReadCloser
	createID string
	startErr error
}

func (c *testClient) PullImageAndStartContainer() (string, error) {
	r := c.pullBody
	if r == nil {
		r = io.NopCloser(bytes.NewReader(nil))
	}
	defer r.Close()

	buf := bytes.NewBuffer(nil)
	_ = jsonmessage.DisplayJSONMessagesToStream(r, streams.NewOut(buf), nil)

	if c.createID == "" {
		return "", errors.New("missing container ID")
	}

	if c.startErr != nil {
		return "", c.startErr
	}

	return c.createID, nil
}

func validPullResponse() io.ReadCloser {
	buf := bytes.NewBuffer(nil)
	_ = jsonmessage.DisplayJSONMessagesToStream(
		io.NopCloser(bytes.NewReader([]byte(`{"status":"pulling"}`+"\n"))),
		streams.NewOut(buf),
		nil,
	)
	return io.NopCloser(bytes.NewReader(buf.Bytes()))
}
