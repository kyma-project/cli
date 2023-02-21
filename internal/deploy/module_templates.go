package deploy

import (
	"context"
	"os"

	"github.com/avast/retry-go"
	"github.com/kyma-project/cli/internal/kube"
)

func ModuleTemplates(ctx context.Context, k8s kube.KymaKube, templates []string) error {
	for _, t := range templates {
		b, err := os.ReadFile(t)
		if err != nil {
			return err
		}
		if err := retry.Do(
			func() error {
				manifests, err := k8s.ParseManifest(b)
				if err != nil {
					return err
				}
				return k8s.Apply(context.Background(), manifests)
			}, retry.Attempts(defaultRetries), retry.Delay(defaultInitialBackoff), retry.DelayType(retry.BackOffDelay),
			retry.LastErrorOnly(false), retry.Context(ctx),
		); err != nil {
			return err
		}
	}

	return nil
}
