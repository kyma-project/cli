package module

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/cli/pkg/errs"

	"github.com/avast/retry-go"
	"github.com/kyma-project/lifecycle-manager/api/shared"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/cli/internal/kube"
)

type Interactor interface {
	// Get retrieves all modules and the channel of the kyma CR that the Interactor deals with.
	Get(ctx context.Context) ([]v1beta2.Module, string, error)
	// Update applies all modules that the Interactor deals with.
	Update(ctx context.Context, modules []v1beta2.Module) error
	// WaitUntilReady blocks until all Modules are confirmed to be applied and ready
	WaitUntilReady(ctx context.Context) error
	GetFilteredModuleTemplates(ctx context.Context, moduleIdentifier string) ([]v1beta2.ModuleTemplate, error)
}

var _ Interactor = &DefaultInteractor{}

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

func (i *DefaultInteractor) Get(ctx context.Context) ([]v1beta2.Module, string, error) {
	kyma := &v1beta2.Kyma{}
	if err := i.K8s.Ctrl().Get(ctx, i.Key, kyma); err != nil {
		return nil, "", fmt.Errorf("could not get Kyma %ss: %w", i.Key, err)
	}

	i.lastKnownResourceVersion = kyma.GetResourceVersion()

	return kyma.Spec.Modules, kyma.Spec.Channel, nil
}

func (i *DefaultInteractor) GetFilteredModuleTemplates(ctx context.Context,
	moduleIdentifier string) ([]v1beta2.ModuleTemplate, error) {
	var allTemplates v1beta2.ModuleTemplateList
	if err := i.K8s.Ctrl().List(ctx, &allTemplates); err != nil {
		return []v1beta2.ModuleTemplate{}, fmt.Errorf("could not get Moduletemplates: %w", err)
	}

	filteredModuleTemplates, err := i.filterModuleTemplates(allTemplates, moduleIdentifier)
	if err != nil {
		return nil, fmt.Errorf("could not filter fetched Moduletemplates: %w", err)
	}
	return filteredModuleTemplates, nil
}

func (i *DefaultInteractor) filterModuleTemplates(allTemplates v1beta2.ModuleTemplateList,
	moduleIdentifier string) ([]v1beta2.ModuleTemplate, error) {
	var filteredModuleTemplates []v1beta2.ModuleTemplate

	for _, mt := range allTemplates.Items {
		if mt.Labels[v1beta2.ModuleName] == moduleIdentifier {
			filteredModuleTemplates = append(filteredModuleTemplates, mt)
			continue
		}
		if mt.ObjectMeta.Name == moduleIdentifier {
			filteredModuleTemplates = append(filteredModuleTemplates, mt)
			continue
		}
		descriptor, err := mt.GetDescriptor()
		if err != nil {
			return nil, fmt.Errorf("invalid ModuleTemplate descriptor: %v", err)
		}
		if descriptor.Name == moduleIdentifier {
			filteredModuleTemplates = append(filteredModuleTemplates, mt)
			continue
		}
	}
	return filteredModuleTemplates, nil
}

// Update tries to update the modules in the Kyma Instance and retries on failure
// It exits without retrying if the Kyma Resource cannot be fetched at least once.
func (i *DefaultInteractor) Update(ctx context.Context, modules []v1beta2.Module) error {
	ctx, cancel := context.WithTimeout(ctx, i.Timeout)
	defer cancel()

	kyma := &v1beta2.Kyma{}
	if err := i.K8s.Ctrl().Get(ctx, i.Key, kyma); err != nil {
		return err
	}
	oldGen := kyma.GetGeneration()
	err := retry.Do(
		func() error {
			kyma.Spec.Modules = modules
			if err := i.K8s.Ctrl().Update(ctx, kyma, &client.UpdateOptions{FieldManager: "kyma"}); err != nil {
				return err
			}
			newGen := kyma.GetGeneration()
			i.changed = oldGen != newGen
			i.lastKnownResourceVersion = kyma.GetResourceVersion()
			return nil
		}, retry.Attempts(3), retry.Delay(3*time.Second), retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(false), retry.Context(ctx),
	)
	return err
}

// WaitUntilReady uses the internal i.changed tracker to determine whether the last apply caused
// any changes on the cluster. If it did not, then it will shortcut to retrieve the latest version
// from the cluster and determine if it is ready. If it has been changed, then
// it will start a watch request and read out the last state. If it is in error,
// it will attempt it again until ctx times out.
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

	kymas := v1beta2.KymaList{}
	if !i.changed {
		if err := i.K8s.Ctrl().List(ctx, &kymas, &options); err != nil {
			return fmt.Errorf("could not start listing for kyma readiness: %w", err)
		}
		return IsKymaReady(i.Logger, &kymas.Items[0])
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
			if res.Object == nil || res.Type != watch.Modified {
				continue
			}
			if objMeta, err := meta.Accessor(res.Object); err != nil ||
				(i.changed && i.lastKnownResourceVersion == objMeta.GetResourceVersion()) {
				i.Logger.Info("changed generation but observed resource version is still the same")
				continue
			}
			if err := IsKymaReady(i.Logger, res.Object); err != nil {
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

// IsKymaReady interprets the status of a Kyma Resource and uses this to determine if it can be considered Ready.
// It checks for shared.StateReady, and if it is set, determines if this state can be trusted by observing
// if the status fields match the desired state, and if the lastOperation is filled by the lifecycle-manager.
func IsKymaReady(l *zap.SugaredLogger, obj runtime.Object) error {
	kyma, ok := obj.(*v1beta2.Kyma)
	if !ok {
		return errs.ErrTypeAssertKyma
	}
	l.Info(kyma.Status)
	switch kyma.Status.State {
	case shared.StateReady:
		if len(kyma.Status.Modules) != len(kyma.Spec.Modules) {
			return fmt.Errorf("kyma has status Ready but cannot be up to date "+
				"since modules tracked in status differ from modules in desired state/spec (%v in status, %v in spec)",
				len(kyma.Status.Modules), len(kyma.Spec.Modules))
		}

		lastOperation := kyma.Status.LastOperation
		if lastOperation.Operation == "" {
			return fmt.Errorf(
				"kyma has status Ready but cannot be up to date "+
					"since last operation was never set: %v", kyma.Status,
			)
		}
		return nil
	case shared.StateWarning:
		return ErrKymaInWarningState
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
