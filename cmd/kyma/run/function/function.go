package function

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/client"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/hydroform/function/pkg/docker"
	"github.com/kyma-project/hydroform/function/pkg/docker/runtimes"
	"github.com/kyma-project/hydroform/function/pkg/workspace"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type command struct {
	opts *Options
	cli.Command
}

// NewCmd creates a new init command
func NewCmd(o *Options) *cobra.Command {
	c := command{
		opts:    o,
		Command: cli.Command{Options: o.Options},
	}
	cmd := &cobra.Command{
		Use:   "function",
		Short: "Runs Functions locally.",
		Long:  `Use this command to run a Function in Docker from local sources.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Run()
		},
	}

	cmd.Flags().StringVarP(&o.Filename, "filename", "f", "", `Full path to the config file.`)
	cmd.Flags().StringVarP(&o.Dir, "source-dir", "d", "", `Full path to the folder with the source code.`)
	cmd.Flags().StringVar(&o.ContainerName, "container-name", "", `The name of the created container.`)
	cmd.Flags().BoolVar(&o.Detach, "detach", false, `Change this flag to "true" if you don't want to follow the container logs after running the Function.`)
	cmd.Flags().StringVarP(&o.FuncPort, "port", "p", "8080", `The port on which the container will be exposed.`)
	cmd.Flags().BoolVar(&o.HotDeploy, "hot-deploy", false, `Change this flag to "true" if you want to start a Function in Hot Deploy mode.`)
	cmd.Flags().BoolVar(&o.Debug, "debug", false, `Change this flag to "true" if you want to expose port 9229 for remote debugging.`)

	return cmd
}

func (c *command) Run() error {
	if err := c.opts.defaultFilename(); err != nil {
		return err
	}

	cfg, err := c.workspaceConfig(c.opts.Filename)
	if err != nil {
		return err
	}

	if err = c.opts.defaultValues(cfg); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return errors.Wrap(err, "white trying to interact with docker")
	}

	return c.runContainer(ctx, client, cfg)
}

func (c *command) runContainer(ctx context.Context, client *client.Client, cfg workspace.Cfg) error {
	step := c.NewStep(fmt.Sprintf("Running container: %s", c.opts.ContainerName))
	ports := map[string]string{
		runtimes.ServerPort: c.opts.FuncPort,
	}

	if c.opts.Debug {
		debugPort := runtimes.RuntimeDebugPort(cfg.Runtime)
		ports[debugPort] = debugPort
	}

	id, err := docker.RunContainer(ctx, client, docker.RunOpts{
		Ports: ports,
		Envs: append(
			runtimes.ContainerEnvs(cfg.Runtime, c.opts.HotDeploy),
			c.parseEnvs(cfg.Env)...,
		),
		ContainerName: c.opts.ContainerName,
		Image:         containerImage(cfg),
		Commands:      c.containerCommands(cfg),
		User:          runtimes.ContainerUser,
		Mounts:        runtimes.GetMounts(cfg.Runtime, cfg.Source.Type, c.opts.Dir),
	})
	if err != nil {
		step.Failure()
		return errors.Wrap(err, "while trying to run container")
	}

	step.Successf("Ran container: %s", c.opts.ContainerName)
	step.LogInfo("Container listening on port: " + runtimes.ServerPort)
	if !c.opts.Detach {
		fmt.Println("Logs from the container:")
		followCtx := context.Background()
		c.Finalizers.Add(docker.Stop(followCtx, client, id, func(i ...interface{}) { fmt.Print(i...) }))
		return docker.FollowRun(followCtx, client, id)
	}
	return nil
}

func (c *command) workspaceConfig(path string) (workspace.Cfg, error) {
	file, err := os.Open(path)
	if err != nil {
		return workspace.Cfg{}, err
	}

	var cfg workspace.Cfg
	if err := yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return workspace.Cfg{}, errors.Wrap(err, "while trying to decode the configuration file")
	}

	supportedRuntimes := map[string]struct{}{
		"nodejs18":  {},
		"python39":  {},
		"python312": {},
	}
	if _, ok := supportedRuntimes[cfg.Runtime]; !ok {
		return workspace.Cfg{}, fmt.Errorf("unsupported runtime: %s", cfg.Runtime)
	}
	sourceFile, depsfile, supported := workspace.InlineFileNames(cfg.Runtime)
	if !supported {
		return workspace.Cfg{}, fmt.Errorf("unsupported runtime: %s", cfg.Runtime)
	}
	if cfg.Source.SourceHandlerName == "" {
		cfg.Source.SourceHandlerName = sourceFile
	}
	if cfg.Source.DepsHandlerName == "" {
		cfg.Source.DepsHandlerName = depsfile
	}
	return cfg, nil
}

func (c *command) containerCommands(cfg workspace.Cfg) []string {
	var cmd []string
	if cfg.Source.Type == workspace.SourceTypeInline {
		cmd = runtimes.MoveInlineCommand(cfg.Runtime, cfg.Source.SourceHandlerName, cfg.Source.DepsHandlerName)
	}

	return append(cmd, runtimes.ContainerCommands(cfg.Runtime, c.opts.Debug, c.opts.HotDeploy)...)
}

func (c *command) parseEnvs(envVars []workspace.EnvVar) []string {
	var envs []string
	for _, env := range envVars {
		envs = append(envs, fmt.Sprintf("%s=%s", env.Name, env.Value))
	}
	return envs
}

func containerImage(cfg workspace.Cfg) string {
	if cfg.RuntimeImageOverride != "" {
		return cfg.RuntimeImageOverride
	}
	return runtimes.ContainerImage(cfg.Runtime)
}
