package docker

import (
	"context"
	"io"
	"os"

	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/go-connections/nat"
	"github.com/kyma-incubator/hydroform/function/pkg/docker"
)

type RunOpts struct {
	ContainerName string
	Envs          []string
	Image         string
	Ports         map[string]string
}

// RunContainer is a helper function for the docker client that pulls and runs a container
func RunContainer(ctx context.Context, cli *client.Client, opts RunOpts) (string, error) {
	body, err := pullAndRun(ctx, cli, &container.Config{
		Env:          opts.Envs,
		ExposedPorts: portSet(opts.Ports),
		Image:        opts.Image,
	}, &container.HostConfig{
		PortBindings: portMap(opts.Ports),
		AutoRemove:   true,
	}, opts.ContainerName)
	if err != nil {
		return "", err
	}

	err = cli.ContainerStart(ctx, body.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", err
	}

	return body.ID, nil
}

func pullAndRun(ctx context.Context, c *client.Client, config *container.Config, hostConfig *container.HostConfig, containerName string) (container.ContainerCreateCreatedBody, error) {
	body, err := c.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if client.IsErrNotFound(err) {
		var r io.ReadCloser
		r, err = c.ImagePull(ctx, config.Image, types.ImagePullOptions{})
		if err != nil {
			return body, err
		}
		defer r.Close()

		streamer := streams.NewOut(os.Stdout)
		if err = jsonmessage.DisplayJSONMessagesToStream(r, streamer, nil); err != nil {
			return body, err
		}

		body, err = c.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	}
	return body, err
}

func FollowRun(ctx context.Context, c *client.Client, ID string, log func(...interface{})) error {
	return docker.FollowRun(ctx, c, ID, log)
}

func Stop(ctx context.Context, c *client.Client, ID string, log func(...interface{})) func() {
	return docker.Stop(ctx, c, ID, log)
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

func IsDockerDesktopOS(ctx context.Context, c *client.Client) (bool, error) {
	info, err := c.Info(ctx)
	if err != nil {
		return false, err
	}

	return info.OperatingSystem == "Docker Desktop", nil
}
