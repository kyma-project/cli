package dashboard

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/mount"
	"github.com/kyma-project/cli/internal/kube"
	cliOS "github.com/kyma-project/cli/internal/os"
	"github.com/kyma-project/cli/pkg/docker"
	"github.com/pkg/browser"
)

const (
	containerKubeconfigFile = "config.yaml" // this is the kubeconfig filename in the container
	image                   = "europe-docker.pkg.dev/kyma-project/prod/kyma-dashboard-local-prod:latest"
)

// Container is a wrapper around the kyma dashboard docker container, providing an easy to use API to manage the kyam dashboard.
type Container struct {
	name              string
	id                string
	port              string
	kubeconfigPath    string
	kubeconfigMounted bool
	docker            docker.Wrapper
	verbose           bool
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

// Start runs the dashboard conrainer.
func (c *Container) Start() error {
	var err error
	if c.docker, err = docker.NewWrapper(); err != nil {
		return fmt.Errorf("could not create docker client: %w", err)
	}

	var envs []string
	if dockerDesktop, err := c.docker.IsDockerDesktopOS(context.Background()); err != nil {
		return fmt.Errorf("failed to interact with docker: %w", err)
	} else if dockerDesktop {
		dockerEnv := "DOCKER_DESKTOP_CLUSTER=true"
		envs = append(envs, dockerEnv)
		if c.verbose {
			fmt.Printf("Docker installation seems to be Docker Desktop. Appending '%s' to env variables of container\n", dockerEnv)
		}
	}

	if c.verbose {
		// when NODE_ENV is set to "development", all kind of logs are printed from the kyma-dashboard container
		envs = append(envs, "NODE_ENV=development")
	} else {
		envs = append(envs, "NODE_ENV=production")
	}
	if err != nil {
		return fmt.Errorf("failed to interact with docker: %w", err)
	}

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
		return fmt.Errorf("Dashboard at %q could not be opened: %w", url, err)
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

func (c *Container) containerOpts(envs []string) docker.ContainerRunOpts {
	containerRunOpts := docker.ContainerRunOpts{
		Envs:          envs,
		ContainerName: c.name,
		Image:         image,
		Ports: map[string]string{
			"3001": c.port,
		},
	}

	kubeconfigPath := kube.KubeconfigPath(c.kubeconfigPath)

	if _, err := kube.NewFromConfig("", kubeconfigPath); err != nil {
		if c.verbose {
			fmt.Printf("WARNING: Could not identify the current kubeconfig. "+
				"The Dashboard will start without the current cluster context. More details: %s", err)
		}
	} else {
		containerRunOpts.Mounts = []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: kubeconfigPath,
				Target: fmt.Sprintf("/app/core-ui/kubeconfig/%s", containerKubeconfigFile),
			},
		}
		c.kubeconfigMounted = true
	}

	if cliOS.IsLinux() {
		if c.verbose {
			fmt.Printf("Operating system seems to be Linux. Changing the Docker network mode to 'host'")
		}
		containerRunOpts.NetworkMode = "host"
	}

	return containerRunOpts
}
