package components

import (
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
	override := make(map[string]interface{})
	override["foo1"] = "bar1"
	t.Run("Add Component in default namespace", func(t *testing.T) {
		//compList := newCompList(t, "./test/data/componentlist.yaml")
		compList := FromStrings([]string{"comp4@kyma-system"}, override)
		require.Equal(t, "comp4", compList.Components[0].Component)
		require.Equal(t, "kyma-system", compList.Components[0].Namespace)
	})
	t.Run("Add Component in custom namespace", func(t *testing.T) {
		//compList := newCompList(t, "./test/data/componentlist.yaml")
		namespace := "test-namespace"
		compList := FromStrings([]string{"comp4@test-namespace"}, override)
		//compList.Add("comp4", namespace)
		require.Equal(t, "comp4", compList.Components[0].Component)
		require.Equal(t, namespace, compList.Components[0].Namespace)
	})
}

func verifyComponentList(t *testing.T, compList ComponentList) {

	prereqs := compList.Prerequisites
	comps := compList.Components
	// verify amount of components

	require.Equal(t, 2, len(prereqs), "Different amount of prerequisite components")
	require.Equal(t, 3, len(comps), "Different amount of components")

	// verify names + namespaces of prerequisites
	require.Equal(t, "prereqcomp1", prereqs[0].Component, "Wrong component name")
	require.Equal(t, "prereqns1", prereqs[0].Namespace, "Wrong namespace")
	require.Equal(t, "prereqcomp2", prereqs[1].Component, "Wrong component name")
	require.Equal(t, "testns", prereqs[1].Namespace, "Wrong namespace")

	// verify names + namespaces of components
	require.Equal(t, "comp1", comps[0].Component, "Wrong component name")
	require.Equal(t, "testns", comps[0].Namespace, "Wrong namespace")
	require.Equal(t, "comp2", comps[1].Component, "Wrong component name")
	require.Equal(t, "compns2", comps[1].Namespace, "Wrong namespace")
	require.Equal(t, "comp3", comps[2].Component, "Wrong component name")
	require.Equal(t, "testns", comps[2].Namespace, "Wrong namespace")
}

func newCompList(t *testing.T, compFile string)  {
	override := make(map[string]interface{})
	override["foo"] = "bar"
	compList, err := NewComponentList(compFile, override)
	require.NoError(t, err)
	verifyComponentList(t, compList)
}