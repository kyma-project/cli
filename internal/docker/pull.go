package docker

import (
	"context"
	"io"

	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/go-connections/nat"
	"github.com/kyma-project/cli.v3/internal/out"
)

const (
	defaultRegistry = "index.docker.io"
)

// ErrorMessage is used to parse error messages coming from Docker
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

// PullImageAndStartContainer creates, pulls and starts a container
func (c *Client) PullImageAndStartContainer(ctx context.Context, opts ContainerRunOpts) (string, error) {
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

	var r io.ReadCloser
	//mozliwosc dlugiego pullowania, wskazanie writera do streams out
	r, err := c.client.ImagePull(ctx, config.Image, image.PullOptions{})
	if err != nil {
		return "", err
	}
	defer r.Close()

	streamer := streams.NewOut(out.New())
	if err = jsonmessage.DisplayJSONMessagesToStream(r, streamer, nil); err != nil {
		return "", err
	}

	body, err := c.client.ContainerCreate(ctx, config, hostConfig, nil, nil, opts.ContainerName)
	if err != nil {
		return "", err
	}

	err = c.client.ContainerStart(ctx, body.ID, container.StartOptions{})
	if err != nil {
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
