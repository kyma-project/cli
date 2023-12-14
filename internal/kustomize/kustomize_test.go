package kustomize

import (
	"os"
	"testing"

	"github.com/kyma-project/lifecycle-manager/api/shared"
	"sigs.k8s.io/kustomize/api/filters/fieldspec"
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/resid"

	"github.com/stretchr/testify/require"
)

func TestParseKustomization(t *testing.T) {
	// URL
	k, err := ParseKustomization("https://my-site.com/owner/repo/sub/path/to/kustomization")
	require.NoError(t, err)
	require.Equal(t,
		Definition{Name: "repo", Ref: "main", Location: "https://my-site.com/owner/repo/sub/path/to/kustomization"}, k)

	// Path
	k, err = ParseKustomization("/Path/to/my/repo/sub/path/kustomization")
	require.NoError(t, err)
	require.Equal(t, Definition{
		Name:     "/Path/to/my/repo/sub/path/kustomization",
		Ref:      "local",
		Location: "/Path/to/my/repo/sub/path/kustomization",
	}, k)

	// URL with ref
	k, err = ParseKustomization("https://my-site.com/owner/repo/sub/path/to/kustomization@branchX")
	require.NoError(t, err)
	require.Equal(t,
		Definition{Name: "repo", Ref: "branchX", Location: "https://my-site.com/owner/repo/sub/path/to/kustomization"},
		k)

	// More than one ref
	_, err = ParseKustomization("https://my-site.com/owner/repo/sub/path/to/kustomization@branchX@branchY")
	require.Error(t, err)

	// URL not a valid repo
	_, err = ParseKustomization("https://my-site.com/owner@branchX")
	require.Error(t, err)
}

func TestBuildMany(t *testing.T) {
	definition := Definition{
		Name:     "",
		Ref:      localRef,
		Location: "./testdata/given",
	}

	result, err := BuildMany([]Definition{definition}, nil)
	require.NoError(t, err)
	require.NotEmpty(t, result)
}

func TestBuildManyWithFilter(t *testing.T) {
	definition := Definition{
		Name:     "",
		Ref:      localRef,
		Location: "./testdata/given",
	}

	filter := fieldspec.Filter{
		FieldSpec: types.FieldSpec{
			Gvk: resid.Gvk{
				Group: shared.OperatorGroup,
				Kind:  string(shared.ModuleTemplateKind),
			},
			Path:               "spec/target",
			CreateIfNotPresent: false,
		},
		SetValue: filtersutil.SetScalar("some-target"),
	}

	result, err := BuildMany([]Definition{definition}, []kio.Filter{kio.FilterAll(filter)})
	require.NoError(t, err)
	require.NotEmpty(t, result)

	expected, err := os.ReadFile("./testdata/expected/manifest.yaml")
	require.NoError(t, err)

	require.Equal(t, string(expected), string(result))
}
