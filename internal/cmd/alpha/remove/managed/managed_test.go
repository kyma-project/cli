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
	type args struct {
	}
	tests := []struct {
		name       string
		kymaCR     *kyma.Kyma
		moduleName string
		want       *kyma.Kyma
	}{
		{
			name: "unchanged modules list",
			kymaCR: &kyma.Kyma{
				Spec: kyma.KymaSpec{
					Modules: []kyma.Module{
						{
							Name: "istio",
						},
					},
				},
			},
			moduleName: "module",
			want: &kyma.Kyma{
				Spec: kyma.KymaSpec{
					Modules: []kyma.Module{
						{
							Name: "istio",
						},
					},
				},
			},
		},
		{
			name: "changed modules list",
			kymaCR: &kyma.Kyma{
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
			moduleName: "module",
			want: &kyma.Kyma{
				Spec: kyma.KymaSpec{
					Modules: []kyma.Module{
						{
							Name: "istio",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		kymaCR := tt.kymaCR
		moduleName := tt.moduleName
		want := tt.want
		t.Run(tt.name, func(t *testing.T) {
			got := disableModule(kymaCR, moduleName)
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
