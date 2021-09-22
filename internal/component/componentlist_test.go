package component

import (
	"fmt"
	"github.com/kyma-incubator/reconciler/pkg/reconciler/workspace"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ComponentList_New(t *testing.T) {
	t.Run("From YAML", func(t *testing.T) {
		newCompList(t, "./testdata/componentlist.yaml")
	})
	t.Run("From JSON", func(t *testing.T) {
		newCompList(t, "./testdata/componentlist.json")
	})
	t.Run("Component list from URL", func(t *testing.T) {
		fakeServer := httptest.NewServer(http.FileServer(http.Dir("testdata")))
		defer fakeServer.Close()
		compFile := fmt.Sprintf("%s:/%s", fakeServer.URL, "componentlist.yaml")
		newCompList(t, compFile)
	})
}

func Test_ComponentList_ComponentsFromStrings(t *testing.T) {
	override := make(map[string]interface{})
	override["foo1"] = "bar1"
	t.Run("Add Component in default namespace", func(t *testing.T) {
		compList := FromStrings([]string{"comp4"}, override)
		require.Equal(t, "comp4", compList.Components[0].Component)
		require.Equal(t, "kyma-system", compList.Components[0].Namespace)
	})
	t.Run("Add Component in custom namespace", func(t *testing.T) {
		namespace := "test-namespace"
		compList := FromStrings([]string{"comp4@test-namespace"}, make(map[string]interface{}))
		require.Equal(t, "comp4", compList.Components[0].Component)
		require.Equal(t, namespace, compList.Components[0].Namespace)
	})
	t.Run("Add component with component overrides", func(t *testing.T) {
		overrides := map[string]interface{}{
			"comp4.enabled": true,
		}
		compList := FromStrings([]string{"comp4@test-namespace"}, overrides)
		require.Equal(t, "enabled", compList.Components[0].Configuration[0].Key)
		require.Equal(t, true, compList.Components[0].Configuration[0].Value)
	})
	t.Run("Add component with global overrides", func(t *testing.T) {
		overrides := map[string]interface{}{
			"global.enabled": true,
		}
		compList := FromStrings([]string{"comp1@test-namespace", "comp2@test-namespace"}, overrides)
		require.Equal(t, "comp1", compList.Components[0].Component)
		require.Equal(t, "global.enabled", compList.Components[0].Configuration[0].Key)
		require.Equal(t, true, compList.Components[0].Configuration[0].Value)

		require.Equal(t, "comp2", compList.Components[1].Component)
		require.Equal(t, "global.enabled", compList.Components[1].Configuration[0].Key)
		require.Equal(t, true, compList.Components[1].Configuration[0].Value)
	})
}

func verifyComponentList(t *testing.T, compList List) {

	prereqs := compList.Prerequisites
	comps := compList.Components
	// verify amount of component

	require.Equal(t, 2, len(prereqs), "Different amount of prerequisite component")
	require.Equal(t, 3, len(comps), "Different amount of component")

	// verify names + namespaces of prerequisites
	require.Equal(t, "prereqcomp1", prereqs[0].Component, "Wrong component name")
	require.Equal(t, "prereqns1", prereqs[0].Namespace, "Wrong namespace")
	require.Equal(t, "prereqcomp2", prereqs[1].Component, "Wrong component name")
	require.Equal(t, "testns", prereqs[1].Namespace, "Wrong namespace")

	// verify names + namespaces of component
	require.Equal(t, "comp1", comps[0].Component, "Wrong component name")
	require.Equal(t, "testns", comps[0].Namespace, "Wrong namespace")
	require.Equal(t, "comp2", comps[1].Component, "Wrong component name")
	require.Equal(t, "compns2", comps[1].Namespace, "Wrong namespace")
	require.Equal(t, "comp3", comps[2].Component, "Wrong component name")
	require.Equal(t, "testns", comps[2].Namespace, "Wrong namespace")
}

func newCompList(t *testing.T, compFile string) {
	override := make(map[string]interface{})
	override["foo"] = "bar"
	compList, err := FromFile(&workspace.Workspace{}, compFile, override)
	require.NoError(t, err)
	verifyComponentList(t, compList)
}
