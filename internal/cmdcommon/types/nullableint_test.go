package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNullableInt64_Set(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		want       *int64
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
			value:      "23",
			want:       toptr(int64(23)),
			wantString: "23",
			wantErr:    false,
		},
		{
			name:       "incorrect",
			value:      "incorrect",
			want:       nil,
			wantString: "",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nb := NullableInt64{}
			err := nb.Set(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, nb.Value)
			require.Equal(t, tt.wantString, nb.String())
			require.Equal(t, "int", nb.Type())
		})
	}
}
