package sourced_test

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/types/sourced"
	"github.com/stretchr/testify/require"
)

func TestParseEnv(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    sourced.Env
		wantErr error
	}{
		{
			name:    "single key from resource",
			value:   "name=NAME,path=PATH,key=KEY",
			want:    sourced.Env{Name: "NAME", Location: "PATH", LocationKey: "KEY"},
			wantErr: nil,
		},
		{
			name:    "single key from resource - shorthand",
			value:   "NAME=PATH:KEY",
			want:    sourced.Env{Name: "NAME", Location: "PATH", LocationKey: "KEY"},
			wantErr: nil,
		},
		{
			name:    "multi keys from resource",
			value:   "path=PATH",
			want:    sourced.Env{Location: "PATH"},
			wantErr: nil,
		},
		{
			name:    "multi keys from resource - shorthand",
			value:   "PATH",
			want:    sourced.Env{Location: "PATH"},
			wantErr: nil,
		},
		{
			name:    "multi keys from resource with prefix",
			value:   "resource=PATH,prefix=PREFIX_",
			want:    sourced.Env{Location: "PATH", LocationKeysPrefix: "PREFIX_"},
			wantErr: nil,
		},
		{
			name:    "multi keys from resource with prefix - shorthand",
			value:   "path=PATH,prefix=PREFIX_",
			want:    sourced.Env{Location: "PATH", LocationKeysPrefix: "PREFIX_"},
			wantErr: nil,
		},
		{
			name:    "invalid format",
			value:   "name=,key=KEY",
			want:    sourced.Env{},
			wantErr: sourced.ErrInvalidEnvFormat,
		},
		{
			name:    "unknown field",
			value:   "name=NAME,path=PATH,unknown=VALUE",
			want:    sourced.Env{},
			wantErr: sourced.ErrUnknownEnvField,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := sourced.ParseEnv(tt.value)
			require.Equal(t, tt.wantErr, gotErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestEnv_String(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		input string
		want  string
	}{
		{
			name:  "single key from resource",
			input: "name=NAME,path=PATH,key=KEY",
			want:  "NAME=PATH:KEY",
		},
		{
			name:  "single key from resource - shorthand",
			input: "NAME=PATH:KEY",
			want:  "NAME=PATH:KEY",
		},
		{
			name:  "multi keys with prefix",
			input: "path=PATH,prefix=PREFIX_",
			want:  "PATH:PREFIX_",
		},
		{
			name:  "multi keys",
			input: "path=PATH",
			want:  "PATH",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := sourced.ParseEnv(tt.input)
			require.NoError(t, err)
			got := e.String()
			require.Equal(t, tt.want, got)
		})
	}
}
