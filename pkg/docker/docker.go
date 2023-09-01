package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	dockerConfig "github.com/docker/cli/cli/config"
	configTypes "github.com/docker/cli/cli/config/types"
	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/kyma-project/cli/pkg/step"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	defaultRegistry = "index.docker.io"
)

type dockerClient struct {
	*docker.Client
}

type dockerWrapper struct {
	Docker Client
}

//go:generate mockery --name Client
type Client interface {
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig,
		platform *specs.Platform, containerName string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
	ContainerAttach(ctx context.Context, container string, options types.ContainerAttachOptions) (types.HijackedResponse, error)
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
	ImagePull(ctx context.Context, refStr string, options types.ImagePullOptions) (io.ReadCloser, error)
	ImagePush(ctx context.Context, image string, options types.ImagePushOptions) (io.ReadCloser, error)
	Info(ctx context.Context) (types.Info, error)
	NegotiateAPIVersion(ctx context.Context)
}

type KymaClient interface {
	PushKymaInstaller(image string, currentStep step.Step) error
	BuildKymaInstaller(localSrcPath, imageName string) error
}

// Wrapper provides helper functions
type Wrapper interface {
	PullImageAndStartContainer(ctx context.Context, opts ContainerRunOpts) (string, error)
	ContainerFollowRun(ctx context.Context, containerID string, forwardOutput bool) error
	Stop(ctx context.Context, containerID string, log func(...interface{})) func()
	IsDockerDesktopOS(ctx context.Context) (bool, error)
}

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

// NewClient creates docker client using docker environment of the OS
func NewClient() (Client, error) {
	dClient, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return nil, err
	}
	return &dockerClient{
		dClient,
	}, nil
}

// NewWrapper creates a new wrapper around the docker client with helper funtions
func NewWrapper() (Wrapper, error) {
	dClient, err := NewClient()
	if err != nil {
		return nil, err
	}
	return &dockerWrapper{
		Docker: dClient,
	}, nil
}

// PullImageAndStartContainer creates, pulls and starts a container
func (w *dockerWrapper) PullImageAndStartContainer(ctx context.Context, opts ContainerRunOpts) (string, error) {
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
	r, err := w.Docker.ImagePull(ctx, config.Image, types.ImagePullOptions{})
	if err != nil {
		return "", err
	}
	defer r.Close()

	streamer := streams.NewOut(os.Stdout)
	if err = jsonmessage.DisplayJSONMessagesToStream(r, streamer, nil); err != nil {
		return "", err
	}

	body, err := w.Docker.ContainerCreate(ctx, config, hostConfig, nil, nil, opts.ContainerName)
	if err != nil {
		return "", err
	}

	err = w.Docker.ContainerStart(ctx, body.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", err
	}

	return body.ID, nil
}

// ContainerFollowRun attaches a connection to a container and logs the output
func (w *dockerWrapper) ContainerFollowRun(ctx context.Context, containerID string, forwardOutput bool) error {
	buf, err := w.Docker.ContainerAttach(ctx, containerID, types.ContainerAttachOptions{
		Stdout: true,
		Stderr: true,
		Stream: true,
	})
	if err != nil {
		return err
	}
	defer buf.Close()

	dstout, dsterr := io.Discard, io.Discard
	if forwardOutput {
		dstout, dsterr = os.Stdout, os.Stderr
	}

	_, err = stdcopy.StdCopy(dstout, dsterr, buf.Reader)

	return err
}

// Stop stops a container with additional logging
func (w *dockerWrapper) Stop(ctx context.Context, containerID string, log func(...interface{})) func() {
	return func() {
		log(fmt.Sprintf("\r- Removing container %s...\n", containerID))
		err := w.Docker.ContainerStop(ctx, containerID, container.StopOptions{})
		if err != nil {
			log(err)
		}
	}
}

func (w *dockerWrapper) IsDockerDesktopOS(ctx context.Context) (bool, error) {
	info, err := w.Docker.Info(ctx)
	if err != nil {
		return false, err
	}

	return info.OperatingSystem == "Docker Desktop", nil
}

func splitDockerDomain(name string) (domain, remainder string) {
	i := strings.IndexRune(name, '/')
	if i == -1 || (!strings.ContainsAny(name[:i], ".:") && name[:i] != "localhost") {
		domain, remainder = defaultRegistry, name
	} else {
		domain, remainder = name[:i], name[i+1:]
	}
	return
}

// resolve finds the Docker credentials to push an image.
func resolve(registry string) (*configTypes.AuthConfig, error) {
	cf, err := dockerConfig.Load(os.Getenv("DOCKER_CONFIG"))
	if err != nil {
		return nil, err
	}

	if registry == defaultRegistry {
		registry = "https://" + defaultRegistry + "/v1/"
	}

	cfg, err := cf.GetAuthConfig(registry)
	if err != nil {
		return nil, err
	}

	empty := configTypes.AuthConfig{}
	if cfg == empty {
		return &empty, nil
	}
	return &configTypes.AuthConfig{
		Username:      cfg.Username,
		Password:      cfg.Password,
		Auth:          cfg.Auth,
		IdentityToken: cfg.IdentityToken,
		RegistryToken: cfg.RegistryToken,
	}, nil
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
