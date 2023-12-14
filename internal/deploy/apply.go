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
	defaultRetries        = 10
	defaultInitialBackoff = 10 * time.Second
)

type applyOpts struct {
	dryRun         bool
	force          bool
	retries        uint
	initialBackoff time.Duration
}

func parseManifests(k8s kube.KymaKube, manifests []byte, dryRun bool) ([]client.Object, error) {
	var manifestObjs []client.Object
	var err error
	if !dryRun {
		if manifestObjs, err = k8s.ParseManifest(manifests); err != nil {
			return nil, err
		}
	}
	return manifestObjs, nil
}

func applyManifests(ctx context.Context, k8s kube.KymaKube, manifests []byte, opts applyOpts,
	manifestObjs []client.Object) error {
	if opts.dryRun {
		fmt.Println(string(manifests))
	} else {
		if err := retry.Do(
			func() error {
				return k8s.Apply(context.Background(), opts.force, manifestObjs...)
			}, retry.Attempts(defaultRetries), retry.Delay(defaultInitialBackoff), retry.DelayType(retry.BackOffDelay),
			retry.LastErrorOnly(false), retry.Context(ctx),
		); err != nil {
			return err
		}

		if err := checkDeploymentReadiness(manifestObjs, k8s); err != nil {
			return err
		}
	}
	return nil
}
