package busola

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/docker"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/pkg/browser"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	dashboardImage = "europe-docker.pkg.dev/kyma-project/prod/busola:latest"
)

// ContainerRunner is a wrapper around the kyma dashboard docker container, providing an easy to use API to manage the kyam dashboard.
type ContainerRunner struct {
	name    string
	id      string
	port    string
	docker  *docker.Client
	verbose bool
}

// New creates a new dashboard container with the given configuration
func New(name, port, id string, verbose bool) (*ContainerRunner, error) {
	dockerClient, err := docker.NewClient()
	if err != nil {
		return nil, fmt.Errorf("could not create docker client: %w", err)
	}

	return &ContainerRunner{
		name:    name,
		port:    port,
		docker:  dockerClient,
		id:      id,
		verbose: verbose,
	}, nil
}

// Start runs the dashboard container.
func (c *ContainerRunner) Start(apiConfig *api.Config) error {
	var envs []string
	tmpDir := filepath.Join(os.TempDir(), "busola", c.id)

	err := os.MkdirAll(tmpDir, 0700)
	if err != nil {
		return fmt.Errorf("failed to create temp dir %q: %w", tmpDir, err)
	}

	config, err := clientcmd.Write(*apiConfig)
	if err != nil {
		return fmt.Errorf("failed to serialize kubeconfig: %w", err)
	}

	kubeconfigPath := filepath.Join(tmpDir, "config")
	if err := os.WriteFile(kubeconfigPath, config, 0700); err != nil {
		return fmt.Errorf("failed to write kubeconfig at %q: %w", kubeconfigPath, err)
	}

	opts := c.containerOpts(envs)
	out.Msg("\n")

	if c.id, err = c.docker.PullImageAndStartContainer(context.Background(), opts); err != nil {
		return fmt.Errorf("unable to start container: %w", err)
	}
	return nil
}

// Open opens the kyma dashboard in a browser.
func (c *ContainerRunner) Open() error {
	url := fmt.Sprintf("http://localhost:%s/clusters", c.port)

	err := browser.OpenURL(url)
	if err != nil {
		return fmt.Errorf("dashboard at %q could not be opened: %w", url, err)
	}
	return nil
}

// Stop stops kyma dashboard container and removes its temporary kubeconfig folder.
func (c *ContainerRunner) Stop(cfg *cmdcommon.KymaConfig, containerName string) clierror.Error {
	cli, err := docker.NewClient()
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to initialize docker client"))
	}

	inspectResponse, err := cli.ContainerInspect(cfg.Ctx, containerName)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to inspect container "+containerName))
	}

	err = cli.ContainerStop(cfg.Ctx, containerName, container.StopOptions{})
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to delete container "+containerName))
	}

	configPath := filepath.Join(os.TempDir(), "busola", inspectResponse.ID)

	err = os.RemoveAll(configPath)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to remove kubeconfig folder at tmp with container id "+inspectResponse.ID))
	}
	return nil
}

// Watch attaches to the running docker container, streams its output, and handles graceful shutdown on user interrupt.
func (c *ContainerRunner) Watch() error {
	return c.docker.ContainerFollowRun(c.id, c.verbose)
}

func (c *ContainerRunner) containerOpts(envs []string) docker.ContainerRunOpts {
	kubeconfigPath := filepath.Join(os.TempDir(), "busola", c.id, "config")
	targetPath := "/app/core-ui/kubeconfig/config.yaml"

	containerRunOpts := docker.ContainerRunOpts{
		Envs:          envs,
		ContainerName: c.name,
		Image:         dashboardImage,
		Mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   kubeconfigPath,
				Target:   targetPath,
				ReadOnly: true,
			},
		},
		Ports: map[string]string{
			"3001": c.port,
		},
	}

	return containerRunOpts
}

type ContainerStopper struct {
	cfg           *cmdcommon.KymaConfig
	containerName string
}

func NewStopper(cfg *cmdcommon.KymaConfig, containerName string) *ContainerStopper {
	return &ContainerStopper{
		cfg:           cfg,
		containerName: containerName,
	}
}

func (c *ContainerStopper) Stop(cfg *cmdcommon.KymaConfig) clierror.Error {
	cli, err := docker.NewClient()
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to initialize docker client"))
	}

	inspectResponse, err := cli.ContainerInspect(cfg.Ctx, c.containerName)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to inspect container "+c.containerName))
	}

	err = cli.ContainerStop(cfg.Ctx, c.containerName, container.StopOptions{})
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to delete container "+c.containerName))
	}

	configPath := filepath.Join(os.TempDir(), "busola", inspectResponse.ID)

	err = os.RemoveAll(configPath)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to remove kubeconfig folder at tmp with container id "+inspectResponse.ID))
	}
	return nil
}
