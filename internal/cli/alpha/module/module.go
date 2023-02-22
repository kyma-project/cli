package module

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/lifecycle-manager/api/v1beta1"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Interactor interface {
	// Get retrieves all modules that the Interactor deals with.
	Get(ctx context.Context) ([]v1beta1.Module, error)
	// Apply applies all modules that the Interactor deals with.
	Apply(ctx context.Context, modules []v1beta1.Module) error
	// WaitUntilReady blocks until all Modules are confirmed to be applied and ready
	WaitUntilReady(ctx context.Context, onError func(kyma *v1beta1.Kyma)) error
}

type DefaultInteractor struct {
	Logger                   *zap.SugaredLogger
	K8s                      kube.KymaKube
	Key                      types.NamespacedName
	ForceUpdate              bool
	Timeout                  time.Duration
	lastKnownResourceVersion string
	changed                  bool
}

func NewInteractor(
	logger *zap.SugaredLogger, k8s kube.KymaKube, name types.NamespacedName, timeout time.Duration, forceUpdate bool,
) DefaultInteractor {
	return DefaultInteractor{
		Logger:      logger,
		K8s:         k8s,
		Key:         name,
		ForceUpdate: forceUpdate,
		Timeout:     timeout,
	}
}

func (i *DefaultInteractor) Get(ctx context.Context) ([]v1beta1.Module, error) {
	kyma := &v1beta1.Kyma{}
	if err := i.K8s.Ctrl().Get(ctx, i.Key, kyma); err != nil {
		return nil, fmt.Errorf("could not get Kyma %ss: %w", i.Key, err)
	}

	i.lastKnownResourceVersion = kyma.GetResourceVersion()

	return kyma.Spec.Modules, nil
}

func (i *DefaultInteractor) Apply(ctx context.Context, modules []v1beta1.Module) error {
	ctx, cancel := context.WithTimeout(ctx, i.Timeout)
	defer cancel()

	kyma := &v1beta1.Kyma{}
	if err := i.K8s.Ctrl().Get(ctx, i.Key, kyma); err != nil {
		return err
	}
	oldGen := kyma.GetGeneration()

	kyma.Spec.Modules = modules
	if err := retry.Do(
		func() error {
			return i.K8s.Apply(context.Background(), i.ForceUpdate, kyma)
		}, retry.Attempts(3), retry.Delay(3*time.Second), retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(false), retry.Context(ctx),
	); err != nil {
		return err
	}
	newGen := kyma.GetGeneration()

	i.changed = oldGen != newGen
	i.lastKnownResourceVersion = kyma.GetResourceVersion()

	return nil
}

// WaitUntilReady uses the internal i.changed tracker to determine wether the last apply caused
// any changes on the cluster. If it did not, then it will shortcut to retrieve the latest version
// from the cluster and determine if its ready. If it has been changed, then
// it will start a watch request and read out the last state. If it is in error
// it will attempt it again until ctx times out or max retries is reached.
// If the logger is active (usually based on verbosity), it will also log the error.
// The error returned will always contain all aggregated errors encountered during the watch if the kyma never got ready
// If there is no error available (e.g. because the lifecycle manager never updated the resource version)
// then the context.DeadlineExceeded Error will be thrown instead.
func (i *DefaultInteractor) WaitUntilReady(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, i.Timeout)
	defer cancel()

	options := client.ListOptions{
		FieldSelector: fields.AndSelectors(
			fields.OneTermEqualSelector("metadata.name", i.Key.Name),
			fields.OneTermEqualSelector("metadata.namespace", i.Key.Namespace),
		),
		Raw: &metav1.ListOptions{
			ResourceVersion: i.lastKnownResourceVersion,
		},
	}

	kymas := v1beta1.KymaList{}
	if !i.changed {
		if err := i.K8s.Ctrl().List(ctx, &kymas, &options); err != nil {
			return fmt.Errorf("could not start listing for kyma readiness: %w", err)
		}
		return i.isKymaReady(&kymas.Items[0])
	}

	watcher, err := i.K8s.Ctrl().Watch(ctx, &kymas, &options)
	if err != nil {
		return fmt.Errorf("could not start watching for kyma readiness: %w", err)
	}
	defer watcher.Stop()

	var errs []error
	for {
		select {
		case res := <-watcher.ResultChan():
			if res.Object == nil {
				continue
			}
			if objMeta, err := meta.Accessor(res.Object); err != nil ||
				(i.changed && i.lastKnownResourceVersion == objMeta.GetResourceVersion()) {
				continue
			}
			if err := i.isKymaReady(res.Object); err != nil {
				i.Logger.Errorf("%s", err)
				errs = append(errs, err)
				continue
			}
			return nil
		case <-ctx.Done():
			if len(errs) > 0 {
				return retry.Error(errs)
			}
			return ctx.Err()
		}
	}
}

func (i *DefaultInteractor) isKymaReady(obj runtime.Object) error {
	kyma := obj.(*v1beta1.Kyma)
	switch kyma.Status.State {
	case v1beta1.StateReady:
		lastOperation := kyma.Status.LastOperation
		if lastOperation.Operation == "" {
			return fmt.Errorf(
				"kyma has status Ready but cannot be up to date "+
					"since last operation was never set: %v", kyma.Status,
			)
		}
		return nil
	default:
		lastOperation := kyma.Status.LastOperation
		if lastOperation.Operation == "" {
			return fmt.Errorf(
				"there is no last operation available to show details: %v", kyma.Status,
			)
		}
		return errors.Errorf(
			"%s at %s", lastOperation.Operation, lastOperation.LastUpdateTime.Time,
		)
	}
}
