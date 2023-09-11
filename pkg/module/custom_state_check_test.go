package module

import (
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"reflect"
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

func TestValidateCustomStateCheck(t *testing.T) {
	type args struct {
		paths  []string
		values []string
		states []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "No error when no arguments are provided",
			args: args{
				paths:  nil,
				values: nil,
				states: nil,
			},
			wantErr: false,
		},
		{
			name: "No error when valid arguments are provided",
			args: args{
				paths:  []string{"status.health"},
				values: []string{"green"},
				states: []string{"Ready"},
			},
			wantErr: false,
		},
		{
			name: "Error when only partial argument is provided",
			args: args{
				paths:  []string{"status.health"},
				values: nil,
				states: []string{"Ready"},
			},
			wantErr: true,
		},
		{
			name: "Error when arguments are of different lengths",
			args: args{
				paths:  []string{"status.health", "status.health"},
				values: []string{"green", "red"},
				states: []string{"Ready"},
			},
			wantErr: true,
		},
		{
			name: "Error when provided state is not a valid kyma state",
			args: args{
				paths:  []string{"status.health"},
				values: []string{"green"},
				states: []string{"FatalError"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateCustomStateCheck(tt.args.paths, tt.args.values, tt.args.states); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCustomStateCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateChecks(t *testing.T) {
	type args struct {
		paths  []string
		values []string
		states []string
	}
	tests := []struct {
		name string
		args args
		want []v1beta2.CustomStateCheck
	}{
		{
			name: "Empty list for no arguments",
			args: args{
				paths:  nil,
				values: nil,
				states: nil,
			},
			want: nil,
		},
		{
			name: "Correct generation of checks",
			args: args{
				paths:  []string{"status.health"},
				values: []string{"green"},
				states: []string{"Ready"},
			},
			want: []v1beta2.CustomStateCheck{
				{
					JSONPath:    "status.health",
					Value:       "green",
					MappedState: v1beta2.StateReady,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateChecks(tt.args.paths, tt.args.values, tt.args.states); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateChecks() = %v, want %v", got, tt.want)
			}
		})
	}
}
