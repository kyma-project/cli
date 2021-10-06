package component

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
	t.Run("From YAML", func(t *testing.T) {
		list, err := FromFile("testdata/components.yaml")
		require.NoError(t, err)
		require.Equal(t, expectedList.DefaultNamespace, list.DefaultNamespace)
		require.ElementsMatch(t, expectedList.Prerequisites, list.Prerequisites)
		require.ElementsMatch(t, expectedList.Components, list.Components)
	})
	t.Run("From JSON", func(t *testing.T) {
		list, err := FromFile("testdata/components.json")
		require.NoError(t, err)
		require.NoError(t, err)
		require.Equal(t, expectedList.DefaultNamespace, list.DefaultNamespace)
		require.ElementsMatch(t, expectedList.Prerequisites, list.Prerequisites)
		require.ElementsMatch(t, expectedList.Components, list.Components)
	})
	t.Run("Component list from URL", func(t *testing.T) {
		fakeServer := httptest.NewServer(http.FileServer(http.Dir("testdata")))
		defer fakeServer.Close()

		url := fmt.Sprintf("%s/%s", fakeServer.URL, "components.yaml")
		list, err := FromFile(url)
		require.NoError(t, err)
		require.NoError(t, err)
		require.Equal(t, expectedList.DefaultNamespace, list.DefaultNamespace)
		require.ElementsMatch(t, expectedList.Prerequisites, list.Prerequisites)
		require.ElementsMatch(t, expectedList.Components, list.Components)
	})
}

//
//func TestFromStrings(t *testing.T) {
//	override := make(map[string]interface{})
//	override["foo1"] = "bar1"
//	t.Run("Add Component in default namespace", func(t *testing.T) {
//		compList := FromStrings([]string{"comp4"})
//		require.Equal(t, "comp4", compList.Components[0].Component)
//		require.Equal(t, "kyma-system", compList.Components[0].Namespace)
//	})
//	t.Run("Add Component in custom namespace", func(t *testing.T) {
//		namespace := "test-namespace"
//		compList := FromStrings([]string{"comp4@test-namespace"})
//		require.Equal(t, "comp4", compList.Components[0].Component)
//		require.Equal(t, namespace, compList.Components[0].Namespace)
//	})
//	t.Run("Add component with component overrides", func(t *testing.T) {
//		overrides := map[string]interface{}{
//			"comp4.enabled": true,
//		}
//		compList := FromStrings([]string{"comp4@test-namespace"})
//		require.Equal(t, "enabled", compList.Components[0].Configuration[0].Key)
//		require.Equal(t, true, compList.Components[0].Configuration[0].Value)
//	})
//	t.Run("Add component with global overrides", func(t *testing.T) {
//		overrides := map[string]interface{}{
//			"global.enabled": true,
//		}
//		compList := FromStrings([]string{"comp1@test-namespace", "comp2@test-namespace"})
//		require.Equal(t, "comp1", compList.Components[0].Component)
//		require.Equal(t, "global.enabled", compList.Components[0].Configuration[0].Key)
//		require.Equal(t, true, compList.Components[0].Configuration[0].Value)
//
//		require.Equal(t, "comp2", compList.Components[1].Component)
//		require.Equal(t, "global.enabled", compList.Components[1].Configuration[0].Key)
//		require.Equal(t, true, compList.Components[1].Configuration[0].Value)
//	})
//}
