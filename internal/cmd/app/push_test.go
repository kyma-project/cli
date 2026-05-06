package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_appPushConfig_complete_imageTag(t *testing.T) {
	tests := []struct {
		name      string
		imageTag  string
		wantError bool
	}{
		{
			name:      "empty tag uses timestamp fallback - no error",
			imageTag:  "",
			wantError: false,
		},
		{
			name:      "valid commit sha",
			imageTag:  "abc1234def5678",
			wantError: false,
		},
		{
			name:      "valid semver",
			imageTag:  "1.0.0",
			wantError: false,
		},
		{
			name:      "valid underscore prefix",
			imageTag:  "_build",
			wantError: false,
		},
		{
			name:      "valid single char",
			imageTag:  "1",
			wantError: false,
		},
		{
			name:      "valid with dots and dashes",
			imageTag:  "v1.2.3-beta.1",
			wantError: false,
		},
		{
			name:      "invalid: contains space",
			imageTag:  "my tag",
			wantError: true,
		},
		{
			name:      "invalid: contains colon",
			imageTag:  "my:tag",
			wantError: true,
		},
		{
			name:      "invalid: contains slash",
			imageTag:  "my/tag",
			wantError: true,
		},
		{
			name:      "invalid: contains @",
			imageTag:  "my@tag",
			wantError: true,
		},
		{
			name:      "invalid: starts with dot",
			imageTag:  ".mytag",
			wantError: true,
		},
		{
			name:      "invalid: starts with dash",
			imageTag:  "-mytag",
			wantError: true,
		},
		{
			name:      "invalid: exceeds 128 chars",
			imageTag:  "a123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &appPushConfig{
				imageTag: tt.imageTag,
			}
			err := cfg.complete()
			if tt.wantError {
				require.NotNil(t, err, "expected validation error for imageTag=%q", tt.imageTag)
			} else {
				require.Nil(t, err, "expected no error for imageTag=%q", tt.imageTag)
			}
		})
	}
}
