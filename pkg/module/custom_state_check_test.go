package module

import (
	"testing"
)

func TestIsValidMappedState(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Ready State is Valid",
			args: args{s: "Ready"},
			want: true,
		},
		{
			name: "Error State is Valid",
			args: args{s: "Error"},
			want: true,
		},
		{
			name: "Warning State is Valid",
			args: args{s: "Warning"},
			want: true,
		},
		{
			name: "Invalid State is Recognized",
			args: args{s: "RandomState"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidMappedState(tt.args.s); got != tt.want {
				t.Errorf("IsValidMappedState() = %v, want %v", got, tt.want)
			}
		})
	}
}
