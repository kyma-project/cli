package module

import (
	"bytes"
	"reflect"
	"testing"
	"text/template"

	"github.com/kyma-project/lifecycle-manager/api/v1beta2"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

var accessVersion ocm.ComponentVersionAccess

var noCustomStateCheck []v1beta2.CustomStateCheck
var defaultCustomStateCheck = []v1beta2.CustomStateCheck{
	{
		JSONPath:    "status.health",
		Value:       "green",
		MappedState: "Ready",
	},
}

func TestTemplate(t *testing.T) {
	accessVersion = createOcmComponentVersionAccess(t)
	type args struct {
		remote             ocm.ComponentVersionAccess
		moduleTemplateName string
		namespace          string
		channel            string
		data               []byte
		labels             map[string]string
		annotations        map[string]string
		checks             []v1beta2.CustomStateCheck
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "ModuleTemplate with default values",
			args: args{
				remote:      accessVersion,
				channel:     "regular",
				labels:      map[string]string{},
				annotations: map[string]string{},
				checks:      noCustomStateCheck,
			},
			want: getExpectedModuleTemplate(t, "",
				map[string]string{
					"operator.kyma-project.io/module-name": "template-operator"}, map[string]string{},
				noCustomStateCheck),
			wantErr: false,
		},
		{
			name: "ModuleTemplate with custom namespace",
			args: args{
				remote:      accessVersion,
				namespace:   "kyma-system",
				channel:     "regular",
				labels:      map[string]string{},
				annotations: map[string]string{},
				checks:      noCustomStateCheck,
			},
			want: getExpectedModuleTemplate(t, "kyma-system",
				map[string]string{
					"operator.kyma-project.io/module-name": "template-operator"}, map[string]string{},
				noCustomStateCheck),
			wantErr: false,
		},
		{
			name: "ModuleTemplate with extra labels",
			args: args{
				remote:      accessVersion,
				namespace:   "kyma-system",
				channel:     "regular",
				labels:      map[string]string{"is-custom-label": "true"},
				annotations: map[string]string{},
				checks:      noCustomStateCheck,
			},
			want: getExpectedModuleTemplate(t, "kyma-system",
				map[string]string{
					"operator.kyma-project.io/module-name": "template-operator", "is-custom-label": "true"},
				map[string]string{}, noCustomStateCheck),
			wantErr: false,
		},
		{
			name: "ModuleTemplate with extra annotations",
			args: args{
				remote:      accessVersion,
				namespace:   "kyma-system",
				channel:     "regular",
				labels:      map[string]string{},
				annotations: map[string]string{"is-custom-annotation": "true"},
				checks:      noCustomStateCheck,
			},
			want: getExpectedModuleTemplate(t, "kyma-system",
				map[string]string{
					"operator.kyma-project.io/module-name": "template-operator"},
				map[string]string{"is-custom-annotation": "true"}, []v1beta2.CustomStateCheck{}),
			wantErr: false,
		},
		{
			name: "ModuleTemplate with Custom State Check",
			args: args{
				remote:      accessVersion,
				channel:     "regular",
				labels:      map[string]string{},
				annotations: map[string]string{},
				checks:      defaultCustomStateCheck,
			},
			want: getExpectedModuleTemplate(t, "",
				map[string]string{"operator.kyma-project.io/module-name": "template-operator"}, map[string]string{},
				defaultCustomStateCheck),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Template(tt.args.remote, tt.args.moduleTemplateName, tt.args.namespace, tt.args.channel,
				tt.args.data, tt.args.labels, tt.args.annotations, tt.args.checks)
			if (err != nil) != tt.wantErr {
				t.Errorf("Template() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Template() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func createOcmComponentVersionAccess(t *testing.T) ocm.ComponentVersionAccess {
	tempFs, err := osfs.NewTempFileSystem()
	assert.Equal(t, nil, err)
	ociSpec, err := ctf.NewRepositorySpec(accessobj.ACC_CREATE, "test", accessio.PathFileSystem(tempFs), accessobj.FormatDirectory)
	assert.Equal(t, nil, err)
	repo, err := ocm.New().RepositoryForSpec(ociSpec)
	assert.Equal(t, nil, err)
	comp, err := repo.LookupComponent("github.com/kyma-project/template-operator")
	assert.Equal(t, nil, err)
	version, err := comp.NewVersion("v1")
	assert.Equal(t, nil, err)

	return version
}

func getExpectedModuleTemplate(t *testing.T,
	namespace string, labels map[string]string,
	annotations map[string]string, checks []v1beta2.CustomStateCheck) []byte {
	cva, err := compdesc.Convert(accessVersion.GetDescriptor())
	assert.Equal(t, nil, err)
	temp, err := template.New("modTemplate").Funcs(template.FuncMap{"yaml": yaml.Marshal, "indent": Indent}).Parse(modTemplate)
	assert.Equal(t, nil, err)
	td := moduleTemplateData{
		ResourceName:      "template-operator-regular",
		Namespace:         namespace,
		Channel:           "regular",
		Annotations:       annotations,
		Labels:            labels,
		Data:              "",
		Descriptor:        cva,
		CustomStateChecks: checks,
	}
	w := &bytes.Buffer{}
	err = temp.Execute(w, td)
	assert.Equal(t, nil, err)
	return w.Bytes()
}
