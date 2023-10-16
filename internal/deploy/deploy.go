package deploy

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/reconciler/pkg/cluster"
	"github.com/kyma-incubator/reconciler/pkg/keb"
	"github.com/kyma-incubator/reconciler/pkg/model"
	"github.com/kyma-incubator/reconciler/pkg/reconciler"
	reconcilerservice "github.com/kyma-incubator/reconciler/pkg/reconciler/service"
	"github.com/kyma-incubator/reconciler/pkg/scheduler/reconciliation"
	"github.com/kyma-incubator/reconciler/pkg/scheduler/service"
	"github.com/kyma-project/cli/internal/deploy/component"
	"github.com/kyma-project/cli/internal/deploy/values"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ComponentState string

const (
	Success            ComponentState = "Success"
	RecoverableError   ComponentState = "RecoverableError"
	UnrecoverableError ComponentState = "UnrecoverableError"
)

type ComponentStatus struct {
	Component string
	State     ComponentState
	Error     error
	Manifest  *string
}

var PrintedStatus = make(map[string]bool)
var manifestsBuffer []ComponentStatus

type Options struct {
	Components     component.List
	Values         values.Values
	StatusFunc     func(status ComponentStatus)
	KubeConfig     []byte
	KymaVersion    string
	KymaProfile    string
	Logger         *zap.SugaredLogger
	WorkerPoolSize int
	DeleteStrategy string
	DryRun         bool
}

func Deploy(opts Options) (*service.ReconciliationResult, error) {
	if opts.DryRun {
		reconcilerservice.EnableReconcilerDryRun()
	}
	return doReconciliation(opts, false)
}

func Undeploy(opts Options) (*service.ReconciliationResult, error) {
	return doReconciliation(opts, true)
}

func doReconciliation(opts Options, delete bool) (*service.ReconciliationResult, error) {
	kebComponents, err := prepareKebComponents(opts.Components, opts.Values)
	if err != nil {
		return nil, err
	}

	kebCluster := prepareKebCluster(opts, kebComponents, delete)

	ds, err := service.NewDeleteStrategy(opts.DeleteStrategy)
	if err != nil {
		return nil, err
	}

	manifests := make(chan ComponentStatus)
	runtimeBuilder := service.NewRuntimeBuilder(reconciliation.NewInMemoryReconciliationRepository(), opts.Logger)
	statusFunc := func(component string, msg *reconciler.CallbackMessage) {
		var status ComponentStatus
		switch msg.Status {
		case reconciler.StatusSuccess:
			status = ComponentStatus{component, Success, nil, msg.Manifest}
		case reconciler.StatusFailed:
			status = ComponentStatus{component,
				RecoverableError,
				errors.Errorf("%s", msg.Error),
				msg.Manifest}
		case reconciler.StatusError:
			status = ComponentStatus{component,
				UnrecoverableError,
				errors.Errorf("%s", msg.Error),
				msg.Manifest}
		}

		if opts.DryRun {
			go manifestCollector(manifests)
			manifests <- status
		}

		opts.StatusFunc(status)
	}

	reconcilationResult, err := runtimeBuilder.RunLocal(statusFunc).
		WithSchedulerConfig(&service.SchedulerConfig{
			PreComponents:  opts.Components.PrerequisiteNames(),
			DeleteStrategy: ds,
		}).
		WithWorkerPoolSize(opts.WorkerPoolSize).
		Run(context.TODO(), kebCluster)

	close(manifests)
	if opts.DryRun {
		printManifests()
	}

	return reconcilationResult, err
}

func manifestCollector(ch chan ComponentStatus) {
	str := <-ch
	manifestsBuffer = append(manifestsBuffer, str)
}

func printManifests() {
	for _, v := range manifestsBuffer {
		if v.Error != nil {
			fmt.Printf("Rendering of Component: %s failed with %s", v.Component, v.Error.Error())
			return
		}
	}
	for _, val := range manifestsBuffer {
		fmt.Printf("%s", *val.Manifest)
	}
}

func prepareKebComponents(components component.List, vals values.Values) ([]*keb.Component, error) {
	var kebComponents []*keb.Component
	all := append(components.Prerequisites, components.Components...)
	for _, c := range all {
		kebComponent := keb.Component{
			Component: c.Name,
			Namespace: c.Namespace,
			URL:       c.URL,
			Version:   c.Version,
		}
		if componentVals, exists := vals[c.Name]; exists {
			valsMap, ok := componentVals.(map[string]interface{})
			if !ok {
				return nil, errors.New("component value must be a map")
			}
			for k, v := range valsMap {
				kebComponent.Configuration = append(kebComponent.Configuration, keb.Configuration{Key: k, Value: v})
			}
		}
		if globalVals, exists := vals["global"]; exists {
			valsMap, ok := globalVals.(map[string]interface{})
			if !ok {
				return nil, errors.New("global value must be a map")
			}
			for k, v := range valsMap {
				kebComponent.Configuration = append(kebComponent.Configuration, keb.Configuration{Key: "global." + k, Value: v})
			}
		}

		kebComponents = append(kebComponents, &kebComponent)
	}

	return kebComponents, nil
}

func prepareKebCluster(opts Options, kebComponents []*keb.Component, delete bool) *cluster.State {
	status := model.ClusterStatusReconciling
	if delete {
		status = model.ClusterStatusDeleting
	}

	return &cluster.State{
		Cluster: &model.ClusterEntity{
			Version:    1,
			RuntimeID:  "local",
			Kubeconfig: string(opts.KubeConfig),
			Metadata:   &keb.Metadata{},
			Contract:   1,
		},
		Configuration: &model.ClusterConfigurationEntity{
			Version:        1,
			RuntimeID:      "local",
			ClusterVersion: 1,
			KymaVersion:    opts.KymaVersion,
			KymaProfile:    opts.KymaProfile,
			Components:     kebComponents,
			Contract:       1,
		},
		Status: &model.ClusterStatusEntity{
			ID:             1,
			RuntimeID:      "local",
			ClusterVersion: 1,
			ConfigVersion:  1,
			Status:         status,
		},
	}
}
