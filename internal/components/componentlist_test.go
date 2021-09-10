package components


import (
	"github.com/kyma-incubator/reconciler/pkg/keb"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ComponentList_New(t *testing.T) {
	t.Run("From YAML", func(t *testing.T) {
		newCompList(t, "./test/data/componentlist.yaml")
	})
	t.Run("From JSON", func(t *testing.T) {
		newCompList(t, "./test/data/componentlist.json")
	})
}

func Test_ComponentList_ComponentsFromStrings(t *testing.T) {
	override := make(map[string]string)
	override["foo1"] = "bar1"
	t.Run("Add Component in default namespace", func(t *testing.T) {
		compList := newCompList(t, "./test/data/componentlist.yaml")
		compList = ComponentsFromStrings([]string{"comp4@defaultNamespace"},  override)
		require.Equal(t, "comp4", compList[5].Component)
		require.Equal(t, defaultNamespace, compList[5].Namespace)
	})
	t.Run("Add Component in custom namespace", func(t *testing.T) {
		compList := newCompList(t, "./test/data/componentlist.yaml")
		namespace := "test-namespace"
		compList = ComponentsFromStrings([]string{"comp4"},  override)
		require.Equal(t, "comp4", compList[5].Component)
		require.Equal(t, namespace, compList[5].Namespace)
	})
}

func verifyComponentList(t *testing.T, compList []keb.Components) {

	require.Equal(t, 5, len(compList), "Different amount of prerequisites and components")

	// verify names + namespaces of prerequisites
	require.Equal(t, "prereqcomp1", compList[0].Component, "Wrong component name")
	require.Equal(t, "prereqns1", compList[0].Namespace, "Wrong namespace")
	require.Equal(t, "prereqcomp2", compList[1].Component, "Wrong component name")
	require.Equal(t, "testns", compList[1].Namespace, "Wrong namespace")

	// verify names + namespaces of components
	require.Equal(t, "comp1", compList[2].Component, "Wrong component name")
	require.Equal(t, "testns", compList[2].Namespace, "Wrong namespace")
	require.Equal(t, "comp2", compList[3].Component, "Wrong component name")
	require.Equal(t, "compns2", compList[3].Namespace, "Wrong namespace")
	require.Equal(t, "comp3", compList[4].Component, "Wrong component name")
	require.Equal(t, "testns", compList[4].Namespace, "Wrong namespace")
}

func newCompList(t *testing.T, compFile string) []keb.Components {
	override := make(map[string]string)
	override["foo"] = "bar"
	compList, err := NewComponentList(compFile, override)
	require.NoError(t, err)
	verifyComponentList(t, compList)
	return compList
}