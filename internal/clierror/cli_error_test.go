package clierror

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CLIError_String(t *testing.T) {
	tests := []struct {
		name string
		err  Error
		want string
	}{
		{
			name: "empty",
			err:  Error{},
			want: "Error:\n  \n\n",
		},
		{
			name: "error",
			err: Error{
				Message: "error",
			},
			want: "Error:\n  error\n\n",
		},
		{
			name: "error and details",
			err: Error{
				Message: "error",
				Details: "details",
			},
			want: "Error:\n  error\n\nError Details:\n  details\n\n",
		},
		{
			name: "error, details and hints",
			err: Error{
				Message: "error",
				Details: "details",
				Hints:   []string{"hint1", "hint2"},
			},
			want: "Error:\n  error\n\nError Details:\n  details\n\nHints:\n  - hint1\n  - hint2\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}

// test Wrap function
func Test_CLIError_Wrap(t *testing.T) {
	tests := []struct {
		name    string
		err     Error
		message string
		hints   []string
		want    string
	}{
		{
			name:    "Add to empty error",
			err:     Error{},
			message: "error",
			want:    "Error:\n  error\n\n",
		},
		{
			name:    "Add with hints to empty error",
			err:     Error{},
			message: "error",
			hints:   []string{"hint1", "hint2"},
			want:    "Error:\n  error\n\nHints:\n  - hint1\n  - hint2\n",
		},
		{
			name: "add to error",
			err: Error{
				Message: "error",
			},
			message: "error",
			hints:   []string{"hint1", "hint2"},
			want:    "Error:\n  error\n\nError Details:\n  error\n\nHints:\n  - hint1\n  - hint2\n",
		},
		{
			name: "add to error with details",
			err: Error{
				Message: "previous",
				Details: "details",
			},
			message: "error",
			hints:   []string{"hint1", "hint2"},
			want:    "Error:\n  error\n\nError Details:\n  previous: details\n\nHints:\n  - hint1\n  - hint2\n",
		},
		{
			name: "add to error with details and hints",
			err: Error{
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
			err := tt.err.Wrap(tt.message, tt.hints)
			assert.Equal(t, tt.want, err.Error())
		})
	}
}
