package module_test

import (
	"fmt"
	"testing"

	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	ocmv1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-project/cli/pkg/module"
	"github.com/kyma-project/cli/pkg/module/mocks"
)

func TestNoSchemaURL(t *testing.T) {
	tests := []struct {
		name         string
		testRegistry string
		want         string
	}{
		{
			name:         "https registry",
			testRegistry: "https://registry.domain",
			want:         "registry.domain",
		},
		{
			name:         "http registry",
			testRegistry: "http://registry.domain",
			want:         "registry.domain",
		},
		{
			name:         "no scheme registry",
			testRegistry: "registry.domain",
			want:         "registry.domain",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := module.NoSchemeURL(tt.testRegistry); got != tt.want {
					t.Errorf("BuildNewOCIRegistryRepository() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestRemote_Push(t *testing.T) {
	archiveFS := memoryfs.New()
	cd := &compdesc.ComponentDescriptor{}
	cd.ComponentSpec.SetName("test.io/module/test")
	cd.ComponentSpec.SetVersion("1.0.0")
	cd.Metadata.ConfiguredVersion = "v2"
	builtByCLI, _ := ocmv1.NewLabel("kyma-project.io/built-by", "cli", ocmv1.WithVersion("v1"))
	cd.Provider = ocmv1.Provider{Name: "kyma-project.io", Labels: ocmv1.Labels{*builtByCLI}}
	compdesc.DefaultResources(cd)

	archive, _ := module.CreateArchive(archiveFS, "./mod", cd)
	ociRepoAccessMock := mocks.OciRepoAccess{}
	ociRepoAccessMock.On("PushComponentVersion", archive, mock.Anything, mock.Anything).Return(nil)
	ociRepoAccessMock.On("GetComponentVersion", archive,
		mock.Anything).Return(mock.AnythingOfType("internal.ComponentVersionAccess"))

	type fields struct {
		OciRepoAccess module.OciRepoAccess
	}
	type args struct {
		archive   *comparch.ComponentArchive
		overwrite bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ocm.ComponentVersionAccess
		want1   bool
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Same version, same content, with overwrite flag",
			args: args{
				overwrite: true,
				archive:   archive,
			},
			fields: fields{
				OciRepoAccess: &ociRepoAccessMock,
			},
		},
		{
			name: "Same version, same content, without overwrite flag",
		},
		{
			name: "Same version, different content, with overwrite flag",
		},
		{
			name: "Same version, different content, without overwrite flag",
		},
		{
			name: "different version, with overwrite flag",
		},
		{
			name: "different version, without overwrite flag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &module.Remote{
				Insecure:      true,
				OciRepoAccess: tt.fields.OciRepoAccess,
			}
			got, got1, err := r.Push(tt.args.archive, tt.args.overwrite)
			if !tt.wantErr(t, err, fmt.Sprintf("Push(%v, %v)", tt.args.archive, tt.args.overwrite)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Push(%v, %v)", tt.args.archive, tt.args.overwrite)
			assert.Equalf(t, tt.want1, got1, "Push(%v, %v)", tt.args.archive, tt.args.overwrite)
		})
	}
}
