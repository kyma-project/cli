package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/communitymodules"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	fakeIstioCR = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "fakegroup/v1",
			"kind":       "istio",
			"metadata": map[string]interface{}{
				"name":      "fake-istio",
				"namespace": "kyma-system",
			},
		},
	}

	fakeServerlessCR = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "fakegroup/v1",
			"kind":       "serverless",
			"metadata": map[string]interface{}{
				"name":      "fake-serverless",
				"namespace": "kyma-system",
			},
		},
	}

	fakeModuleDetails = []moduleDetails{
		{
			name:    "serverless",
			version: "1.0.0",
			cr:      fakeServerlessCR,
			resources: []unstructured.Unstructured{
				{
					// empty
				},
				{
					// empty
				},
				{
					// empty
				},
			},
		},
		{
			name:    "istio",
			version: "0.1.0",
			cr:      fakeIstioCR,
			resources: []unstructured.Unstructured{
				{
					// empty
				},
				{
					// empty
				},
			},
		},
	}
)

func Test_downloadSpecifiedModules(t *testing.T) {
	t.Run("Download module details", func(t *testing.T) {
		mux := &http.ServeMux{}
		mux.HandleFunc("/istio/cr", func(w http.ResponseWriter, _ *http.Request) {
			bytes, err := json.Marshal(fakeIstioCR.Object)
			require.NoError(t, err)

			_, err = w.Write(bytes)
			require.NoError(t, err)
		})
		mux.HandleFunc("/istio/resources", func(w http.ResponseWriter, _ *http.Request) {
			data := "\n---\n---\n---\n" // three resources

			_, err := w.Write([]byte(data))
			require.NoError(t, err)
		})

		server := httptest.NewServer(mux)
		defer server.Close()

		availableModules := fixFakeAvailableDetails(
			fmt.Sprintf("%s/istio/cr", server.URL),
			fmt.Sprintf("%s/istio/resources", server.URL),
		)

		modules, err := downloadSpecifiedModules([]ModuleInfo{{"istio", "0.0.1"}}, availableModules)
		require.Nil(t, err)
		require.Len(t, modules, 1)
		require.Equal(t, fakeIstioCR, modules[0].cr)
		require.Len(t, modules[0].resources, 3)
	})

	t.Run("broken cr url", func(t *testing.T) {
		availableModules := fixFakeAvailableDetails(
			"does-not-exist",
			"does-not-exist",
		)

		module, err := downloadSpecifiedModules([]ModuleInfo{{"istio", "0.0.1"}}, availableModules)
		require.Equal(t, clierror.Wrap(
			errors.New("Get \"does-not-exist\": unsupported protocol scheme \"\""),
			clierror.New("failed to download cr from 'does-not-exist' url"),
		), err)
		require.Empty(t, module)
	})

	t.Run("bad response from resources url", func(t *testing.T) {
		mux := &http.ServeMux{}
		mux.HandleFunc("/istio/cr", func(w http.ResponseWriter, _ *http.Request) {
			bytes, err := json.Marshal(fakeIstioCR.Object)
			require.NoError(t, err)

			_, err = w.Write(bytes)
			require.NoError(t, err)
		})
		mux.HandleFunc("/istio/resources", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(404)
		})

		server := httptest.NewServer(mux)
		defer server.Close()

		availableModules := fixFakeAvailableDetails(
			fmt.Sprintf("%s/istio/cr", server.URL),
			fmt.Sprintf("%s/istio/resources", server.URL),
		)

		modules, err := downloadSpecifiedModules([]ModuleInfo{{"istio", "0.0.1"}}, availableModules)
		require.Equal(t, clierror.Wrap(
			errors.New("unexpected status code '404'"),
			clierror.New(fmt.Sprintf("failed to download manifest from '%s/istio/resources' url", server.URL)),
		), err)
		require.Empty(t, modules)
	})
}

func Test_applySpecifiedModules(t *testing.T) {
	t.Run("Apply fake serverless and istio modules", func(t *testing.T) {
		fakerootlessdynamic := &rootlessdynamicMock{}

		err := applySpecifiedModules(context.Background(),
			fakerootlessdynamic, fakeModuleDetails, []unstructured.Unstructured{})
		require.Nil(t, err)

		require.Len(t, fakerootlessdynamic.appliedObjects, 7)
		require.Contains(t, fakerootlessdynamic.appliedObjects, fakeServerlessCR)
		require.Contains(t, fakerootlessdynamic.appliedObjects, fakeIstioCR)
	})

	t.Run("Apply fake serverless and istio with CRs", func(t *testing.T) {
		fakerootlessdynamic := &rootlessdynamicMock{}

		istioCR := fakeIstioCR
		istioCR.Object["spec"] = "test"

		err := applySpecifiedModules(context.Background(),
			fakerootlessdynamic, fakeModuleDetails, []unstructured.Unstructured{
				istioCR,
			})
		require.Nil(t, err)

		require.Len(t, fakerootlessdynamic.appliedObjects, 7)
		require.Contains(t, fakerootlessdynamic.appliedObjects, istioCR)
		require.Contains(t, fakerootlessdynamic.appliedObjects, fakeServerlessCR)
	})

	t.Run("Apply client error", func(t *testing.T) {
		fakerootlessdynamic := &rootlessdynamicMock{
			returnErr: errors.New("test error"),
		}

		err := applySpecifiedModules(context.Background(),
			fakerootlessdynamic, fakeModuleDetails, []unstructured.Unstructured{})
		require.Equal(t, clierror.Wrap(
			errors.New("test error"),
			clierror.New("failed to apply module resources"),
		), err)
	})
}

func TestParseModules(t *testing.T) {
	t.Run("parse input", func(t *testing.T) {
		input := []string{"test", "", "test2:1.2.3"}

		moduleInfoList := ParseModules(input)
		require.Len(t, moduleInfoList, 2)
		require.Contains(t, moduleInfoList, ModuleInfo{"test", ""})
		require.Contains(t, moduleInfoList, ModuleInfo{"test2", "1.2.3"})
	})
}

func Test_verifyVersion(t *testing.T) {
	t.Run("Version found", func(t *testing.T) {
		versions := []communitymodules.Version{
			{
				Version: "1.0.0",
			},
			{
				Version: "1.0.1",
			},
		}
		moduleInfo := ModuleInfo{
			Name:    "test",
			Version: "1.0.0",
		}

		got := getDesiredVersion(moduleInfo, versions)
		require.Equal(t, got, versions[0])
	})
	t.Run("Version not found", func(t *testing.T) {
		versions := []communitymodules.Version{
			{
				Version: "1.0.0",
			},
			{
				Version: "1.0.1",
			},
		}
		moduleInfo := ModuleInfo{
			Name:    "test",
			Version: "1.0.2",
		}

		got := getDesiredVersion(moduleInfo, versions)
		require.Equal(t, got, versions[1])
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

func Test_removeSpecifiedModules(t *testing.T) {
	t.Run("Remove module", func(t *testing.T) {
		fakerootlessdynamic := &rootlessdynamicMock{}

		err := removeSpecifiedModules(context.Background(), fakerootlessdynamic, fakeModuleDetails)
		require.Nil(t, err)
		require.Len(t, fakerootlessdynamic.removedObjects, 7)
		require.Contains(t, fakerootlessdynamic.removedObjects, fakeServerlessCR)
		require.Contains(t, fakerootlessdynamic.removedObjects, fakeIstioCR)
	})

	t.Run("Remove module cr error", func(t *testing.T) {
		fakerootlessdynamic := &rootlessdynamicMock{
			returnErr: errors.New("test error"),
		}

		err := removeSpecifiedModules(context.Background(), fakerootlessdynamic, fakeModuleDetails)
		require.Equal(t, clierror.Wrap(
			errors.New("test error"),
			clierror.New("failed to remove module cr"),
		), err)
	})
}

func Test_retryUntilRemoved(t *testing.T) {
	t.Run("Object removed", func(t *testing.T) {
		fakerootlessdynamic := &rootlessdynamicMock{}

		err := retryUntilRemoved(context.Background(), fakerootlessdynamic, fakeModuleDetails[0], 1)
		require.Nil(t, err)
	})

	t.Run("Failed to remove object", func(t *testing.T) {
		fakerootlessdynamic := &rootlessdynamicMock{
			returnErr: errors.New("test error"),
		}

		err := retryUntilRemoved(context.Background(), fakerootlessdynamic, fakeModuleDetails[0], 1)
		require.Contains(t, err.String(), "failed to remove module cr")
	})

	t.Run("Object still exists", func(t *testing.T) {
		fakerootlessdynamic := &rootlessdynamicMock{
			appliedObjects: []unstructured.Unstructured{fakeServerlessCR},
		}

		err := retryUntilRemoved(context.Background(), fakerootlessdynamic, fakeModuleDetails[0], 1)
		require.Contains(t, err.String(), "object still exists")
	})
}

func fixFakeAvailableDetails(istioCRURL, istioResourcesURL string) communitymodules.Modules {
	return communitymodules.Modules{
		{
			Name: "serverless",
			Versions: []communitymodules.Version{
				{
					Version: "0.0.1",
				},
			},
		},
		{
			Name: "istio",
			Versions: []communitymodules.Version{
				{
					Version: "0.0.2",
				},
				{
					Version:        "0.0.1",
					DeploymentYaml: istioResourcesURL,
					CrYaml:         istioCRURL,
				},
			},
		},
		{
			Name: "eventing",
			Versions: []communitymodules.Version{
				{
					Version: "1.0.0",
				},
			},
		},
	}
}

type rootlessdynamicMock struct {
	returnErr      error
	appliedObjects []unstructured.Unstructured
	removedObjects []unstructured.Unstructured
}

func (m *rootlessdynamicMock) Apply(_ context.Context, obj *unstructured.Unstructured) error {
	m.appliedObjects = append(m.appliedObjects, *obj)
	return m.returnErr
}

func (m *rootlessdynamicMock) ApplyMany(_ context.Context, objs []unstructured.Unstructured) error {
	m.appliedObjects = append(m.appliedObjects, objs...)
	return m.returnErr
}

func (m *rootlessdynamicMock) Get(_ context.Context, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	if len(m.appliedObjects) == 0 {
		return nil, m.returnErr
	}
	return obj, m.returnErr
}

func (m *rootlessdynamicMock) Remove(_ context.Context, obj *unstructured.Unstructured) error {
	m.removedObjects = append(m.removedObjects, *obj)
	return m.returnErr
}

func (m *rootlessdynamicMock) RemoveMany(_ context.Context, objs []unstructured.Unstructured) error {
	m.removedObjects = append(m.removedObjects, objs...)
	return m.returnErr
}
