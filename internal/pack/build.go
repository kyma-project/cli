package pack

import (
	"context"
	"fmt"
	"os"

	"github.com/buildpacks/pack/pkg/client"
	"github.com/buildpacks/pack/pkg/logging"
	"github.com/pkg/errors"
)

func Build(ctx context.Context, appName, appPath string) error {
	pack, err := client.NewClient(client.WithLogger(logging.NewLogWithWriters(os.Stdout, os.Stderr)))
	if err != nil {
		return errors.Wrap(err, "failed to create buildpack client")
	}

	err = pack.Build(ctx, client.BuildOptions{
		Image:    appName,
		AppPath:  appPath,
		Platform: "linux/amd64",
		Builder:  "paketobuildpacks/builder-jammy-base",
	})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to build %s app from the %s dir", appName, appPath))
	}

	return nil
}
