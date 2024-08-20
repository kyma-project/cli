package cluster

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/communitymodules"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	discovery_fake "k8s.io/client-go/discovery/fake"
	dynamic_fake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

var (
	fakeIstioCR = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "fakegroup/v1",
			"kind":       "istio",
			"metadata": map[string]interface{}{
				"name":      "fake-istio",
				"namespace": "kyma-system",
			},
		},
	}

	fakeServerlessCR = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "fakegroup/v1",
			"kind":       "istio",
			"metadata": map[string]interface{}{
				"name":      "fake-istio",
				"namespace": "kyma-system",
			},
		},
	}

	fakeAvailableModules = communitymodules.Modules{
		{
			Name: "serverless",
			Versions: []communitymodules.Version{
				{
					Version: "0.0.2",
				},
				{
					Version: "0.0.1",
					CR:      fakeIstioCR.Object,
					Resources: []communitymodules.Resource{
						{
							//empty
						},
						{
							//empty
						},
					},
				},
			},
		},
		{
			Name: "eventing",
			Versions: []communitymodules.Version{
				{
					Version: "1.0.0",
				},
			},
		},
		{
			Name: "istio",
			Versions: []communitymodules.Version{
				{
					Version: "0.1.0",
					CR:      fakeServerlessCR.Object,
					Resources: []communitymodules.Resource{
						{
							//empty
						},
						{
							//empty
						},
						{
							//empty
						},
					},
				},
			},
		},
	}
)

func Test_applySpecifiedModules(t *testing.T) {
	t.Run("Apply fake serverless and istio modules", func(t *testing.T) {
		fakerootlessdynamic := &rootlessdynamicMock{}

		err := applySpecifiedModules(context.Background(), fakerootlessdynamic, []ModuleInfo{
			{"serverless", "0.0.1"},
			{"istio", "0.1.0"},
		}, []unstructured.Unstructured{}, fakeAvailableModules)
		require.Nil(t, err)

		require.Len(t, fakerootlessdynamic.appliedObjects, 7)
		require.Contains(t, fakerootlessdynamic.appliedObjects, fakeServerlessCR)
		require.Contains(t, fakerootlessdynamic.appliedObjects, fakeIstioCR)
	})

	t.Run("Apply fake serverless and istio with CRs", func(t *testing.T) {
		fakerootlessdynamic := &rootlessdynamicMock{}

		istioCR := fakeIstioCR
		istioCR.Object["spec"] = "test"

		err := applySpecifiedModules(context.Background(), fakerootlessdynamic, []ModuleInfo{
			{Name: "istio"},
		}, []unstructured.Unstructured{
			istioCR,
		}, fakeAvailableModules)
		require.Nil(t, err)

		require.Len(t, fakerootlessdynamic.appliedObjects, 4)
		require.Contains(t, fakerootlessdynamic.appliedObjects, istioCR)
	})

	t.Run("Apply client error", func(t *testing.T) {
		fakerootlessdynamic := &rootlessdynamicMock{
			returnErr: clierror.New("test error"),
		}

		err := applySpecifiedModules(context.Background(), fakerootlessdynamic, []ModuleInfo{
			{"serverless", "0.0.1"},
		}, []unstructured.Unstructured{}, fakeAvailableModules)
		require.Equal(t, clierror.Wrap(
			errors.New("test error"),
			clierror.New("failed to apply module resources"),
		), err)
	})
}

func TestParseModules(t *testing.T) {
	t.Run("parse input", func(t *testing.T) {
		input := []string{"test", "", "test2:1.2.3"}

		moduleInfoList := ParseModules(input)
		require.Len(t, moduleInfoList, 2)
		require.Contains(t, moduleInfoList, ModuleInfo{"test", ""})
		require.Contains(t, moduleInfoList, ModuleInfo{"test2", "1.2.3"})
	})
}

func fixFakeRootlessDynamicClient() rootlessdynamic.Interface {
	return rootlessdynamic.NewClient(
		dynamic_fake.NewSimpleDynamicClient(scheme.Scheme),
		&discovery_fake.FakeDiscovery{},
	)
}

func Test_verifyVersion(t *testing.T) {
	t.Run("Version found", func(t *testing.T) {
		rec := communitymodules.Module{
			Name: "test",
			Versions: []communitymodules.Version{
				{
					Version: "1.0.0",
				},
				{
					Version: "1.0.1",
				},
			},
		}
		moduleInfo := ModuleInfo{
			Name:    "test",
			Version: "1.0.0",
		}

		got := verifyVersion(moduleInfo, rec)
		require.Equal(t, got, rec.Versions[0])
	})
	t.Run("Version not found", func(t *testing.T) {
		rec := communitymodules.Module{
			Name: "test",
			Versions: []communitymodules.Version{
				{
					Version: "1.0.0",
				},
				{
					Version: "1.0.1",
				},
			},
		}
		moduleInfo := ModuleInfo{
			Name:    "test",
			Version: "1.0.2",
		}

		got := verifyVersion(moduleInfo, rec)
		require.Equal(t, got, rec.Versions[1])
	})
}

func Test_containsModule(t *testing.T) {
	t.Run("Module found", func(t *testing.T) {
		have := "serverless"
		want := []ModuleInfo{
			{"serverless", "1.0.0"},
			{"keda", "1.0.1"},
		}

		got := containsModule(have, want)
		if got.Name != "serverless" {
			t.Errorf("containsModule() got = %v, want %v", got, "test:1.0.0")
		}
	})
	t.Run("Module not found", func(t *testing.T) {
		have := "test"
		want := []ModuleInfo{
			{"Serverless", "1.0.0"},
			{"Keda", "1.0.1"},
		}

		got := containsModule(have, want)
		if got != nil {
			t.Errorf("containsModule() got = %v, want %v", got, nil)
		}
	})
}

type rootlessdynamicMock struct {
	returnErr      clierror.Error
	appliedObjects []unstructured.Unstructured
}

func (m *rootlessdynamicMock) Apply(_ context.Context, obj *unstructured.Unstructured) clierror.Error {
	m.appliedObjects = append(m.appliedObjects, *obj)
	return m.returnErr
}

func (m *rootlessdynamicMock) ApplyMany(_ context.Context, objs []unstructured.Unstructured) clierror.Error {
	m.appliedObjects = append(m.appliedObjects, objs...)
	return m.returnErr
}
