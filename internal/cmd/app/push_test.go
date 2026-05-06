package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_appPushConfig_complete_imageTag(t *testing.T) {
	tests := []struct {
		name    string
		imageTag string
		wantErr bool
	}{
		{
			name:     "empty tag skips validation - no error",
			imageTag: "",
			wantErr:  false,
		},
		{
			name:     "valid commit sha",
			imageTag: "abc1234def5678",
			wantErr:  false,
		},
		{
			name:     "valid semver",
			imageTag: "1.0.0",
			wantErr:  false,
		},
		{
			name:     "valid underscore prefix",
			imageTag: "_build",
			wantErr:  false,
		},
		{
			name:     "valid single char",
			imageTag: "1",
			wantErr:  false,
		},
		{
			name:     "valid with dots and dashes",
			imageTag: "v1.2.3-beta.1",
			wantErr:  false,
		},
		{
			name:     "invalid: contains space",
			imageTag: "my tag",
			wantErr:  true,
		},
		{
			name:     "invalid: contains colon",
			imageTag: "my:tag",
			wantErr:  true,
		},
		{
			name:     "invalid: contains slash",
			imageTag: "my/tag",
			wantErr:  true,
		},
		{
			name:     "invalid: contains @",
			imageTag: "my@tag",
			wantErr:  true,
		},
		{
			name:     "invalid: starts with dot",
			imageTag: ".mytag",
			wantErr:  true,
		},
		{
			name:     "invalid: starts with dash",
			imageTag: "-mytag",
			wantErr:  true,
		},
		{
			name:     "invalid: exceeds 128 chars",
			imageTag: "abbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &appPushConfig{
				imageTag: tt.imageTag,
			}
			err := cfg.complete()
			if tt.wantErr {
				require.NotNil(t, err, "expected validation error for imageTag=%q", tt.imageTag)
			} else {
				require.Nil(t, err, "expected no error for imageTag=%q", tt.imageTag)
			}
		})
	}
}

func Test_resolveImageTag(t *testing.T) {
	t.Run("provided tag is used in image name", func(t *testing.T) {
		imageTag := "abc1234"
		resolvedTag := resolveImageTag(imageTag)
		require.Equal(t, "abc1234", resolvedTag)
	})

	t.Run("empty tag resolves to timestamp format", func(t *testing.T) {
		resolvedTag := resolveImageTag("")
		require.Regexp(t, `^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}$`, resolvedTag)
	})
}
