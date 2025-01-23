package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNullableString_Set(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		want       *string
		wantString string
		wantErr    bool
	}{
		{
			name:       "empty",
			value:      "",
			want:       nil,
			wantString: "",
			wantErr:    false,
		},
		{
			name:       "set value",
			value:      "test value",
			want:       toptr("test value"),
			wantString: "test value",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nb := NullableString{}
			err := nb.Set(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, nb.Value)
			require.Equal(t, tt.wantString, nb.String())
			require.Equal(t, "string", nb.Type())
		})
	}
}

func toptr[T any](val T) *T {
	return &val
}
