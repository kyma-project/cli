package deploy

import (
	"github.com/kyma-incubator/reconciler/pkg/cluster"
	"github.com/kyma-incubator/reconciler/pkg/model"
	"github.com/kyma-project/cli/internal/deploy/component"
	"github.com/kyma-project/cli/internal/deploy/values"
	"github.com/stretchr/testify/require"
	"testing"
)

var components = component.List{
	DefaultNamespace: "ns-1",
	Prerequisites:    []component.Definition{
		{"pre-1","ns-1"},
		{"pre-2","ns-2"},
		{"pre-3","ns-1"},
	},
	Components:       []component.Definition{
		{"comp-1","ns-2"},
		{"comp-2","ns-1"},
		{"comp-3","ns-2"},
	} ,
}

var vals = values.Values{
	"global": map[string]interface{}{
		"domainName": "domain-1",
		"tlsCrt":     "tls-crt-1",
		"tlsKey":     "tls-key-1",
	},
	"component": map[string]interface{}{
		"key": "baz",
		"outer": map[string]interface{}{
			"inner": map[string]interface{}{
				"key": "baz",
			},
		},
	},
}

var expected string = ""

func TestPrepareKebComponents(t *testing.T) {
	t.Run("prepareKebComponents", func(t *testing.T) {
		result, err := prepareKebComponents(components, vals)
		require.NoError(t, err)
		require.Equal(t, expected, result)
	})
}

func TestPrepareKebCluster(t *testing.T) {
	t.Run("prepareKebCluster", func(t *testing.T) {
		options := Options{
			Components:  component.List{},
			Values: nil,
			StatusFunc:  nil,
			KubeConfig:  []byte("kubeconfig-1"),
			KymaVersion: "version-1",
			KymaProfile: "profile-1",
			Logger:      nil,
		}

		expectedState :=  &cluster.State{
			Cluster: &model.ClusterEntity{
				Version:    1,
				Cluster:    "local",
				Kubeconfig: "kubeconfig-1",
				Contract:   1,
			},
			Configuration: &model.ClusterConfigurationEntity{
				Version:        1,
				Cluster:        "local",
				ClusterVersion: 1,
				KymaVersion:    "version-1",
				KymaProfile:    "profile-1",
				Components:     "components-json-1",
				Contract:       1,
			},
			Status: &model.ClusterStatusEntity{
				ID:             1,
				Cluster:        "local",
				ClusterVersion: 1,
				ConfigVersion:  1,
				Status:         model.ClusterStatusReconcilePending,
			},
		}

		result := prepareKebCluster(options, "components-json-1")
		require.Equal(t, expectedState, result)
	})
}