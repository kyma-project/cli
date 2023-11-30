package module_test

import (
	"fmt"
	"testing"

	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	ocmv1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
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

func TestRemote_ShouldPushArchive(t *testing.T) {
	archiveFS := memoryfs.New()
	cd := &compdesc.ComponentDescriptor{}
	cd.ComponentSpec.SetName("test.io/module/test")
	cd.ComponentSpec.SetVersion("1.0.0")
	cd.Metadata.ConfiguredVersion = "v2"
	builtByCLI, _ := ocmv1.NewLabel("kyma-project.io/built-by", "cli", ocmv1.WithVersion("v1"))
	cd.Provider = ocmv1.Provider{Name: "kyma-project.io", Labels: ocmv1.Labels{*builtByCLI}}
	compdesc.DefaultResources(cd)

	archive, _ := module.CreateArchive(archiveFS, "./mod", cd)
	var ociRepoAccessMock mocks.OciRepoAccess

	type args struct {
		archive   *comparch.ComponentArchive
		overwrite bool
	}
	tests := []struct {
		name                 string
		versionAlreadyExists bool
		sameContent          bool
		args                 args
		wantIsPushed         bool
		assertFn             func(err error, i ...interface{})
	}{
		{
			name: "Same version, same content, with overwrite flag",
			args: args{
				overwrite: true,
				archive:   archive,
			},
			assertFn: func(err error, i ...interface{}) {
				assert.NoError(t, err)
				ociRepoAccessMock.AssertNumberOfCalls(t, "ComponentVersionExists", 0)
				ociRepoAccessMock.AssertNumberOfCalls(t, "GetComponentVersion", 0)
				ociRepoAccessMock.AssertNumberOfCalls(t, "DescriptorResourcesAreEquivalent", 0)
			},
			wantIsPushed:         true,
			versionAlreadyExists: false,
			sameContent:          true,
		},
		{
			name: "Same version, same content, without overwrite flag",
			args: args{
				overwrite: false,
				archive:   archive,
			},
			assertFn: func(err error, i ...interface{}) {
				assert.NoError(t, err)
				ociRepoAccessMock.AssertNumberOfCalls(t, "ComponentVersionExists", 1)
				ociRepoAccessMock.AssertNumberOfCalls(t, "GetComponentVersion", 1)
				ociRepoAccessMock.AssertNumberOfCalls(t, "DescriptorResourcesAreEquivalent", 1)
			},
			wantIsPushed:         false,
			versionAlreadyExists: true,
			sameContent:          true,
		},
		{
			name: "Same version, different content, with overwrite flag",
			args: args{
				overwrite: true,
				archive:   archive,
			},
			assertFn: func(err error, i ...interface{}) {
				assert.NoError(t, err)
				ociRepoAccessMock.AssertNumberOfCalls(t, "ComponentVersionExists", 0)
				ociRepoAccessMock.AssertNumberOfCalls(t, "GetComponentVersion", 0)
				ociRepoAccessMock.AssertNumberOfCalls(t, "DescriptorResourcesAreEquivalent", 0)
			},
			wantIsPushed:         true,
			versionAlreadyExists: true,
			sameContent:          false,
		},
		{
			name: "Same version, different content, without overwrite flag",
			args: args{
				overwrite: false,
				archive:   archive,
			},
			assertFn: func(err error, i ...interface{}) {
				assert.Errorf(t, err,
					"version 1.0.0 already exists with different content, please use --module-archive-version-overwrite flag to overwrite it")
				ociRepoAccessMock.AssertNumberOfCalls(t, "ComponentVersionExists", 1)
				ociRepoAccessMock.AssertNumberOfCalls(t, "GetComponentVersion", 1)
				ociRepoAccessMock.AssertNumberOfCalls(t, "DescriptorResourcesAreEquivalent", 1)
			},
			wantIsPushed:         false,
			versionAlreadyExists: true,
			sameContent:          false,
		},
		{
			name: "different version, with overwrite flag",
			args: args{
				overwrite: true,
				archive:   archive,
			},
			assertFn: func(err error, i ...interface{}) {
				assert.NoError(t, err)
				ociRepoAccessMock.AssertNumberOfCalls(t, "ComponentVersionExists", 0)
				ociRepoAccessMock.AssertNumberOfCalls(t, "GetComponentVersion", 0)
				ociRepoAccessMock.AssertNumberOfCalls(t, "DescriptorResourcesAreEquivalent", 0)
			},
			wantIsPushed:         true,
			versionAlreadyExists: false,
			sameContent:          false,
		},
		{
			name: "different version, without overwrite flag",
			args: args{
				overwrite: false,
				archive:   archive,
			},
			assertFn: func(err error, i ...interface{}) {
				assert.NoError(t, err)
				ociRepoAccessMock.AssertNumberOfCalls(t, "ComponentVersionExists", 1)
				ociRepoAccessMock.AssertNumberOfCalls(t, "GetComponentVersion", 0)
				ociRepoAccessMock.AssertNumberOfCalls(t, "DescriptorResourcesAreEquivalent", 0)
			},
			wantIsPushed:         true,
			versionAlreadyExists: false,
			sameContent:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ociRepoAccessMock = mocks.OciRepoAccess{}
			ociRepoAccessMock.On("PushComponentVersion", archive, mock.Anything, mock.Anything).Return(nil)
			ociRepoAccessMock.On("GetComponentVersion", archive,
				mock.Anything).Return(nil, nil)
			ociRepoAccessMock.On("ComponentVersionExists", archive, mock.Anything).Return(tt.versionAlreadyExists, nil)
			ociRepoAccessMock.On("DescriptorResourcesAreEquivalent", mock.Anything,
				mock.Anything).Return(tt.sameContent)
			r := &module.Remote{
				Insecure:      true,
				OciRepoAccess: &ociRepoAccessMock,
			}
			repo, _ := r.GetRepository(cpi.DefaultContext())
			got, err := r.ShouldPushArchive(repo, tt.args.archive, tt.args.overwrite)
			tt.assertFn(err, fmt.Sprintf("Push(%v, %v)", tt.args.archive, tt.args.overwrite))

			assert.Equalf(t, tt.wantIsPushed, got, "Push(%v, %v)", tt.args.archive, tt.args.overwrite)
		})
	}
}
