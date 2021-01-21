package function

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"os"
	"path"
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
		Short: "",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Run()
		},
	}

	cmd.Flags().StringVarP(&o.Filename, "filename", "f", "", `Full path to the config file.`)
	cmd.Flags().StringVarP(&o.ImageName, "name", "n", "", `Full name with tag of the container.`)
	cmd.Flags().DurationVar(&o.BuildTimeout, "timeout", 0, `Maximum time during which the local resources are being built, where "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
	cmd.Flags().BoolVar(&o.BuildOnly, "buildonly", false, `Change this flag to true if you want build container only.`)
	// interactive mode?
	// port publish list?
	// which Dockerfile? maybe custom Dockerfile should be allowed? Or maybe no xD

	return cmd
}

func (c *command) Run() error {
	s := c.NewStep("Build project")
	if err := c.opts.setDefaults(); err != nil {
		s.Failure()
		return err
	}

	file, err := os.Open(c.opts.Filename)
	if err != nil {
		return err
	}

	var cfg workspace.Cfg
	if err := yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return errors.Wrap(err, "Could not decode the configuration file")
	}

	ctx, cancel := context.WithCancel(context.Background())
	if c.opts.BuildTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, c.opts.BuildTimeout)
	}
	defer cancel()

	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		s.Failure()
		return err
	}

	buildOpts := &BuildOptions{
		ImgCtx:     path.Dir(c.opts.Filename),
		Dockerfile: fmt.Sprintf("%s/Dockerfile", path.Dir(c.opts.Filename)),
		Tags:       []string{c.opts.ImageName},
	}
	BuildImage(client, ctx, buildOpts)

	s.Success()

	s = c.NewStep("Run container")

	runOpts := &RunOptions{
		ImgName: c.opts.ImageName,
	}
	if err := RunContainer(client, ctx, runOpts); err != nil {
		s.Failure()
		return err
	}

	s.Success()
	return nil
}

type BuildOptions struct {
	ImgCtx     string
	Dockerfile string
	Tags       []string
}

func BuildImage(c *client.Client, ctx context.Context, opts *BuildOptions) error {
	reader, err := archive.TarWithOptions(opts.ImgCtx, &archive.TarOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	response, err := c.ImageBuild(ctx, reader, types.ImageBuildOptions{
		Dockerfile: opts.Dockerfile,
		Tags:       opts.Tags,
	})
	if err != nil {
		return err
	}
	defer response.Body.Close()

	//why "return log(response.Body)"?
	return nil
}

type RunOptions struct {
	ImgName string
}

func RunContainer(c *client.Client, ctx context.Context, opts *RunOptions) error {
	return c.ContainerStart(ctx, opts.ImgName, types.ContainerStartOptions{})
}
