package deploy

import (
	"context"
	"fmt"
	"os"

	"github.com/kyma-project/cli/pkg/errs"

	"github.com/avast/retry-go"
	"github.com/kyma-project/lifecycle-manager/api/shared"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	v1 "k8s.io/api/core/v1"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/kyma-project/cli/internal/kube"
)

// Kyma deploys the Kyma CR. If no kymaCRPath is provided, it deploys the default CR.
func Kyma(
	ctx context.Context, k8s kube.KymaKube, namespace, channel, kymaCRpath string, force, dryRun, kcpMode bool,
) error {
	namespaceObj := &v1.Namespace{}
	namespaceObj.SetName(namespace)

	kyma := &v1beta2.Kyma{}
	if kymaCRpath != "" {
		data, err := os.ReadFile(kymaCRpath)
		if err != nil {
			return fmt.Errorf("could not read kyma CR file: %w", err)
		}
		if err := yaml.Unmarshal(data, kyma); err != nil {
			return fmt.Errorf("kyma cr file is not valid: %w", err)
		}
	} else {
		kyma.SetName("default-kyma")
		kyma.SetNamespace(namespace)
		kyma.SetAnnotations(map[string]string{"cli.kyma-project.io/source": "deploy"})
		kyma.SetLabels(map[string]string{v1beta2.ManagedBy: "lifecycle-manager"})
		kyma.Spec.Channel = channel
		if !kcpMode {
			kyma.SetLabels(map[string]string{v1beta2.SyncLabel: v1beta2.DisableLabelValue})
		}
		kyma.Spec.Modules = []v1beta2.Module{}
	}

	if dryRun {
		result, err := yaml.Marshal(namespaceObj)
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
			return k8s.Apply(context.Background(), force, namespaceObj, kyma)
		}, retry.Attempts(defaultRetries), retry.Delay(defaultInitialBackoff), retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(false), retry.Context(ctx),
	); err != nil {
		return err
	}

	if err := k8s.WatchObject(
		ctx, kyma,
		func(obj ctrlClient.Object) (bool, error) {
			tKyma, ok := obj.(*v1beta2.Kyma)
			if !ok {
				return false, errs.ErrTypeAssertKyma
			}
			return string(tKyma.Status.State) == string(shared.StateReady), nil
		},
	); err != nil {
		return fmt.Errorf("kyma custom resource did not get ready: %w", err)
	}

	return nil
}
