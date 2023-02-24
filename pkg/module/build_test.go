package module_test

import (
	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/kyma-project/cli/pkg/module"
	"reflect"
	"testing"
)

func TestBuildNewOCIRegistryRepository(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {

			if got := module.BuildNewOCIRegistryRepository(tt.testRegistry, cdv2.OCIRegistryURLPathMapping); !reflect.DeepEqual(got, &cdv2.OCIRegistryRepository{
				ObjectType: cdv2.ObjectType{
					Type: cdv2.OCIRegistryType,
				},
				BaseURL:              tt.want,
				ComponentNameMapping: cdv2.OCIRegistryURLPathMapping,
			}) {
				t.Errorf("BuildNewOCIRegistryRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}
