package deploy

import (
	"context"
	"fmt"
	"os"

	"github.com/avast/retry-go"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/lifecycle-manager/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/resource"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// Kyma deploys the Kyma CR. If no kymaCRPath is provided, it deploys the default CR.
func Kyma(
	ctx context.Context, k8s kube.KymaKube, namespace, channel, kymaCRpath string, dryRun bool,
) error {
	nsObj := &v1.Namespace{}
	nsObj.SetName(namespace)

	kyma := &v1beta1.Kyma{}
	kyma.SetName("default-kyma")
	kyma.SetNamespace(namespace)
	kyma.SetAnnotations(map[string]string{"cli.kyma-project.io/source": "deploy"})
	kyma.SetLabels(map[string]string{"operator.kyma-project.io/managed-by": "lifecycle-manager"})
	kyma.Spec.Channel = channel
	kyma.Spec.Sync.Enabled = false
	kyma.Spec.Modules = []v1beta1.Module{}

	if kymaCRpath != "" {
		data, err := os.ReadFile(kymaCRpath)
		if err != nil {
			return fmt.Errorf("could not read kyma CR file: %w", err)
		}
		if err := yaml.Unmarshal(data, kyma); err != nil {
			return fmt.Errorf("kyma cr file is not valid: %w", err)
		}
	}

	if dryRun {
		result, err := yaml.Marshal(nsObj)
		if err != nil {
			return err
		}
		fmt.Printf("%s---\n", result)
		kyma, err := yaml.Marshal(kyma)
		if err != nil {
			return err
		}
		fmt.Printf("%s---\n", kyma)
		return nil
	}
	if err := retry.Do(
		func() error {
			return k8s.Apply(
				context.Background(), []*resource.Info{
					{
						Name:   nsObj.GetName(),
						Object: nsObj,
					},
					{
						Name:      kyma.GetName(),
						Namespace: kyma.GetNamespace(),
						Object:    kyma,
					},
				},
			)
		}, retry.Attempts(defaultRetries), retry.Delay(defaultInitialBackoff), retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(false), retry.Context(ctx),
	); err != nil {
		return err
	}

	if err := k8s.WatchObject(
		ctx,
		kyma,
		func(kyma ctrlClient.Object) (bool, error) {
			return string(kyma.(*v1beta1.Kyma).Status.State) == string(v1beta1.StateReady), nil
		},
	); err != nil {
		return fmt.Errorf("kyma custom resource did not get ready: %w", err)
	}

	return nil
}
