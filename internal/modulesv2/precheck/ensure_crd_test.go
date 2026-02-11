package precheck

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	kubefake "github.com/kyma-project/cli.v3/internal/kube/fake"
	m2fake "github.com/kyma-project/cli.v3/internal/modulesv2/fake"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	testCRDName       = "moduletemplates.operator.kyma-project.io"
	testCRDAPIVersion = "apiextensions.k8s.io/v1"
	testCRDKind       = "CustomResourceDefinition"
)

var crdGroupResource = schema.GroupResource{Group: "apiextensions.k8s.io", Resource: "customresourcedefinitions"}

func minimalCRDUnstructuredWithVersion(version string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": testCRDAPIVersion,
		"kind":       testCRDKind,
		"metadata":   map[string]any{"name": testCRDName},
		"spec": map[string]any{
			"group": "operator.kyma-project.io",
			"names": map[string]any{
				"kind":     "ModuleTemplate",
				"listKind": "ModuleTemplateList",
				"plural":   "moduletemplates",
				"singular": "moduletemplate",
			},
			"scope": "Namespaced",
			"versions": []any{
				map[string]any{"name": version, "served": true, "storage": true, "schema": map[string]any{"openAPIV3Schema": map[string]any{"type": "object"}}},
			},
		},
	}}
}

func minimalCRDYAML(version string) string {
	return fmt.Sprintf(`apiVersion: %s
kind: %s
metadata:
  name: %s
spec:
  group: operator.kyma-project.io
  names:
    kind: ModuleTemplate
    listKind: ModuleTemplateList
    plural: moduletemplates
    singular: moduletemplate
  scope: Namespaced
  versions:
    - name: %s
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
`, testCRDAPIVersion, testCRDKind, testCRDName, version)
}

func startCRDServer(t *testing.T, body string, status int) (*httptest.Server, *int) {
	t.Helper()
	callCount := 0
	srv := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				callCount++
				w.WriteHeader(status)
				fmt.Fprint(w, body)
			}))
	t.Cleanup(srv.Close)
	return srv, &callCount
}

func newTestCRDEnsurer(rootless *kubefake.RootlessDynamicClient, isManaged bool, remoteURL string) *CRDEnsurer {
	kubeClient := &kubefake.KubeClient{TestRootlessDynamicInterface: rootless}
	metadataRepo := &m2fake.ClusterMetadataRepository{IsManagedByKLM: isManaged}
	return NewCRDEnsurer(kubeClient, metadataRepo, http.DefaultClient, remoteURL)
}

func TestEnsureCRD_SkipsWhenManagedByKLM(t *testing.T) {
	t.Parallel()
	rootlessClient := &kubefake.RootlessDynamicClient{}
	remoteServer, callCount := startCRDServer(t, "", http.StatusInternalServerError)
	ensurer := newTestCRDEnsurer(rootlessClient, true, remoteServer.URL)

	require.NoError(t, ensurer.run(context.Background(), true))
	require.Zero(t, *callCount, "expected server not to be called")
}

func TestEnsureCRD_AppliesWhenStoredNotFound(t *testing.T) {
	t.Parallel()
	rootlessClient := &kubefake.RootlessDynamicClient{}
	rootlessClient.ReturnGetErr = k8serrors.NewNotFound(crdGroupResource, testCRDName)
	remoteServer, _ := startCRDServer(t, minimalCRDYAML("v1beta2"), http.StatusOK)
	ensurer := newTestCRDEnsurer(rootlessClient, false, remoteServer.URL)

	require.NoError(t, ensurer.run(context.Background(), true))
	require.NotEmpty(t, rootlessClient.ApplyObjs, "expected apply to be called")
}

func TestEnsureCRD_SkipsWhenSpecsEqual(t *testing.T) {
	t.Parallel()
	storedCRD := minimalCRDUnstructuredWithVersion("v1beta2")
	rootlessClient := &kubefake.RootlessDynamicClient{}
	rootlessClient.ReturnGetObj = *storedCRD
	remoteServer, _ := startCRDServer(t, minimalCRDYAML("v1beta2"), http.StatusOK)
	ensurer := newTestCRDEnsurer(rootlessClient, false, remoteServer.URL)

	require.NoError(t, ensurer.run(context.Background(), true))
	require.Empty(t, rootlessClient.ApplyObjs, "did not expect apply when specs equal")
}

func TestEnsureCRD_RemoteFetchError(t *testing.T) {
	t.Parallel()
	rootlessClient := &kubefake.RootlessDynamicClient{
		ReturnGetErr: k8serrors.NewNotFound(crdGroupResource, testCRDName),
	}
	remoteServer, _ := startCRDServer(t, "error", http.StatusInternalServerError)
	ensurer := newTestCRDEnsurer(rootlessClient, false, remoteServer.URL)

	err := ensurer.run(context.Background(), true)
	require.NoError(t, err)
}

func Test_crdSpecDigest_ErrorsOnMissingSpec(t *testing.T) {
	t.Parallel()
	u := &unstructured.Unstructured{Object: map[string]any{"apiVersion": testCRDAPIVersion}}
	_, err := crdSpecDigest(u)
	require.Error(t, err, "expected error on missing spec")
}
