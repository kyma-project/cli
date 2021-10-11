package component

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var expectedList = List{
	DefaultNamespace: "ns",
	Prerequisites: []Definition{
		{Name: "pre-1", Namespace: "ns-1"},
		{Name: "pre-2", Namespace: "ns"},
	},
	Components: []Definition{
		{Name: "comp-1", Namespace: "ns"},
		{Name: "comp-2", Namespace: "ns-2"},
		{Name: "comp-3", Namespace: "ns"},
	},
}

func TestFromFile(t *testing.T) {
	t.Run("From JSON", func(t *testing.T) {
		list, err := FromFile("testdata/components.json")
		require.NoError(t, err)
		require.Equal(t, expectedList.DefaultNamespace, list.DefaultNamespace)
		require.ElementsMatch(t, expectedList.Prerequisites, list.Prerequisites)
		require.ElementsMatch(t, expectedList.Components, list.Components)
	})

	t.Run("From YAML", func(t *testing.T) {
		list, err := FromFile("testdata/components.yaml")
		require.NoError(t, err)
		require.Equal(t, expectedList.DefaultNamespace, list.DefaultNamespace)
		require.ElementsMatch(t, expectedList.Prerequisites, list.Prerequisites)
		require.ElementsMatch(t, expectedList.Components, list.Components)
	})

	t.Run("From YML", func(t *testing.T) {
		list, err := FromFile("testdata/components.yml")
		require.NoError(t, err)
		require.Equal(t, expectedList.DefaultNamespace, list.DefaultNamespace)
		require.ElementsMatch(t, expectedList.Prerequisites, list.Prerequisites)
		require.ElementsMatch(t, expectedList.Components, list.Components)
	})
}

func TestFromStrings(t *testing.T) {
	t.Run("Add Component in default namespace", func(t *testing.T) {
		list := FromStrings([]string{"comp-1"})
		require.Equal(t, "kyma-system", list.DefaultNamespace)
		require.Equal(t, "comp-1", list.Components[0].Name)
		require.Equal(t, "kyma-system", list.Components[0].Namespace)
	})

	t.Run("Add Component in custom namespace", func(t *testing.T) {
		list := FromStrings([]string{"comp-1@ns-1"})
		require.Equal(t, "kyma-system", list.DefaultNamespace)
		require.Equal(t, "comp-1", list.Components[0].Name)
		require.Equal(t, "ns-1", list.Components[0].Namespace)
	})

	t.Run("Add multiple Components", func(t *testing.T) {
		list := FromStrings([]string{"comp-1@ns-1", "comp-2", "comp-3", "comp-4@ns-2"})
		require.Equal(t, "kyma-system", list.DefaultNamespace)
		require.Equal(t, "comp-1", list.Components[0].Name)
		require.Equal(t, "ns-1", list.Components[0].Namespace)
		require.Equal(t, "comp-2", list.Components[1].Name)
		require.Equal(t, "kyma-system", list.Components[1].Namespace)
		require.Equal(t, "comp-3", list.Components[2].Name)
		require.Equal(t, "kyma-system", list.Components[2].Namespace)
		require.Equal(t, "comp-4", list.Components[3].Name)
		require.Equal(t, "ns-2", list.Components[3].Namespace)
	})
}
