package managed

import (
	"encoding/json"
	"reflect"
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
		channel    string
		want       map[string]interface{}
	}{
		{
			name:       "unchanged modules list",
			moduleName: "module",
			channel:    "",
			kymaCR: map[string]interface{}{
				"keep": "unchanged",
				"spec": map[string]interface{}{
					"modules": []interface{}{
						map[string]interface{}{
							"name": "module",
						},
					},
				},
			},
			want: map[string]interface{}{
				"keep": "unchanged",
				"spec": map[string]interface{}{
					"modules": []interface{}{
						map[string]interface{}{
							"name": "module",
						},
					},
				},
			},
		},
		{
			name:       "added module",
			moduleName: "module",
			channel:    "",
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
			want: map[string]interface{}{
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
		},
		{
			name:       "added module with channel",
			moduleName: "module",
			channel:    "channel",
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
			want: map[string]interface{}{
				"keep": "unchanged",
				"spec": map[string]interface{}{
					"modules": []interface{}{
						map[string]interface{}{
							"name": "istio",
						},
						map[string]interface{}{
							"name":    "module",
							"channel": "channel",
						},
					},
				},
			},
		},
		{
			name:       "added channel to existing module",
			moduleName: "module",
			channel:    "channel",
			kymaCR: map[string]interface{}{
				"keep": "unchanged",
				"spec": map[string]interface{}{
					"modules": []interface{}{
						map[string]interface{}{
							"name": "module",
						},
					},
				},
			},
			want: map[string]interface{}{
				"keep": "unchanged",
				"spec": map[string]interface{}{
					"modules": []interface{}{
						map[string]interface{}{
							"name":    "module",
							"channel": "channel",
						},
					},
				},
			},
		},
		{
			name:       "removed channel from existing module",
			moduleName: "module",
			channel:    "",
			kymaCR: map[string]interface{}{
				"keep": "unchanged",
				"spec": map[string]interface{}{
					"modules": []interface{}{
						map[string]interface{}{
							"name":    "module",
							"channel": "channel",
						},
					},
				},
			},
			want: map[string]interface{}{
				"keep": "unchanged",
				"spec": map[string]interface{}{
					"modules": []interface{}{
						map[string]interface{}{
							"name": "module",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		kymaCR := &unstructured.Unstructured{Object: tt.kymaCR}
		moduleName := tt.moduleName
		moduleChannel := tt.channel
		want := &unstructured.Unstructured{Object: tt.want}
		t.Run(tt.name, func(t *testing.T) {
			got := updateCR(kymaCR, moduleName, moduleChannel)
			gotBytes, err := json.Marshal(got)
			require.NoError(t, err)
			wantBytes, err := json.Marshal(want)
			require.NoError(t, err)
			var gotInterface map[string]interface{}
			var wantInterface map[string]interface{}
			err = json.Unmarshal(gotBytes, &gotInterface)
			require.NoError(t, err)
			err = json.Unmarshal(wantBytes, &wantInterface)
			require.NoError(t, err)
			if !reflect.DeepEqual(gotInterface, wantInterface) {
				t.Errorf("updateCR() = %v, want %v", gotInterface, wantInterface)
			}
		})
	}
}
