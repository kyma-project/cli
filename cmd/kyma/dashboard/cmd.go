package dashboard

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/docker/docker/api/types/mount"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/docker"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type command struct {
	opts *Options
	cli.Command
}

// NewCmd creates a new dashboard command
func NewCmd(o *Options) *cobra.Command {
	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Runs the Kyma dashboard locally and opens it directly in a web browser.",
		Long:  `Use this command to run the Kyma dashboard locally in a docker container and open it directly in a web browser. This command only works with a local installation of Kyma.`,
		RunE:  func(_ *cobra.Command, _ []string) error { return c.Run() },
	}

	cmd.Flags().StringVarP(&o.Port, "port", "p", "3001", `Specify the port on which the local dashboard will be exposed.`)
	cmd.Flags().StringVar(&o.ContainerName, "container-name", "busola", `Specify the name of the local container.`)

	return cmd
}

// Run runs the command
func (cmd *command) Run() error {
	var err error

	if err = cmd.opts.validateFlags(); err != nil {
		return err
	}
	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}
	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	return cmd.runDashboardContainer()
}

func (cmd *command) runDashboardContainer() error {
	step := cmd.NewStep(fmt.Sprintf("Starting container: %s:", cmd.opts.ContainerName))

	dockerWrapper, err := docker.NewWrapper()
	if err != nil {
		return errors.Wrap(err, "failed to interact with docker")
	}

	ctx := context.Background()

	var envs []string
	if dockerDesktop, err := dockerWrapper.IsDockerDesktopOS(ctx); err != nil {
		step.Failure()
		return errors.Wrap(err, "failed to interact with docker")
	} else if dockerDesktop {
		dockerEnv := "DOCKER_DESKTOP_CLUSTER=true"
		envs = append(envs, dockerEnv)
		if cmd.Verbose {
			fmt.Printf("Docker installation seems to be Docker Desktop. Appending '%s' to env variables of container\n", dockerEnv)
		}
	}

	if cmd.Verbose {
		// when NODE_ENV is set to "development", all kind of logs are printed from the busola container
		envs = append(envs, "NODE_ENV=development")
	}

	containerRunOpts, dashboardURL := cmd.initContainerRunOpts(envs)
	id, err := dockerWrapper.PullImageAndStartContainer(ctx, containerRunOpts)

	if err != nil {
		step.Failure()
		return errors.Wrap(err, "unable to start container")
	}
	step.Successf("Started container: %s", cmd.opts.ContainerName)

	cmd.openDashboard(dashboardURL)

	followCtx := context.Background()
	cmd.Finalizers.Add(dockerWrapper.Stop(followCtx, id, func(i ...interface{}) { fmt.Print(i...) }))
	return dockerWrapper.ContainerFollowRun(followCtx, id)
}

func (cmd *command) initContainerRunOpts(envs []string) (docker.ContainerRunOpts, string) {
	containerRunOpts := docker.ContainerRunOpts{
		Envs:          envs,
		ContainerName: cmd.opts.ContainerName,
		Image:         "eu.gcr.io/kyma-project/busola:latest",
		Ports: map[string]string{
			"3001": cmd.opts.Port,
		},
	}

	containerKubeconfigFilename := "config.yaml" // this is the kubeconfig filename in the container
	kubeconfigPath := kube.KubeconfigPath(cmd.KubeconfigPath)
	dashboardURL := fmt.Sprintf("http://localhost:%s", cmd.opts.Port)
	if _, err := kube.NewFromConfig("", kubeconfigPath); err != nil {
		cmd.CurrentStep.LogInfof("WARNING: Could not identify the current kubeconfig. "+
			"Thus, the Dashboard will start without the current cluster context. More details: %v", err)
	} else {
		if cmd.Verbose {
			fmt.Printf("Mounting kubeconfig '%s' into container ...\n", kubeconfigPath)
		}
		containerRunOpts.Mounts = []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: kubeconfigPath,
				Target: fmt.Sprintf("/app/core/kubeconfig/%s", containerKubeconfigFilename),
			},
		}
		dashboardURL = fmt.Sprintf("%s?kubeconfigID=%s", dashboardURL, containerKubeconfigFilename)
	}

	if isLinuxOperatingSystem() {
		if cmd.Verbose {
			fmt.Printf("Operating system seems to be Linux. Changing the Docker network mode to 'host'")
		}
		containerRunOpts.NetworkMode = "host"
	}

	return containerRunOpts, dashboardURL
}

func (cmd *command) openDashboard(url string) {
	step := cmd.NewStep("Opening the Kyma dashboard in the default browser using the following url: " + url)

	err := browser.OpenURL(url)
	if err != nil {
		step.Failuref("Failed to open the Kyma dashboard. Try to open the url manually")
		if cmd.opts.Verbose {
			step.LogErrorf("error: %v\n", err)
		}
		return
	}
	step.Success()
}

func isLinuxOperatingSystem() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	// check if the operating system is Windows Subsystem for Linux
	if kernelRelease, err := cli.RunCmd("uname", "-r"); err == nil {
		if strings.Contains(kernelRelease, "microsoft") {
			return false
		}
	}

	return true
}
