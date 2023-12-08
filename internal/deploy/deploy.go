package deploy

import (
	"github.com/kyma-incubator/reconciler/pkg/cluster"
	"github.com/kyma-incubator/reconciler/pkg/keb"
	"github.com/kyma-incubator/reconciler/pkg/model"
	"github.com/kyma-project/cli/internal/deploy/component"
	"github.com/kyma-project/cli/internal/deploy/values"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ComponentState string

type ComponentStatus struct {
	Component string
	State     ComponentState
	Error     error
	Manifest  *string
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
	DeleteStrategy string
	DryRun         bool
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
