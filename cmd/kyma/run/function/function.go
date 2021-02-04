package function

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/hydroform/function/pkg/docker/runtimes"
	"os"

	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/kyma-incubator/hydroform/function/pkg/docker"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"gopkg.in/yaml.v2"
)

type command struct {
	opts *Options
	cli.Command
}

//NewCmd creates a new init command
func NewCmd(o *Options) *cobra.Command {
	c := command{
		opts:    o,
		Command: cli.Command{Options: o.Options},
	}
	cmd := &cobra.Command{
		Use:   "function",
		Short: "Run functions locally.",
		Long:  `Use this command to run function in docker from local sources.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Run()
		},
	}

	cmd.Flags().StringVarP(&o.Filename, "filename", "f", "", `Full path to the config file.`)
	cmd.Flags().StringVar(&o.ImageName, "imageName", "", `Full name with tag of the container.`)
	cmd.Flags().StringVar(&o.ContainerName, "containerName", "", `The name of the created container.`)
	cmd.Flags().DurationVarP(&o.Timeout, "timeout", "t", 0, `Maximum time during which the local resources are being built, where "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
	cmd.Flags().BoolVarP(&o.Detach, "detach", "d", false, `Change this flag to true if you don't want to follow the container logs after run'.`)
	cmd.Flags().StringVarP(&o.FuncPort, "port", "p", "8080", `The port on which the container will be exposed.`)
	cmd.Flags().StringArrayVarP(&o.Envs, "env", "e", []string{}, `The system environments witch which the container will be run.`)

	//cmd.Flags().BoolVarP(&o.Debug, "debug", "d", false, `Change this flat to true if you want to expose port 9229 for remote debugging.`)

	return cmd
}

func (c *command) Run() error {
	if c.opts.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if err := c.opts.setDefaults(); err != nil {
		return err
	}

	cfg, err := workspaceConfig(c.opts.Filename)
	if err != nil {
		return err
	}

	// git functions are not supported yer
	if cfg.Source.Type == workspace.SourceTypeGit {
		return errors.New("The git source type of functions is not supported yet")
	}

	if c.opts.ImageName == "" {
		name := cfg.Name
		tag := uuid.New()
		c.opts.ImageName = fmt.Sprintf("%s:%s", name, tag)
	}

	if c.opts.ContainerName == "" {
		c.opts.ContainerName = cfg.Name
	}

	ctx, cancel := context.WithCancel(context.Background())
	if c.opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, c.opts.Timeout)
	}
	defer cancel()

	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return errors.Wrap(err, "white trying to interact with docker")
	}

	err = c.build(ctx, client, cfg)
	if err != nil {
		return err
	}

	return c.runContainer(ctx, client, cfg)
}

func workspaceConfig(path string) (workspace.Cfg, error) {
	file, err := os.Open(path)
	if err != nil {
		return workspace.Cfg{}, err
	}
	var cfg workspace.Cfg
	if err := yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return workspace.Cfg{}, errors.Wrap(err, "while trying to decode the configuration file")
	}

	if cfg.Runtime != "nodejs12" &&
		cfg.Runtime != "nodejs10" &&
		cfg.Runtime != "python38" {
		return workspace.Cfg{}, fmt.Errorf("unsupported runtime: %s", cfg.Runtime)
	}

	return cfg, nil
}

func (c *command) build(ctx context.Context, client *client.Client, cfg workspace.Cfg) error {
	context, err := docker.InlineContext(docker.ContextOpts{
		DirPrefix:  c.opts.ContainerName,
		Dockerfile: runtimes.Dockerfile(cfg.Runtime),
		SrcDir:     cfg.Source.SourcePath,
		SrcFiles:   sourceFiles(cfg),
	}, logrus.Debugf)
	if err != nil {
		return errors.Wrap(err, "white trying to prepare folder for docker build")
	}

	c.newStep(fmt.Sprintf("Building project: %s", c.opts.ImageName))
	resp, err := docker.BuildImage(client, ctx, docker.BuildOpts{
		Context: context,
		Tags:    []string{c.opts.ImageName},
	})
	if err != nil {
		return errors.Wrap(err, "white trying to build image")
	}
	defer resp.Body.Close()

	err = docker.FollowBuild(resp.Body, logrus.Debug)
	if err != nil {
		return errors.Wrap(err, "during the build")
	}

	c.successStepf(fmt.Sprintf("Built image: %s", c.opts.ImageName))
	return nil
}

func (c *command) runContainer(ctx context.Context, client *client.Client, cfg workspace.Cfg) error {
	c.newStep(fmt.Sprintf("Running container: %s", c.opts.ContainerName))
	id, err := docker.RunContainer(client, ctx, docker.RunOpts{
		Ports: runtimes.ContainerPorts(cfg.Runtime, c.opts.FuncPort, c.opts.Debug),
		Envs: append(
			runtimes.ContainerEnvs(cfg.Runtime, c.opts.Debug),
			c.opts.Envs...,
		),
		ContainerName: c.opts.ContainerName,
		ImageName:     c.opts.ImageName,
	})
	if err != nil {
		return errors.Wrap(err, "white trying to run container")
	}

	c.successStepf(fmt.Sprintf("Runned container: %s", c.opts.ContainerName))

	if !c.opts.Detach {
		fmt.Println("Logs from the container:")
		followCtx := context.Background()
		c.Finalizer.Add(docker.Stop(client, followCtx, id, func(i ...interface{}) { fmt.Print(i...) }))
		return docker.FollowRun(client, followCtx, id, func(i ...interface{}) { fmt.Print(i...) })
	}
	return nil
}

func (c *command) newStep(message string) {
	if c.opts.Verbose {
		logrus.Debug(message)
	} else {
		c.NewStep(message)
	}
}

func (c *command) successStepf(message string) {
	if c.opts.Verbose {
		logrus.Debug(message)
	} else if c.CurrentStep != nil {
		c.CurrentStep.Successf(message)
	}
}

func sourceFiles(cfg workspace.Cfg) []string {
	var files []string
	defaultHandler, defaultDeps, _ := workspace.InlineFileNames(cfg.Runtime)

	handlerFilename := cfg.Source.SourceHandlerName
	if handlerFilename == "" {
		handlerFilename = defaultHandler
	}
	files = append(files, handlerFilename)

	depsFilename := cfg.Source.DepsHandlerName
	if depsFilename == "" {
		depsFilename = defaultDeps
	}
	files = append(files, depsFilename)

	return files
}
