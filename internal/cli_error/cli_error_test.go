package cli_error

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CLIError_String(t *testing.T) {
	tests := []struct {
		name string
		f    DetailedError
		want string
	}{
		{
			name: "empty",
			f:    DetailedError{},
			want: "Error:\n  \n\n",
		},
		{
			name: "error",
			f: DetailedError{
				Message: "error",
			},
			want: "Error:\n  error\n\n",
		},
		{
			name: "error and details",
			f: DetailedError{
				Message: "error",
				Details: "details",
			},
			want: "Error:\n  error\n\nError Details:\n  details\n\n",
		},
		{
			name: "error, details and hints",
			f: DetailedError{
				Message: "error",
				Details: "details",
				Hints:   []string{"hint1", "hint2"},
			},
			want: "Error:\n  error\n\nError Details:\n  details\n\nHints:\n  - hint1\n  - hint2\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.f.Error())
		})
	}
}

// test Wrap function
func Test_CLIError_Wrap(t *testing.T) {
	tests := []struct {
		name    string
		f       DetailedError
		message string
		hints   []string
		want    string
	}{
		{
			name:    "Add to empty error",
			f:       DetailedError{},
			message: "error",
			want:    "Error:\n  error\n\n",
		},
		{
			name:    "Add with hints to empty error",
			f:       DetailedError{},
			message: "error",
			hints:   []string{"hint1", "hint2"},
			want:    "Error:\n  error\n\nHints:\n  - hint1\n  - hint2\n",
		},
		{
			name: "add to error",
			f: DetailedError{
				Message: "error",
			},
			message: "error",
			hints:   []string{"hint1", "hint2"},
			want:    "Error:\n  error\n\nError Details:\n  error\n\nHints:\n  - hint1\n  - hint2\n",
		},
		{
			name: "add to error with details",
			f: DetailedError{
				Message: "previous",
				Details: "details",
			},
			message: "error",
			hints:   []string{"hint1", "hint2"},
			want:    "Error:\n  error\n\nError Details:\n  previous: details\n\nHints:\n  - hint1\n  - hint2\n",
		},
		{
			name: "add to error with details and hints",
			f: DetailedError{
				Message: "previous",
				Details: "details",
				Hints:   []string{"hint1", "hint2"},
			},
			message: "error",
			hints:   []string{"hint3", "hint4"},
			want:    "Error:\n  error\n\nError Details:\n  previous: details\n\nHints:\n  - hint3\n  - hint4\n  - hint1\n  - hint2\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.Wrap(tt.message, tt.hints)
			assert.Equal(t, tt.want, tt.f.Error())
		})
	}
}
