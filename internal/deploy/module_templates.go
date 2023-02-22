package deploy

import (
	"context"
	"fmt"
	"os"

	"github.com/avast/retry-go"
	"github.com/kyma-project/cli/internal/kube"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

func ModuleTemplates(ctx context.Context, k8s kube.KymaKube, templates []string, force, dryRun bool) error {
	totalObjs := []ctrl.Object{}

	for _, t := range templates {
		b, err := os.ReadFile(t)
		if err != nil {
			return err
		}

		if dryRun {
			fmt.Printf("%s---\n", b)
			continue
		}

		objs, err := k8s.ParseManifest(b)
		if err != nil {
			return err
		}
		totalObjs = append(totalObjs, objs...)

	}

	if err := retry.Do(
		func() error {
			return k8s.Apply(context.Background(), force, totalObjs...)
		}, retry.Attempts(defaultRetries), retry.Delay(defaultInitialBackoff), retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(false), retry.Context(ctx),
	); err != nil {
		return err
	}

	return nil
}
