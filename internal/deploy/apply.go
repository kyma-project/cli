package deploy

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-project/cli/internal/kube"
)

const (
	defaultRetries        = 3
	defaultInitialBackoff = 3 * time.Second
)

type applyOpts struct {
	dryRun         bool
	force          bool
	retries        uint
	initialBackoff time.Duration
}

func applyManifests(ctx context.Context, k8s kube.KymaKube, manifests []byte, opts applyOpts) error {
	// apply manifests with incremental retry
	if opts.dryRun {
		fmt.Println(string(manifests))
	} else {
		objs, err := k8s.ParseManifest(manifests)
		if err != nil {
			return err
		}

		if err := retry.Do(
			func() error {
				return k8s.Apply(context.Background(), opts.force, objs...)
			}, retry.Attempts(defaultRetries), retry.Delay(defaultInitialBackoff), retry.DelayType(retry.BackOffDelay),
			retry.LastErrorOnly(false), retry.Context(ctx),
		); err != nil {
			return err
		}

		if err := checkDeploymentReadiness(objs, k8s); err != nil {
			return err
		}
	}
	return nil
}
