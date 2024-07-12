package managed

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/stretchr/testify/require"
)

func Test_updateCR(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		kymaCR     *kyma.Kyma
		moduleName string
		channel    string
		want       *kyma.Kyma
	}{
		{
			name:       "unchanged modules list",
			moduleName: "module",
			channel:    "",
			kymaCR: &kyma.Kyma{
				Spec: kyma.KymaSpec{
					Modules: []kyma.Module{
						{
							Name: "module",
						},
					},
				},
			},
			want: &kyma.Kyma{
				Spec: kyma.KymaSpec{
					Modules: []kyma.Module{
						{
							Name: "module",
						},
					},
				},
			},
		},
		{
			name:       "added module",
			moduleName: "module",
			channel:    "",
			kymaCR: &kyma.Kyma{
				Spec: kyma.KymaSpec{
					Modules: []kyma.Module{
						{
							Name: "istio",
						},
					},
				},
			},
			want: &kyma.Kyma{
				Spec: kyma.KymaSpec{
					Modules: []kyma.Module{
						{
							Name: "istio",
						},
						{
							Name: "module",
						},
					},
				},
			},
		},
		{
			name:       "added module with channel",
			moduleName: "module",
			channel:    "channel",
			kymaCR: &kyma.Kyma{
				Spec: kyma.KymaSpec{
					Modules: []kyma.Module{
						{
							Name: "istio",
						},
					},
				},
			},
			want: &kyma.Kyma{
				Spec: kyma.KymaSpec{
					Modules: []kyma.Module{
						{
							Name: "istio",
						},
						{
							Name:    "module",
							Channel: "channel",
						},
					},
				},
			},
		},
		{
			name:       "added channel to existing module",
			moduleName: "module",
			channel:    "channel",
			kymaCR: &kyma.Kyma{
				Spec: kyma.KymaSpec{
					Modules: []kyma.Module{
						{
							Name: "module",
						},
					},
				},
			},
			want: &kyma.Kyma{
				Spec: kyma.KymaSpec{
					Modules: []kyma.Module{
						{
							Name:    "module",
							Channel: "channel",
						},
					},
				},
			},
		},
		{
			name:       "removed channel from existing module",
			moduleName: "module",
			channel:    "",
			kymaCR: &kyma.Kyma{
				Spec: kyma.KymaSpec{
					Modules: []kyma.Module{
						{
							Name:    "module",
							Channel: "channel",
						},
					},
				},
			},
			want: &kyma.Kyma{
				Spec: kyma.KymaSpec{
					Modules: []kyma.Module{
						{
							Name: "module",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		kymaCR := tt.kymaCR
		moduleName := tt.moduleName
		moduleChannel := tt.channel
		want := tt.want
		t.Run(tt.name, func(t *testing.T) {
			got := enableModule(kymaCR, moduleName, moduleChannel)
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
