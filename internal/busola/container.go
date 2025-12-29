package busola

import (
	"context"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/docker"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/pkg/browser"
)

const (
	dashboardImage = "europe-docker.pkg.dev/kyma-project/prod/kyma-dashboard-local-prod:latest"
)

// Container is a wrapper around the kyma dashboard docker container, providing an easy to use API to manage the kyam dashboard.
type Container struct {
	name    string
	id      string
	port    string
	docker  *docker.Client
	verbose bool
}

// New creates a new dashboard container with the given configuration
func New(name, port string, verbose bool) (*Container, error) {
	dockerClient, err := docker.NewClient()
	if err != nil {
		return nil, fmt.Errorf("could not create docker client: %w", err)
	}

	return &Container{
		name:    name,
		port:    port,
		docker:  dockerClient,
		verbose: verbose,
	}, nil
}

// Start runs the dashboard container.
func (c *Container) Start() error {
	var envs []string

	opts := c.containerOpts(envs)
	out.Msg("\n")

	var err error
	if c.id, err = c.docker.PullImageAndStartContainer(context.Background(), opts); err != nil {
		return fmt.Errorf("unable to start container: %w", err)
	}
	return nil
}

// Opens the kyma dashboard in a browser.
func (c *Container) Open(path string) error {
	url := fmt.Sprintf("http://localhost:%s%s/clusters", c.port, path)

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

func (c *Container) containerOpts(envs []string) docker.ContainerRunOpts {
	containerRunOpts := docker.ContainerRunOpts{
		Envs:          envs,
		ContainerName: c.name,
		Image:         dashboardImage,
		Ports: map[string]string{
			"3001": c.port,
		},
	}

	return containerRunOpts
}
