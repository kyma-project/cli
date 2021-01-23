package function

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io"
	"math/rand"
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
	if err := c.opts.setDefaults(); err != nil {
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

	if c.opts.ImageName == "" {
		name := cfg.Name
		tag := fmt.Sprint(rand.Int())
		c.opts.ImageName = fmt.Sprintf("%s:%s", name, tag)
	}

	s := c.NewStep(fmt.Sprintf("Build project %s", c.opts.ImageName))
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
		Dockerfile: "Dockerfile",
		Tags:       []string{c.opts.ImageName},
	}

	if err := BuildImage(client, ctx, buildOpts); err != nil {
		return err
	}

	s.Success()

	s = c.NewStep("Run container")

	runOpts := &RunOptions{
		ImgName: c.opts.ImageName,
	}

	id, err := RunContainer(client, ctx, runOpts)
	if err != nil {
		s.Failure()
		return err
	}

	s.Success()

	return FollowContainer(client, ctx, id)
}

type BuildOptions struct {
	ImgCtx     string
	Dockerfile string
	Tags       []string
}

const dockerfile = `https://raw.githubusercontent.com/kyma-project/kyma/master/components/function-runtimes/nodejs12/Dockerfile`

func BuildImage(c *client.Client, ctx context.Context, opts *BuildOptions) error {
	reader, err := archive.TarWithOptions(opts.ImgCtx, &archive.TarOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	response, err := c.ImageBuild(ctx, reader, types.ImageBuildOptions{
		//RemoteContext:     dockerfile,
		Tags:           opts.Tags,
		SuppressOutput: true,
	})
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return log(response.Body)
}

type ErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ResultEntry struct {
	Stream      string      `json:"stream,omitempty"`
	ErrorDetail ErrorDetail `json:"errorDetail,omitempty"`
	Error       string      `json:"error,omitempty"`
}

func log(readerCloser io.ReadCloser) error {
	buf := bufio.NewReader(readerCloser)
	for {
		line, err := buf.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		var entryResult ResultEntry
		if err = json.Unmarshal(line, &entryResult); err != nil {
			return err
		}
		if entryResult.Error != "" {
			return errors.Wrap(errors.New("image build failed"), entryResult.Error)
		}
		fmt.Print(entryResult.Stream)
	}
	return nil
}

type RunOptions struct {
	ImgName string
}

func RunContainer(c *client.Client, ctx context.Context, opts *RunOptions) (string, error) {
	body, err := c.ContainerCreate(ctx, &container.Config{
		ExposedPorts: nat.PortSet{
			"8080/tcp": struct{}{},
		},
		Image: opts.ImgName,
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			"8080/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "8080",
				},
			},
		},
	}, nil, "")
	if err != nil {
		return "", err
	}

	err = c.ContainerStart(ctx, body.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", err
	}

	return body.ID, nil
}

func FollowContainer(c *client.Client, ctx context.Context, ID string) error {
	//defer func(ID string) {
	//	err := c.ContainerStop(ctx, ID, nil)
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//}(ID)
	resp, err := c.ContainerLogs(ctx, ID, types.ContainerLogsOptions{
		ShowStdout: false,
		ShowStderr: false,
		Follow:     true,
	})
	if err != nil {
		return err
	}
	defer resp.Close()

	for {
		var b []byte
		_, err := resp.Read(b)
		if err == io.EOF {
			fmt.Sprintf("Łooo chuj...")
			break
		}
		if err != nil {
			return err
		}

		fmt.Println(string(b))
	}

	return nil
}
