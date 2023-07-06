package cli

import (
	"errors"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/kyma-project/cli/internal/cli/alpha/module"
	"testing"
)

func TestGetExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "correct exit code for single error",
			err:  module.ErrKymaInWarningState,
			want: ErrorCodeMap[module.ErrKymaInWarningState],
		},
		{
			name: "correct exit code for wrapped error",
			err: fmt.Errorf("wrapped error 2: %w",
				fmt.Errorf("wrapped error 1: %w",
					module.ErrKymaInWarningState)),
			want: ErrorCodeMap[module.ErrKymaInWarningState],
		},
		{
			name: "correct exit code for error in a list",
			err: retry.Error([]error{
				errors.New("random non-mapped error"),
				module.ErrKymaInWarningState,
			}),
			want: ErrorCodeMap[module.ErrKymaInWarningState],
		},
		{
			name: "default exit code for non-mapped error",
			err:  errors.New("random non-mapped error"),
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetExitCode(tt.err); got != tt.want {
				t.Errorf("GetExitCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
