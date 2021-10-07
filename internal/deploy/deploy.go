package deploy

import (
	"context"
	"encoding/json"
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
}

type Options struct {
	Components  component.List
	Values      values.Values
	StatusFunc  func(status ComponentStatus)
	KubeConfig  []byte
	KymaVersion string
	KymaProfile string
	Logger      *zap.SugaredLogger
}

func Deploy(opts Options) error {
	kebComponents, err := prepareKebComponents(opts.Components, opts.Values)
	if err != nil {
		return nil
	}

	kebCluster, err := prepareKebCluster(opts, kebComponents)
	if err != nil {
		return nil
	}

	runtimeBuilder := service.NewRuntimeBuilder(reconciliation.NewInMemoryReconciliationRepository(), opts.Logger)
	return runtimeBuilder.RunLocal(opts.Components.PrerequisiteNames(), func(component string, msg *reconciler.CallbackMessage) {
		var state ComponentState
		switch msg.Status {
		case reconciler.StatusSuccess:
			state = Success
		case reconciler.StatusFailed:
			state = RecoverableError
		case reconciler.StatusError:
			state = UnrecoverableError
		}

		opts.StatusFunc(ComponentStatus{component, state})
	}).Run(context.TODO(), kebCluster)
}

func prepareKebComponents(components component.List, vals values.Values) ([]keb.Component, error) {
	var kebComponents []keb.Component
	all := append(components.Prerequisites, components.Components...)
	for _, c := range all {
		kebComponent := keb.Component{
			Component: c.Name,
			Namespace: c.Namespace,
		}
		if componentVals, exists := vals[c.Name]; exists {
			valsMap, ok := componentVals.(values.Values)
			if !ok {
				return nil, errors.New("Component value must be a map")
			}
			for k, v := range valsMap {
				kebComponent.Configuration = append(kebComponent.Configuration, keb.Configuration{Key: k, Value: v})
			}
		}
		if globalVals, exists := vals["global"]; exists {
			valsMap, ok := globalVals.(values.Values)
			if !ok {
				return nil, errors.New("Global value must be a map")
			}
			for k, v := range valsMap {
				kebComponent.Configuration = append(kebComponent.Configuration, keb.Configuration{Key: "global." + k, Value: v})
			}
		}

		kebComponents = append(kebComponents, kebComponent)
	}

	return kebComponents, nil
}

func prepareKebCluster(opts Options, kebComponents []keb.Component) (*cluster.State, error) {
	kebComponentsJSON, err := json.Marshal(kebComponents)
	if err != nil {
		return nil, err
	}

	return &cluster.State{
		Cluster: &model.ClusterEntity{
			Version:    1,
			Cluster:    "local",
			Kubeconfig: string(opts.KubeConfig),
			Contract:   1,
		},
		Configuration: &model.ClusterConfigurationEntity{
			Version:        1,
			Cluster:        "local",
			ClusterVersion: 1,
			KymaVersion:    opts.KymaVersion,
			KymaProfile:    opts.KymaProfile,
			Components:     string(kebComponentsJSON),
			Contract:       1,
		},
		Status: &model.ClusterStatusEntity{
			ID:             1,
			Cluster:        "local",
			ClusterVersion: 1,
			ConfigVersion:  1,
			Status:         model.ClusterStatusReconcilePending,
		},
	}, nil
}
