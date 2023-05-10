package deploy

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-project/cli/internal/kube"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func applyManifests(ctx context.Context, k8s kube.KymaKube, manifests []byte, opts applyOpts) ([]client.Object, error) {
	// apply manifests with incremental retry
	var manifestObjs []client.Object
	if opts.dryRun {
		fmt.Println(string(manifests))
	} else {
		var err error
		manifestObjs, err = k8s.ParseManifest(manifests)
		if err != nil {
			return nil, err
		}

		if err := retry.Do(
			func() error {
				return k8s.Apply(context.Background(), opts.force, manifestObjs...)
			}, retry.Attempts(defaultRetries), retry.Delay(defaultInitialBackoff), retry.DelayType(retry.BackOffDelay),
			retry.LastErrorOnly(false), retry.Context(ctx),
		); err != nil {
			return nil, err
		}

		if err := checkDeploymentReadiness(manifestObjs, k8s); err != nil {
			return nil, err
		}
	}
	return manifestObjs, nil
}
