package docker

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

func TestPullImageAndStartContainer(t *testing.T) {
	t.Run("successful pulling and container start", func(t *testing.T) {
		utils := utils{
			imagePull: func(ctx context.Context, imageName string, opts image.PullOptions) (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte(""))), nil
			},
			containerCreate: func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
				return container.CreateResponse{ID: "test-container"}, nil
			},
			containerStart: func(ctx context.Context, containerID string, opts container.StartOptions) error {
				return nil
			},
			displayJSON: func(r io.Reader, outStream *streams.Out) error {
				return jsonmessage.DisplayJSONMessagesToStream(r, outStream, nil)
			},
		}

		opts := ContainerRunOpts{
			ContainerName: "test",
			Image:         "test-image",
		}

		id, err := pullImageAndStartContainer(context.Background(), opts, utils)
		require.NoError(t, err)
		require.Equal(t, "test-container", id)
	})

	t.Run("error while pulling image", func(t *testing.T) {
		utils := utils{
			imagePull: func(ctx context.Context, imageName string, opts image.PullOptions) (io.ReadCloser, error) {
				return nil, errors.New("pull error")
			},
			containerCreate: func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
				return container.CreateResponse{}, nil
			},
			containerStart: func(ctx context.Context, containerID string, opts container.StartOptions) error { return nil },
			displayJSON:    func(r io.Reader, outStream *streams.Out) error { return nil },
		}

		opts := ContainerRunOpts{ContainerName: "test", Image: "test-image"}

		_, err := pullImageAndStartContainer(context.Background(), opts, utils)
		require.ErrorContains(t, err, "pull error")
	})

	t.Run("error while creating container", func(t *testing.T) {
		utils := utils{
			imagePull: func(ctx context.Context, imageName string, opts image.PullOptions) (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte(""))), nil
			},
			containerCreate: func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
				return container.CreateResponse{}, errors.New("create error")
			},
			containerStart: func(ctx context.Context, containerID string, opts container.StartOptions) error { return nil },
			displayJSON:    func(r io.Reader, outStream *streams.Out) error { return nil },
		}

		opts := ContainerRunOpts{ContainerName: "test", Image: "test-image"}

		_, err := pullImageAndStartContainer(context.Background(), opts, utils)
		require.ErrorContains(t, err, "create error")
	})

	t.Run("error while starting container", func(t *testing.T) {
		utils := utils{
			imagePull: func(ctx context.Context, imageName string, opts image.PullOptions) (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte(""))), nil
			},
			containerCreate: func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
				return container.CreateResponse{ID: "test-container"}, nil
			},
			containerStart: func(ctx context.Context, containerID string, opts container.StartOptions) error {
				return errors.New("start error")
			},
			displayJSON: func(r io.Reader, outStream *streams.Out) error { return nil },
		}

		opts := ContainerRunOpts{ContainerName: "test", Image: "test-image"}

		_, err := pullImageAndStartContainer(context.Background(), opts, utils)
		require.ErrorContains(t, err, "start error")
	})
}
