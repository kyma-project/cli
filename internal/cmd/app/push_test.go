package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_appPushConfig_complete_buildTag(t *testing.T) {
	tests := []struct {
		name     string
		buildTag string
		wantErr  bool
	}{
		{
			name:     "empty tag skips validation - no error",
			buildTag: "",
			wantErr:  false,
		},
		{
			name:     "valid commit sha",
			buildTag: "abc1234def5678",
			wantErr:  false,
		},
		{
			name:     "valid semver",
			buildTag: "1.0.0",
			wantErr:  false,
		},
		{
			name:     "valid underscore prefix",
			buildTag: "_build",
			wantErr:  false,
		},
		{
			name:     "valid single char",
			buildTag: "1",
			wantErr:  false,
		},
		{
			name:     "valid with dots and dashes",
			buildTag: "v1.2.3-beta.1",
			wantErr:  false,
		},
		{
			name:     "invalid: contains space",
			buildTag: "my tag",
			wantErr:  true,
		},
		{
			name:     "invalid: contains colon",
			buildTag: "my:tag",
			wantErr:  true,
		},
		{
			name:     "invalid: contains slash",
			buildTag: "my/tag",
			wantErr:  true,
		},
		{
			name:     "invalid: contains @",
			buildTag: "my@tag",
			wantErr:  true,
		},
		{
			name:     "invalid: starts with dot",
			buildTag: ".mytag",
			wantErr:  true,
		},
		{
			name:     "invalid: starts with dash",
			buildTag: "-mytag",
			wantErr:  true,
		},
		{
			name:     "invalid: exceeds 128 chars",
			buildTag: "abbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &appPushConfig{
				buildTag: tt.buildTag,
			}
			err := cfg.validate()
			if tt.wantErr {
				require.NotNil(t, err, "expected validation error for buildTag=%q", tt.buildTag)
			} else {
				require.Nil(t, err, "expected no error for buildTag=%q", tt.buildTag)
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
