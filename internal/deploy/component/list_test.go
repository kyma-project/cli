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
		{Name: "pre-3", Namespace: "ns", URL: "some-url", Version: "1.2.3"},
	},
	Components: []Definition{
		{Name: "comp-1", Namespace: "ns"},
		{Name: "comp-2", Namespace: "ns-2"},
		{Name: "comp-3", Namespace: "ns"},
		{Name: "comp-4", Namespace: "ns-2", URL: "some-url", Version: "1.2.3"},
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
		list, _ := FromStrings([]string{"comp-1"})
		require.Equal(t, "kyma-system", list.DefaultNamespace)
		require.Equal(t, "comp-1", list.Components[0].Name)
		require.Equal(t, "kyma-system", list.Components[0].Namespace)
	})

	t.Run("Add Component in custom namespace", func(t *testing.T) {
		list, _ := FromStrings([]string{"comp-1@ns-1"})
		require.Equal(t, "kyma-system", list.DefaultNamespace)
		require.Equal(t, "comp-1", list.Components[0].Name)
		require.Equal(t, "ns-1", list.Components[0].Namespace)
	})

	t.Run("Add multiple Components", func(t *testing.T) {
		list, _ := FromStrings([]string{"comp-1@ns-1", "comp-2", "comp-3", "comp-4@ns-2"})
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

	t.Run("Add Component via JSON format notation", func(t *testing.T) {
		list, _ := FromStrings([]string{"{\"name\": \"comp-1\",\"namespace\": \"ns-1\",\"url\": \"some-url\",\"version\": \"1.2.3\"}"})
		require.Equal(t, "kyma-system", list.DefaultNamespace)
		require.Equal(t, "comp-1", list.Components[0].Name)
		require.Equal(t, "ns-1", list.Components[0].Namespace)
		require.Equal(t, "some-url", list.Components[0].URL)
		require.Equal(t, "1.2.3", list.Components[0].Version)
	})

	t.Run("Add Component via JSON format notation with default namespace and no version", func(t *testing.T) {
		list, _ := FromStrings([]string{"{\"name\": \"comp-1\",\"url\": \"some-url\"}"})
		require.Equal(t, "kyma-system", list.DefaultNamespace)
		require.Equal(t, "comp-1", list.Components[0].Name)
		require.Equal(t, "kyma-system", list.Components[0].Namespace)
		require.Equal(t, "some-url", list.Components[0].URL)
		require.Empty(t, list.Components[0].Version)
	})
 
	t.Run("Add multiple Components via different notations", func(t *testing.T) {
		list, _ := FromStrings([]string{"comp-1@ns-1", "{\"name\": \"comp-2\",\"url\": \"some-url\"}", "comp-3", "{\"name\": \"comp-4\",\"namespace\": \"ns-2\",\"url\": \"some-url-2\",\"version\": \"3.2\"}"})
		require.Equal(t, "kyma-system", list.DefaultNamespace)
		require.Equal(t, "comp-1", list.Components[0].Name)
		require.Equal(t, "ns-1", list.Components[0].Namespace)
		require.Equal(t, "comp-2", list.Components[1].Name)
		require.Equal(t, "kyma-system", list.Components[1].Namespace)
		require.Equal(t, "some-url", list.Components[1].URL)
		require.Empty(t, list.Components[1].Version)
		require.Equal(t, "comp-3", list.Components[2].Name)
		require.Equal(t, "kyma-system", list.Components[2].Namespace)
		require.Equal(t, "comp-4", list.Components[3].Name)
		require.Equal(t, "ns-2", list.Components[3].Namespace)
		require.Equal(t, "some-url-2", list.Components[3].URL)
		require.Equal(t, "3.2", list.Components[3].Version)
	})
}
