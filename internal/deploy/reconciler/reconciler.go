package reconciler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/reconciler/pkg/cluster"
	"github.com/kyma-incubator/reconciler/pkg/keb"
	"github.com/kyma-incubator/reconciler/pkg/model"
	"github.com/kyma-incubator/reconciler/pkg/scheduler/reconciliation"
	scheduleService "github.com/kyma-incubator/reconciler/pkg/scheduler/service"
	"github.com/kyma-project/cli/internal/deploy/component"
	"github.com/kyma-project/cli/internal/deploy/values"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type DeploymentArgs struct {
	Components component.List
	Values values.Values
	PrintStatus func()
	KubeConfig []byte
	KymaVersion string
	KymaProfile string
	Logger *zap.SugaredLogger
}

func Deploy(deployment DeploymentArgs) error{

	kebComponentsJSON, err := prepareKebComponents(deployment.Components, deployment.Values)
	if err != nil {
		return errors.Wrap(err, "Failed to prepare components to install")
	}
	fmt.Println(kebComponentsJSON)

	runtimeBuilder := scheduleService.NewRuntimeBuilder(reconciliation.NewInMemoryReconciliationRepository(), deployment.Logger)
	err = runtimeBuilder.RunLocal(deployment.Components.PrerequisiteNames(), deployment.PrintStatus).Run(context.TODO(),
		&cluster.State{
			Cluster: &model.ClusterEntity{
				Version:    1,
				Cluster:    "local",
				Kubeconfig: string(deployment.KubeConfig),
				Contract:   1,
			},
			Configuration: &model.ClusterConfigurationEntity{
				Version:        1,
				Cluster:        "local",
				ClusterVersion: 1,
				KymaVersion:    deployment.KymaVersion,
				KymaProfile:    deployment.KymaProfile,
				Components:     kebComponentsJSON,
				Contract:       1,
			},
			Status: &model.ClusterStatusEntity{
				ID:             1,
				Cluster:        "local",
				ClusterVersion: 1,
				ConfigVersion:  1,
				Status:         model.ClusterStatusReconcilePending,
			},
		})

	return err
}


func prepareKebComponents(components component.List, vals values.Values) (string, error) {
	var kebComponents []keb.Component
	all := append(components.Prerequisites, components.Components...)
	for _, c := range all {
		kebComponent := keb.Component{
			Component: c.Name,
			Namespace: c.Namespace,
		}
		if componentVals, exists := vals[c.Name]; exists {
			for k, v := range componentVals.(map[string]interface{}) {
				kebComponent.Configuration = append(kebComponent.Configuration, keb.Configuration{Key: k, Value: v})
			}
		}
		if globalVals, exists := vals["global"]; exists {
			for k, v := range globalVals.(map[string]interface{}) {
				kebComponent.Configuration = append(kebComponent.Configuration, keb.Configuration{Key: "global." + k, Value: v})
			}
		}

		kebComponents = append(kebComponents, kebComponent)
	}

	kebComponentsJSON, err := json.Marshal(kebComponents)
	if err != nil {
		return "", err
	}
	return string(kebComponentsJSON), nil
}
