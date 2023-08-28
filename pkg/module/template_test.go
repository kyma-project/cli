package module

import (
	"bytes"
	"reflect"
	"testing"
	"text/template"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"gotest.tools/assert"
	"sigs.k8s.io/yaml"
)

func TestTemplate(t *testing.T) {
	accessVersion := createOcmComponentVersionAccess(t)
	type args struct {
		remote             ocm.ComponentVersionAccess
		moduleTemplateName string
		namespace          string
		channel            string
		data               []byte
		labels             map[string]string
		annotations        map[string]string
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
			},
			want: getExpectedModuleTemplate(t, "",
				"regular", map[string]string{
					"operator.kyma-project.io/module-name": "template-operator"}, map[string]string{}, accessVersion),
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
			},
			want: getExpectedModuleTemplate(t, "kyma-system",
				"regular", map[string]string{
					"operator.kyma-project.io/module-name": "template-operator"}, map[string]string{}, accessVersion),
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
			},
			want: getExpectedModuleTemplate(t, "kyma-system",
				"regular", map[string]string{
					"operator.kyma-project.io/module-name": "template-operator", "is-custom-label": "true"},
				map[string]string{}, accessVersion),
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
			},
			want: getExpectedModuleTemplate(t, "kyma-system",
				"regular", map[string]string{
					"operator.kyma-project.io/module-name": "template-operator"},
				map[string]string{"is-custom-annotation": "true"}, accessVersion),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Template(tt.args.remote, tt.args.moduleTemplateName, tt.args.namespace, tt.args.channel, tt.args.data, tt.args.labels, tt.args.annotations)
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
	namespace string, channel string, labels map[string]string,
	annotations map[string]string, version ocm.ComponentVersionAccess) []byte {
	cva, _ := compdesc.Convert(version.GetDescriptor())
	temp, _ := template.New("modTemplate").Funcs(template.FuncMap{"yaml": yaml.Marshal, "indent": Indent}).Parse(modTemplate)
	td := moduleTemplateData{
		ResourceName: "template-operator-regular",
		Namespace:    namespace,
		Channel:      channel,
		Annotations:  annotations,
		Labels:       labels,
		Data:         "",
		Descriptor:   cva,
	}
	w := &bytes.Buffer{}
	err := temp.Execute(w, td)
	assert.Equal(t, nil, err)
	return w.Bytes()
}
