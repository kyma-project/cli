package dashboard

import (
	"context"
	"fmt"

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

//NewCmd creates a new dashboard command
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
	cmd.Flags().BoolVarP(&o.Detach, "detach", "d", false, `Change this flag to "true" if you don't want to follow the logs of the local container.`)

	return cmd
}

//Run runs the command
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

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "failed to initialize the Kubernetes client from given kubeconfig")
	}

	kubeconfigPath := kube.KubeconfigPath(cmd.KubeconfigPath)
	kubeconfigFilename := "config.yaml" // this is the kubeconfig filename in the container

	localDashboardURL := fmt.Sprintf("http://localhost:%s?kubeconfigID=%s", cmd.opts.Port, kubeconfigFilename)
	return cmd.runDashboardContainer(localDashboardURL, kubeconfigPath, kubeconfigFilename)
}

func (cmd *command) runDashboardContainer(dashboardURL, kubeconfigPath, kubeconfigFilename string) error {
	step := cmd.NewStep(fmt.Sprintf("Starting container: %s", cmd.opts.ContainerName))

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
		fmt.Printf("Mounting kubeconfig '%s' into container ...\n", kubeconfigPath)
	}

	id, err := dockerWrapper.PullImageAndStartContainer(ctx, docker.ContainerRunOpts{
		Envs:          envs,
		ContainerName: cmd.opts.ContainerName,
		Image:         "eu.gcr.io/kyma-project/busola:latest",
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: kubeconfigPath,
				Target: fmt.Sprintf("/app/core/kubeconfig/%s", kubeconfigFilename),
			},
		},
		Ports: map[string]string{
			"3001": cmd.opts.Port,
		},
	})

	if err != nil {
		step.Failure()
		return errors.Wrap(err, "unable to start container")
	}
	step.Successf("Started container: %s", cmd.opts.ContainerName)

	cmd.openDashboard(dashboardURL)

	if !cmd.opts.Detach {
		step.LogInfo("Logs from the container:")
		followCtx := context.Background()
		cmd.Finalizers.Add(dockerWrapper.Stop(followCtx, id, func(i ...interface{}) { fmt.Print(i...) }))
		return dockerWrapper.ContainerFollowRun(followCtx, id)
	}
	return nil
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
