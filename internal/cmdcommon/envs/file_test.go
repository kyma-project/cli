package envs

import (
	"errors"
	"os"
	"syscall"
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types/sourced"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestBuildEnvsFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := tmpDir + "/file.env"
	err := os.WriteFile(filePath, []byte("KEY1=VALUE1\nKEY2=VALUE2\nKEY3=VALUE3\n"), 0600)
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		envs    types.SourcedEnvArray
		want    []corev1.EnvVar
		wantErr error
	}{
		{
			name: "missing file path",
			envs: types.SourcedEnvArray{
				Values: []sourced.Env{
					{
						Name:        "KEY",
						LocationKey: "PREFIX_",
					},
				},
			},
			wantErr: errors.New("missing file path in env: 'KEY=:PREFIX_'"),
		},
		{
			name: "single env vars from files",
			envs: types.SourcedEnvArray{
				Values: []sourced.Env{
					{
						Name:        "KEY",
						Location:    filePath,
						LocationKey: "KEY1",
					},
					{
						Name:        "KEY2",
						Location:    filePath,
						LocationKey: "KEY2",
					},
				},
			},
			want: []corev1.EnvVar{
				{
					Name:  "KEY",
					Value: "VALUE1",
				},
				{
					Name:  "KEY2",
					Value: "VALUE2",
				},
			},
		},
		{
			name: "missing key for single env var from file",
			envs: types.SourcedEnvArray{
				Values: []sourced.Env{
					{
						Name:     "KEY",
						Location: filePath,
					},
				},
			},
			wantErr: errors.New("key '' not found in env file '" + filePath + "'"),
		},
		{
			name: "multi keys from file",
			envs: types.SourcedEnvArray{
				Values: []sourced.Env{
					{
						Location: filePath,
					},
					{
						Location:           filePath,
						LocationKeysPrefix: "PREFIX_",
					},
				},
			},
			want: []corev1.EnvVar{
				{
					Name:  "KEY1",
					Value: "VALUE1",
				},
				{
					Name:  "KEY2",
					Value: "VALUE2",
				},
				{
					Name:  "KEY3",
					Value: "VALUE3",
				},
				{
					Name:  "PREFIX_KEY1",
					Value: "VALUE1",
				},
				{
					Name:  "PREFIX_KEY2",
					Value: "VALUE2",
				},
				{
					Name:  "PREFIX_KEY3",
					Value: "VALUE3",
				},
			},
		},
		{
			name: "file does not exist",
			envs: types.SourcedEnvArray{
				Values: []sourced.Env{
					{
						Location: tmpDir + "nonexistent",
					},
				},
			},
			wantErr: &os.PathError{Op: "open", Path: tmpDir + "nonexistent", Err: syscall.Errno(2)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := BuildFromFile(tt.envs)
			require.Equal(t, tt.wantErr, gotErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}
