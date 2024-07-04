package managed

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_updateCR(t *testing.T) {
	t.Parallel()
	type args struct {
	}
	tests := []struct {
		name       string
		kymaCR     map[string]interface{}
		moduleName string
		want       map[string]interface{}
	}{
		{
			name: "unchanged modules list",
			kymaCR: map[string]interface{}{
				"keep": "unchanged",
				"spec": map[string]interface{}{
					"modules": []interface{}{
						map[string]interface{}{
							"name": "istio",
						},
					},
				},
			},
			moduleName: "module",
			want: map[string]interface{}{
				"keep": "unchanged",
				"spec": map[string]interface{}{
					"modules": []interface{}{
						map[string]interface{}{
							"name": "istio",
						},
					},
				},
			},
		},
		{
			name: "changed modules list",
			kymaCR: map[string]interface{}{
				"keep": "unchanged",
				"spec": map[string]interface{}{
					"modules": []interface{}{
						map[string]interface{}{
							"name": "istio",
						},
						map[string]interface{}{
							"name": "module",
						},
					},
				},
			},
			moduleName: "module",
			want: map[string]interface{}{
				"keep": "unchanged",
				"spec": map[string]interface{}{
					"modules": []interface{}{
						map[string]interface{}{
							"name": "istio",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		kymaCR := &unstructured.Unstructured{Object: tt.kymaCR}
		moduleName := tt.moduleName
		want := &unstructured.Unstructured{Object: tt.want}
		t.Run(tt.name, func(t *testing.T) {
			got := updateCR(kymaCR, moduleName)
			gotBytes, err := json.Marshal(got)
			require.NoError(t, err)
			wantBytes, err := json.Marshal(want)
			require.NoError(t, err)
			require.Equal(t, string(gotBytes), string(wantBytes))
		})
	}
}
