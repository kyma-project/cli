package module_test

import (
	"testing"

	"github.com/kyma-project/cli/pkg/module"
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
