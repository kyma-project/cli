package repo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	communityModuleJSON = `[
	{
    	"apiVersion": "operator.kyma-project.io/v1beta2",
    	"kind": "ModuleTemplate",
    	"metadata": {
      		"name": "test-module-2",
      		"labels": {
        		"operator.kyma-project.io/module-name": "test-module"
      		}
    	},
    	"spec": {
			"moduleName": "test-module",
			"version": "0.0.2",
			"manager": {
				"name": "dockerregistry-operator",
				"namespace": "kyma-system",
				"group": "apps",
				"version": "v1",
				"kind": "Deployment"
			},
			"resources": [],
			"descriptor": {}
		}
	}
]
`
)

func TestModuleTemplateRemoteRepo_Community(t *testing.T) {
	validTemplates := []kyma.ModuleTemplate{
		{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "operator.kyma-project.io/v1beta2",
				Kind:       "ModuleTemplate",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-module-2",
				Namespace: "test-module-namespace",
			},
			Spec: kyma.ModuleTemplateSpec{
				ModuleName: "test-module",
				Version:    "0.0.2",
			},
		}}

	tests := []struct {
		name         string
		serverFunc   http.HandlerFunc
		expectErr    bool
		expectResult []kyma.ModuleTemplate
	}{
		{
			name: "success",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(communityModuleJSON))
			},
			expectErr:    false,
			expectResult: validTemplates,
		},
		{
			name: "server error",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("error"))
			},
			expectErr: true,
		},
		{
			name: "invalid json",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("not json"))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			ts := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer ts.Close()

			repo := newModuleTemplatesRemoteRepoWithURL(ts.URL)
			result, err := repo.Community()

			if tt.expectErr {
				require.Error(err, "expected error, got nil")
			} else {
				require.NoError(err, "unexpected error: %v", err)
				require.Equal(len(tt.expectResult), len(result), "unexpected number of modules")

				require.Equal(tt.expectResult[0].Spec.ModuleName, result[0].Spec.ModuleName, "module name mismatch")
				require.Equal(tt.expectResult[0].Spec.Version, result[0].Spec.Version, "module version mismatch")
			}
		})
	}
}
