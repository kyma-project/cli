package deploy

import (
	"context"
	"github.com/kyma-incubator/reconciler/pkg/cluster"
	"github.com/kyma-incubator/reconciler/pkg/keb"
	"github.com/kyma-incubator/reconciler/pkg/model"
	"github.com/kyma-incubator/reconciler/pkg/reconciler"
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
}

type Options struct {
	Components     component.List
	Values         values.Values
	StatusFunc     func(status ComponentStatus)
	KubeConfig     []byte
	KymaVersion    string
	KymaProfile    string
	Logger         *zap.SugaredLogger
	WorkerPoolSize int
}

func Deploy(opts Options) (error, *service.ReconciliationResult) {
	kebComponents, err := prepareKebComponents(opts.Components, opts.Values)
	if err != nil {
		return nil, nil
	}

	kebCluster := prepareKebCluster(opts, kebComponents)

	runtimeBuilder := service.NewRuntimeBuilder(reconciliation.NewInMemoryReconciliationRepository(), opts.Logger)
	reconcilationResult, err := runtimeBuilder.RunLocal(opts.Components.PrerequisiteNames(), func(component string, msg *reconciler.CallbackMessage) {
		var state ComponentState
		var errorRecieved error
		switch msg.Status {
		case reconciler.StatusSuccess:
			state = Success
			errorRecieved = nil
		case reconciler.StatusFailed:
			errorRecieved = errors.Errorf("%s", msg.Error)
			state = RecoverableError
		case reconciler.StatusError:
			errorRecieved = errors.Errorf("%s", msg.Error)
			state = UnrecoverableError
		}

		opts.StatusFunc(ComponentStatus{component, state, errorRecieved})
	}).WithWorkerPoolSize(opts.WorkerPoolSize).Run(context.TODO(), kebCluster)

	return err, reconcilationResult
}

func prepareKebComponents(components component.List, vals values.Values) ([]*keb.Component, error) {
	var kebComponents []*keb.Component
	all := append(components.Prerequisites, components.Components...)
	for _, c := range all {
		kebComponent := keb.Component{
			Component: c.Name,
			Namespace: c.Namespace,
		}
		if componentVals, exists := vals[c.Name]; exists {
			valsMap, ok := componentVals.(map[string]interface{})
			if !ok {
				return nil, errors.New("Component value must be a map")
			}
			for k, v := range valsMap {
				kebComponent.Configuration = append(kebComponent.Configuration, keb.Configuration{Key: k, Value: v})
			}
		}
		if globalVals, exists := vals["global"]; exists {
			valsMap, ok := globalVals.(map[string]interface{})
			if !ok {
				return nil, errors.New("Global value must be a map")
			}
			for k, v := range valsMap {
				kebComponent.Configuration = append(kebComponent.Configuration, keb.Configuration{Key: "global." + k, Value: v})
			}
		}

		kebComponents = append(kebComponents, &kebComponent)
	}

	return kebComponents, nil
}

func prepareKebCluster(opts Options, kebComponents []*keb.Component) *cluster.State {
	return &cluster.State{
		Cluster: &model.ClusterEntity{
			Version:    1,
			RuntimeID:  "local",
			Kubeconfig: string(opts.KubeConfig),
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
			Status:         model.ClusterStatusReconcilePending,
		},
	}
}
