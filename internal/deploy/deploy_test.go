package deploy

import (
	"github.com/kyma-incubator/reconciler/pkg/cluster"
	"github.com/kyma-incubator/reconciler/pkg/keb"
	"github.com/kyma-incubator/reconciler/pkg/model"
	"github.com/kyma-project/cli/internal/deploy/component"
	"github.com/kyma-project/cli/internal/deploy/values"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPrepareKebComponents(t *testing.T) {
	var components = component.List{
		DefaultNamespace: "ns-1",
		Prerequisites:    []component.Definition{{Name: "pre-1", Namespace: "ns-1"}},
		Components:       []component.Definition{{Name: "comp-1", Namespace: "ns-2"}, {Name: "comp-2", Namespace: "ns-1"}},
	}

	var vals = values.Values{
		"global": map[string]interface{}{
			"domainName": "domain-1",
			"tlsCrt":     "tls-crt-1",
			"tlsKey":     "tls-key-1",
		},
		"comp-2": map[string]interface{}{
			"key": "baz",
			"outer": map[string]interface{}{
				"inner": map[string]interface{}{
					"key": "baz",
				},
			},
		},
		"comp-3": map[string]interface{}{
			"key": "baz",
		},
	}

	var expected = []keb.Component{
		{
			Component: "pre-1",
			Configuration: []keb.Configuration{
				{Key: "global.domainName", Value: "domain-1"},
				{Key: "global.tlsCrt", Value: "tls-crt-1"},
				{Key: "global.tlsKey", Value: "tls-key-1"},
			},
			Namespace: "ns-1",
		},
		{
			Component: "comp-1",
			Configuration: []keb.Configuration{
				{Key: "global.domainName", Value: "domain-1"},
				{Key: "global.tlsCrt", Value: "tls-crt-1"},
				{Key: "global.tlsKey", Value: "tls-key-1"},
			},
			Namespace: "ns-2",
		},
		{
			Component: "comp-2",
			Configuration: []keb.Configuration{
				{Key: "global.domainName", Value: "domain-1"},
				{Key: "global.tlsCrt", Value: "tls-crt-1"},
				{Key: "global.tlsKey", Value: "tls-key-1"},
				{Key: "key", Value: "baz"},
				{Key: "outer", Value: map[string]interface{}{"inner": map[string]interface{}{"key": "baz"}}},
			},
			Namespace: "ns-1",
		},
	}

	result, err := prepareKebComponents(components, vals)
	require.NoError(t, err)
	require.Len(t, expected, 3)
	require.Equal(t, expected[0].Component, (result)[0].Component)
	require.ElementsMatch(t, expected[0].Configuration, (result)[0].Configuration)
	require.Equal(t, expected[0].Namespace, (result)[0].Namespace)
	require.Equal(t, expected[1].Component, (result)[1].Component)
	require.ElementsMatch(t, expected[1].Configuration, (result)[1].Configuration)
	require.Equal(t, expected[1].Namespace, (result)[1].Namespace)
	require.Equal(t, expected[2].Component, (result)[2].Component)
	require.ElementsMatch(t, expected[2].Configuration, (result)[2].Configuration)
	require.Equal(t, expected[2].Namespace, (result)[2].Namespace)
}
func TestPrepareKebCluster(t *testing.T) {
	var expected = []*keb.Component{
		{
			Component: "comp-1",
			Configuration: []keb.Configuration{
				{Key: "global.domainName", Value: "domain-1"},
			},
			Namespace: "ns-2",
		},
	}

	var expCompWithConf = []*keb.Component{
		{
			URL:       "",
			Version:   "",
			Component: "comp-1",
			Configuration: []keb.Configuration{
				{Key: "global.domainName", Secret: false, Value: "domain-1"},
			},
			Namespace: "ns-2",
		},
	}

	options := Options{
		Components:  component.List{},
		Values:      nil,
		StatusFunc:  nil,
		KubeConfig:  []byte("kubeconfig-1"),
		KymaVersion: "version-1",
		KymaProfile: "profile-1",
		Logger:      nil,
	}

	expectedState := &cluster.State{
		Cluster: &model.ClusterEntity{
			Version:    1,
			RuntimeID:  "local",
			Kubeconfig: "kubeconfig-1",
			Metadata:   &keb.Metadata{},
			Contract:   1,
		},
		Configuration: &model.ClusterConfigurationEntity{
			Version:        1,
			RuntimeID:      "local",
			ClusterVersion: 1,
			KymaVersion:    "version-1",
			KymaProfile:    "profile-1",
			Components:     expCompWithConf,
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

	result := prepareKebCluster(options, expected)
	require.Equal(t, expectedState, result)
}
