package cluster

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/communitymodules"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	"github.com/stretchr/testify/require"
	discovery_fake "k8s.io/client-go/discovery/fake"
	dynamic_fake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

func TestParseModules(t *testing.T) {
	t.Run("parse input", func(t *testing.T) {
		input := []string{"test", "", "test2:1.2.3"}

		moduleInfoList := ParseModules(input)
		require.Len(t, moduleInfoList, 2)
		require.Contains(t, moduleInfoList, ModuleInfo{"test", ""})
		require.Contains(t, moduleInfoList, ModuleInfo{"test2", "1.2.3"})
	})
}

func fixTestRootlessDynamicClient() rootlessdynamic.Interface {
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
		if got != rec.Versions[0] {
			t.Errorf("verifyVersion() got = %v, want %v", got, rec.Versions[0])
		}
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
		if got != rec.Versions[1] {
			t.Errorf("verifyVersion() got = %v, want %v", got, nil)
		}
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
