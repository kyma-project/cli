package module

import (
	_ "embed"
	"testing"
)

//go:embed testdata/clusterScopedCRD.yaml
var clusterScopedCrd []byte

//go:embed testdata/namespacedScopedCRD.yaml
var namespacedScopedCrd []byte

func Test_isCrdClusterScoped(t *testing.T) {
	type args struct {
		crdBytes []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Cluster scoped module",
			args: args{
				crdBytes: clusterScopedCrd,
			},
			want: true,
		},
		{
			name: "Namespaced scoped module",
			args: args{
				crdBytes: namespacedScopedCrd,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCrdClusterScoped(tt.args.crdBytes); got != tt.want {
				t.Errorf("isCrdClusterScoped() = %v, want %v", got, tt.want)
			}
		})
	}
}
