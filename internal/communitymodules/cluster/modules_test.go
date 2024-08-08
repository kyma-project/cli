package cluster

import (
	"bytes"
	"github.com/kyma-project/cli.v3/internal/communitymodules"
	"testing"
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
		var versionedName []string
		versionedName = append(versionedName, "test")
		versionedName = append(versionedName, "1.0.0")

		got := verifyVersion(versionedName, rec)
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
		var versionedName []string
		versionedName = append(versionedName, "test")
		versionedName = append(versionedName, "1.0.2")

		got := verifyVersion(versionedName, rec)
		if got != rec.Versions[1] {
			t.Errorf("verifyVersion() got = %v, want %v", got, nil)
		}
	})
}

func Test_containsModule(t *testing.T) {
	t.Run("Module found", func(t *testing.T) {
		have := "serverless"
		want := []string{"serverless:1.0.0", "keda:1.0.1"}

		got := containsModule(have, want)
		if got[0] != "serverless" {
			t.Errorf("containsModule() got = %v, want %v", got, "test:1.0.0")
		}
	})
	t.Run("Module not found", func(t *testing.T) {
		have := "test"
		want := []string{"serverless:1.0.0", "keda:1.0.1"}

		got := containsModule(have, want)
		if got != nil {
			t.Errorf("containsModule() got = %v, want %v", got, nil)
		}
	})
}

func Test_decodeYaml(t *testing.T) {
	t.Run("Decode YAML", func(t *testing.T) {
		yaml := []byte("apiVersion: v1\nkind: Pod\nmetadata:\n  name: test\nspec:\n  containers:\n  - name: test\n    image: test")
		unstructured, err := decodeYaml(bytes.NewReader(yaml))
		if unstructured[0].GetKind() != "Pod" {
			t.Errorf("decodeYaml() got = %v, want %v", unstructured[0].GetKind(), "Pod")
		}
		if err != nil {
			t.Errorf("decodeYaml() got = %v, want %v", err, nil)
		}
	})
}

// func Test_readCustomConfig(t *testing.T)
