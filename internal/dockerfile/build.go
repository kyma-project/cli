package dockerfile

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/cli/cli/command/image/build"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/progress"
	"github.com/docker/docker/pkg/streamformatter"
	"github.com/moby/term"
	"github.com/pkg/errors"
)

type BuildOptions struct {
	ImageName      string
	BuildContext   string
	DockerfilePath string
}

func Build(ctx context.Context, opts *BuildOptions) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	excludes, err := build.ReadDockerignore(opts.BuildContext)
	if err != nil {
		return err
	}

	err = build.ValidateContextDirectory(opts.BuildContext, excludes)
	if err != nil {
		return errors.Wrap(err, "error checking context")
	}

	buildCtx, err := archive.TarWithOptions(opts.BuildContext, &archive.TarOptions{
		ExcludePatterns: excludes,
		ChownOpts:       &idtools.Identity{UID: 0, GID: 0},
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

	progressOutput := streamformatter.NewProgressOutput(os.Stdout)
	bodyProgressReader := progress.NewProgressReader(buildCtx, progressOutput, 0, "", "Sending build context to Docker daemon")

	response, err := cli.ImageBuild(
		ctx,
		bodyProgressReader,
		types.ImageBuildOptions{
			Context:    buildCtx,
			Dockerfile: dockerFile,
			Tags:       []string{opts.ImageName},
			Platform:   "linux/amd64",
		},
	)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	fd, isTerm := term.GetFdInfo(os.Stdout)

	err = jsonmessage.DisplayJSONMessagesStream(response.Body, os.Stdout, fd, isTerm, nil)
	if err != nil {
		if jerr, ok := err.(*jsonmessage.JSONError); ok {
			// If no error code is set, default to 1
			if jerr.Code == 0 {
				jerr.Code = 1
			}
			return fmt.Errorf("failed to build image: %d - %s", jerr.Code, jerr.Message)
		}
		return err
	}

	return nil
}
