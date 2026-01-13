package docker

import (
	"context"
	"io"

	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/go-connections/nat"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

type ErrorMessage struct {
	Error string
}

type ContainerRunOpts struct {
	ContainerName string
	Envs          []string
	Image         string
	Mounts        []mount.Mount
	NetworkMode   string
	Ports         map[string]string
}

// Utils struct allows injecting dependencies for testing
type utils struct {
	imagePull       func(ctx context.Context, imageName string, opts image.PullOptions) (io.ReadCloser, error)
	containerCreate func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error)
	containerStart  func(ctx context.Context, containerID string, opts container.StartOptions) error
	displayJSON     func(reader io.Reader, outStream *streams.Out) error
}

// PullImageAndStartContainer is the public function used in production
func (c *Client) PullImageAndStartContainer(ctx context.Context, opts ContainerRunOpts) (string, error) {
	return pullImageAndStartContainer(ctx, opts, utils{
		imagePull:       c.ImagePull,
		containerCreate: c.ContainerCreate,
		containerStart:  c.ContainerStart,
		displayJSON: func(r io.Reader, outStream *streams.Out) error {
			return jsonmessage.DisplayJSONMessagesToStream(r, outStream, nil)
		},
	})
}

// PullImageAndStartContainer is the private function that can be tested with mocks
func pullImageAndStartContainer(ctx context.Context, opts ContainerRunOpts, u utils) (string, error) {
	config := &container.Config{
		Env:          opts.Envs,
		ExposedPorts: portSet(opts.Ports),
		Image:        opts.Image,
	}
	hostConfig := &container.HostConfig{
		PortBindings: portMap(opts.Ports),
		AutoRemove:   true,
		Mounts:       opts.Mounts,
		NetworkMode:  container.NetworkMode(opts.NetworkMode),
	}

	r, err := u.imagePull(ctx, config.Image, image.PullOptions{})
	if err != nil {
		return "", err
	}
	defer r.Close()

	streamer := streams.NewOut(out.New())
	if err = jsonmessage.DisplayJSONMessagesToStream(r, streamer, nil); err != nil {
		return "", err
	}

	body, err := u.containerCreate(ctx, config, hostConfig, nil, nil, opts.ContainerName)
	if err != nil {
		return "", err
	}

	if err := u.containerStart(ctx, body.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	return body.ID, nil
}

func portSet(ports map[string]string) nat.PortSet {
	portSet := nat.PortSet{}
	for from := range ports {
		portSet[nat.Port(from)] = struct{}{}
	}
	return portSet
}

func portMap(ports map[string]string) nat.PortMap {
	portMap := nat.PortMap{}
	for from, to := range ports {
		portMap[nat.Port(from)] = []nat.PortBinding{
			{
				HostPort: to,
			},
		}
	}
	return portMap
}
