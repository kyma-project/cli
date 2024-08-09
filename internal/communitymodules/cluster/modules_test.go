package cluster

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/communitymodules"
)

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
