package dockerfile

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/docker"
)

func Build(ctx context.Context, opts docker.BuildOptions) error {
	cli, err := docker.NewClient()
	if err != nil {
		return err
	}

	return cli.Build(ctx, opts)
}
