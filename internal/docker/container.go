package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/mount"
	"github.com/pkg/browser"
)

const (
	containerKubeconfigFile = "config.yaml" // this is the kubeconfig filename in the container
	dashboardImage          = "europe-docker.pkg.dev/kyma-project/prod/kyma-dashboard-local-prod:latest"
)

// Container is a wrapper around the kyma dashboard docker container, providing an easy to use API to manage the kyam dashboard.
type Container struct {
	name              string
	id                string
	port              string
	kubeconfigPath    string
	kubeconfigMounted bool
	docker            *Client
	verbose           bool
}

type ContainerRunOpts struct {
	ContainerName string
	Envs          []string
	Image         string
	Mounts        []mount.Mount
	NetworkMode   string
	Ports         map[string]string
}

// New creates a new dashboard container with the given configuration
func New(name, port, kubeconfigPath string, verbose bool) *Container {
	return &Container{
		name:           name,
		port:           port,
		kubeconfigPath: kubeconfigPath,
		verbose:        verbose,
	}
}

// Start runs the dashboard container.
func (c *Container) Start() error {
	var err error
	c.docker, err = NewClient()

	if err != nil {
		return fmt.Errorf("could not create docker client: %w", err)
	}

	var envs []string

	opts := c.containerOpts(envs)
	fmt.Print("\n")
	if c.id, err = c.docker.PullImageAndStartContainer(context.Background(), opts); err != nil {
		return fmt.Errorf("unable to start container: %w", err)
	}
	return nil
}

// Opens the kyma dashboard in a browser.
func (c *Container) Open(path string) error {
	url := fmt.Sprintf("http://localhost:%s%s", c.port, path)

	if c.kubeconfigMounted {
		url = fmt.Sprintf("%s?kubeconfigID=%s", url, containerKubeconfigFile)
	}

	err := browser.OpenURL(url)
	if err != nil {
		return fmt.Errorf("dashboard at %q could not be opened: %w", url, err)
	}
	return nil
}

// Watch attaches to the running docker container and forwards its output.
func (c *Container) Watch(ctx context.Context) error {
	return c.docker.ContainerFollowRun(ctx, c.id, c.verbose)
}

// StopFunc returns a function to stop the dashboard container with the given context and logging function.
func (c *Container) StopFunc(ctx context.Context, log func(...interface{})) func() {
	return c.docker.Stop(ctx, c.id, log)
}

func (c *Container) containerOpts(envs []string) ContainerRunOpts {
	containerRunOpts := ContainerRunOpts{
		Envs:          envs,
		ContainerName: c.name,
		Image:         dashboardImage,
		Ports: map[string]string{
			"3001": c.port,
		},
	}

	return containerRunOpts
}
