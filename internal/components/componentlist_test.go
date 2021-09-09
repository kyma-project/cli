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

func verifyComponentList(t *testing.T, compList []keb.Components) {
	//prereqs := compList.Prerequisites
	//comps := compList.Components
	// verify amount of components

	//require.Equal(t, 2, len(prereqs), "Different amount of prerequisite components")
	//require.Equal(t, 3, len(comps), "Different amount of components")

	// verify names + namespaces of prerequisites
	require.Equal(t, "prereqcomp1", compList[0].Component, "Wrong component name")
	require.Equal(t, "prereqns1", compList[0].Namespace, "Wrong namespace")
	require.Equal(t, "prereqcomp2", compList[1].Component, "Wrong component name")
	require.Equal(t, "testns", compList[1].Namespace, "Wrong namespace")

	// verify names + namespaces of components
	require.Equal(t, "comp1", compList[0].Component, "Wrong component name")
	require.Equal(t, "testns", compList[0].Namespace, "Wrong namespace")
	require.Equal(t, "comp2", compList[1].Component, "Wrong component name")
	require.Equal(t, "compns2", compList[1].Namespace, "Wrong namespace")
	require.Equal(t, "comp3", compList[2].Component, "Wrong component name")
	require.Equal(t, "testns", compList[2].Namespace, "Wrong namespace")
}

func newCompList(t *testing.T, compFile string) []keb.Components {
	var override map[string]string
	override["foo"] = "bar"
	compList, err := NewComponentList(compFile, override)
	require.NoError(t, err)
	verifyComponentList(t, compList)
	return compList
}