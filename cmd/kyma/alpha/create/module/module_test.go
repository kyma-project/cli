package module

import (
	_ "embed"
	"reflect"
	"testing"

	"github.com/kyma-project/lifecycle-manager/api/shared"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/pkg/module"
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

func Test_command_getModuleTemplateLabels(t *testing.T) {
	type fields struct {
		Command cli.Command
		opts    *Options
	}
	type args struct {
		modCnf *Config
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]string
	}{
		{
			name: "beta module with moduleConfig labels set",
			fields: fields{
				opts: &Options{},
			},
			args: args{
				modCnf: &Config{
					Beta: true,
					Labels: map[string]string{
						"label1": "value1",
						"label2": "value2",
					},
					Version: "1.1.1",
				},
			},
			want: map[string]string{
				"label1":         "value1",
				"label2":         "value2",
				shared.BetaLabel: shared.EnableLabelValue,
			},
		},
		{
			name: "internal module",
			fields: fields{
				opts: &Options{},
			},
			args: args{
				modCnf: &Config{
					Internal: true,
					Version:  "1.1.1",
				},
			},
			want: map[string]string{
				shared.InternalLabel: shared.EnableLabelValue,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command{
				Command: tt.fields.Command,
				opts:    tt.fields.opts,
			}
			if got := cmd.getModuleTemplateLabels(tt.args.modCnf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getModuleTemplateLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_command_getModuleTemplateAnnotations(t *testing.T) {
	type fields struct {
		Command cli.Command
		opts    *Options
	}
	type args struct {
		modCnf      *Config
		crValidator validator
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]string
	}{
		{
			name: "module with moduleConfig annotations set",
			fields: fields{
				opts: &Options{},
			},
			args: args{
				modCnf: &Config{
					Internal: true,
					Annotations: map[string]string{
						"annotation1": "value1",
						"annotation2": "value2",
					},
					Version: "1.1.1",
				},
				crValidator: &module.SingleManifestFileCRValidator{
					Crd: namespacedScopedCrd,
				},
			},
			want: map[string]string{
				"annotation1":                    "value1",
				"annotation2":                    "value2",
				shared.ModuleVersionAnnotation:   "1.1.1",
				shared.IsClusterScopedAnnotation: shared.DisableLabelValue,
			},
		},
		{
			name: "cluster scoped module with moduleConfig annotations set",
			fields: fields{
				opts: &Options{},
			},
			args: args{
				modCnf: &Config{
					Annotations: map[string]string{
						"annotation1": "value1",
						"annotation2": "value2",
					},
					Version: "1.1.1",
				},
				crValidator: &module.SingleManifestFileCRValidator{
					Crd: clusterScopedCrd,
				},
			},
			want: map[string]string{
				"annotation1":                    "value1",
				"annotation2":                    "value2",
				shared.IsClusterScopedAnnotation: shared.EnableLabelValue,
				shared.ModuleVersionAnnotation:   "1.1.1",
			},
		},
		{
			name: "module versions set from version flag",
			fields: fields{
				opts: &Options{Version: "1.0.0"},
			},
			args: args{
				crValidator: &module.SingleManifestFileCRValidator{
					Crd: namespacedScopedCrd,
				},
			},
			want: map[string]string{
				shared.ModuleVersionAnnotation:   "1.0.0",
				shared.IsClusterScopedAnnotation: shared.DisableLabelValue,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command{
				Command: tt.fields.Command,
				opts:    tt.fields.opts,
			}
			if got := cmd.getModuleTemplateAnnotations(tt.args.modCnf, tt.args.crValidator); !reflect.DeepEqual(got,
				tt.want) {
				t.Errorf("getModuleTemplateAnnotations() = %v, want %v", got, tt.want)
			}
		})
	}
}
