package docker

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/cli/cli/command/image/build"
	dockerbuild "github.com/docker/docker/api/types/build"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/progress"
	"github.com/docker/docker/pkg/streamformatter"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/moby/go-archive"
	"github.com/moby/term"
	"github.com/pkg/errors"
)

type Client struct {
	client client.APIClient
}

type BuildOptions struct {
	ImageName      string
	BuildContext   string
	DockerfilePath string
	Args           map[string]*string
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Client{client: cli}, nil
}

func NewTestClient(mock client.APIClient) *Client {
	return &Client{client: mock}
}

// Build validates the build context, creates a tar archive of it, builds the image,
// and sets up progress reporting to the standard output.
func (c *Client) Build(ctx context.Context, opts BuildOptions) error {
	excludes, err := build.ReadDockerignore(opts.BuildContext)
	if err != nil {
		return err
	}

	if err := build.ValidateContextDirectory(opts.BuildContext, excludes); err != nil {
		return errors.Wrap(err, "error validating docker context")
	}

	buildCtx, err := archive.TarWithOptions(opts.BuildContext, &archive.TarOptions{
		ExcludePatterns: excludes,
		ChownOpts:       &archive.ChownOpts{UID: 0, GID: 0},
	})
	if err != nil {
		return err
	}
	defer buildCtx.Close()

	dockerFileReader, err := os.Open(opts.DockerfilePath)
	if err != nil {
		return err
	}

	buildCtx, dockerFile, err := build.AddDockerfileToBuildContext(dockerFileReader, buildCtx)
	if err != nil {
		return err
	}

	progressOutput := streamformatter.NewProgressOutput(out.Default.MsgWriter())
	bodyProgressReader := progress.NewProgressReader(buildCtx, progressOutput, 0, "", "Sending build context to Docker daemon")

	response, err := c.client.ImageBuild(
		ctx,
		bodyProgressReader,
		dockerbuild.ImageBuildOptions{
			Context:    buildCtx,
			Dockerfile: dockerFile,
			Tags:       []string{opts.ImageName},
			Platform:   "linux/amd64",
			BuildArgs:  opts.Args,
		},
	)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	fd, isTerm := term.GetFdInfo(out.Default.MsgWriter())

	if err := jsonmessage.DisplayJSONMessagesStream(response.Body, out.Default.MsgWriter(), fd, isTerm, nil); err != nil {
		if jerr, ok := err.(*jsonmessage.JSONError); ok {
			if jerr.Code == 0 {
				jerr.Code = 1
			}
			return fmt.Errorf("failed to build image: %d - %s", jerr.Code, jerr.Message)
		}
		return err
	}

	return nil
}
