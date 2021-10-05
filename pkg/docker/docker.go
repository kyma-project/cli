package docker

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	dockerConfig "github.com/docker/cli/cli/config"
	configTypes "github.com/docker/cli/cli/config/types"
	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/go-connections/nat"
	hydroformDocker "github.com/kyma-incubator/hydroform/function/pkg/docker"
	"github.com/kyma-project/cli/internal/minikube"
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

type kymaDockerClient struct {
	Docker Client
}

//go:generate mockery --name Client
type Client interface {
	ArchiveDirectory(srcPath string, options *archive.TarOptions) (io.ReadCloser, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig,
		platform *specs.Platform, containerName string) (container.ContainerCreateCreatedBody, error)
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
	ContainerAttach(ctx context.Context, container string, options types.ContainerAttachOptions) (types.HijackedResponse, error)
	ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error
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
	ContainerCreateAndStart(ctx context.Context, opts ContainerRunOpts) (string, error)
	ContainerFollowRun(ctx context.Context, containerID string, log func(...interface{})) error
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
	Ports         map[string]string
}

//NewClient creates docker client using docker environment of the OS
func NewClient() (Client, error) {
	dClient, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return nil, err
	}
	return &dockerClient{
		dClient,
	}, nil
}

//NewMinikubeClient creates docker client for minikube docker-env
func NewMinikubeClient(verbosity bool, profile string, timeout time.Duration) (Client, error) {
	dClient, err := minikube.DockerClient(verbosity, profile, timeout)
	if err != nil {
		return nil, err
	}

	return &dockerClient{
		dClient,
	}, nil
}

func NewKymaClient(isLocal bool, verbosity bool, profile string, timeout time.Duration) (KymaClient, error) {
	var err error
	var dc Client
	if isLocal {
		dc, err = NewMinikubeClient(verbosity, profile, timeout)
	} else {
		dc, err = NewClient()
	}
	return &kymaDockerClient{
		Docker: dc,
	}, err
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

func (d *dockerClient) ArchiveDirectory(srcPath string, options *archive.TarOptions) (io.ReadCloser, error) {
	return archive.TarWithOptions(srcPath, &archive.TarOptions{})
}

func (k *kymaDockerClient) BuildKymaInstaller(localSrcPath, imageName string) error {
	reader, err := k.Docker.ArchiveDirectory(localSrcPath, &archive.TarOptions{})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(300)*time.Second)
	defer cancel()
	k.Docker.NegotiateAPIVersion(ctx)
	args := make(map[string]*string)
	_, err = k.Docker.ImageBuild(
		ctx,
		reader,
		types.ImageBuildOptions{
			Tags:           []string{strings.TrimSpace(string(imageName))},
			SuppressOutput: true,
			Remove:         true,
			Dockerfile:     path.Join("tools", "kyma-installer", "kyma.Dockerfile"),
			BuildArgs:      args,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (k *kymaDockerClient) PushKymaInstaller(image string, currentStep step.Step) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(300)*time.Second)
	defer cancel()
	k.Docker.NegotiateAPIVersion(ctx)
	domain, _ := splitDockerDomain(image)
	auth, err := resolve(domain)
	if err != nil {
		return err
	}

	encodedJSON, err := json.Marshal(auth)
	if err != nil {
		return err
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	currentStep.LogInfof("Pushing Docker image: '%s'", image)

	pusher, err := k.Docker.ImagePush(ctx, image, types.ImagePushOptions{RegistryAuth: authStr})
	if err != nil {
		return err
	}

	defer pusher.Close()

	var errorMessage ErrorMessage
	buffIOReader := bufio.NewReader(pusher)

	for {
		streamBytes, err := buffIOReader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		err = json.Unmarshal(streamBytes, &errorMessage)
		if err != nil {
			return err
		}
		if errorMessage.Error != "" {
			if strings.Contains(errorMessage.Error, "unauthorized") || strings.Contains(errorMessage.Error, "requested access to the resource is denied") {
				return fmt.Errorf("missing permissions to push Docker image: %s\nPlease run `docker login` to authenticate", errorMessage.Error)
			}
			return fmt.Errorf("failed to push Docker image: %s", errorMessage.Error)
		}
	}

	return nil
}

// ContainerCreateAndStart creates, pulls (if necessary) and starts a container
func (w *dockerWrapper) ContainerCreateAndStart(ctx context.Context, opts ContainerRunOpts) (string, error) {
	config := &container.Config{
		Env:          opts.Envs,
		ExposedPorts: portSet(opts.Ports),
		Image:        opts.Image,
	}
	hostConfig := &container.HostConfig{
		PortBindings: portMap(opts.Ports),
		AutoRemove:   true,
	}

	body, err := w.Docker.ContainerCreate(ctx, config, hostConfig, nil, nil, opts.ContainerName)
	if docker.IsErrNotFound(err) {
		var r io.ReadCloser
		r, err = w.Docker.ImagePull(ctx, config.Image, types.ImagePullOptions{})
		if err != nil {
			return "", err
		}
		defer r.Close()

		streamer := streams.NewOut(os.Stdout)
		if err = jsonmessage.DisplayJSONMessagesToStream(r, streamer, nil); err != nil {
			return "", err
		}

		body, err = w.Docker.ContainerCreate(ctx, config, hostConfig, nil, nil, opts.ContainerName)
	}
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
func (w *dockerWrapper) ContainerFollowRun(ctx context.Context, containerID string, log func(...interface{})) error {
	return hydroformDocker.FollowRun(ctx, w.Docker, containerID, log)
}

// Stop stops a container with additional logging
func (w *dockerWrapper) Stop(ctx context.Context, containerID string, log func(...interface{})) func() {
	return hydroformDocker.Stop(ctx, w.Docker, containerID, log)
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
