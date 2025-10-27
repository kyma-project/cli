package pack

import (
	"context"
	"fmt"
	"os"

	"github.com/buildpacks/pack/pkg/cache"
	"github.com/buildpacks/pack/pkg/client"
	"github.com/buildpacks/pack/pkg/logging"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/pkg/errors"
)

func Build(ctx context.Context, appName, appPath string) error {
	pack, err := client.NewClient(client.WithLogger(logging.NewLogWithWriters(out.Default.MsgWriter(), out.Default.ErrWriter())))
	if err != nil {
		return errors.Wrap(err, "failed to create buildpack client")
	}

	tmpDir := os.TempDir()

	err = pack.Build(ctx, client.BuildOptions{
		Image:    appName,
		AppPath:  appPath,
		Platform: "linux/amd64",
		Builder:  "paketobuildpacks/builder-jammy-base",
		Cache: cache.CacheOpts{
			Build: cache.CacheInfo{
				Format: cache.CacheBind,
				Source: fmt.Sprintf("%s/kyma-cache/app-push/build", tmpDir),
			},
			Launch: cache.CacheInfo{
				Format: cache.CacheBind,
				Source: fmt.Sprintf("%s/kyma-cache/app-push/launch", tmpDir),
			},
			Kaniko: cache.CacheInfo{
				Format: cache.CacheBind,
				Source: fmt.Sprintf("%s/kyma-cache/app-push/kaniko", tmpDir),
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to build %s app from the %s dir", appName, appPath))
	}

	return nil
}
