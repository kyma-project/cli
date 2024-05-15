package clierror

import (
	"fmt"
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
		outside *Error
		want    string
	}{
		{
			name:    "Add to empty error",
			err:     Error{},
			outside: &Error{Message: "error"},
			want:    "Error:\n  error\n\n",
		},
		{
			name: "Add with hints to empty error",
			err:  Error{},
			outside: &Error{
				Message: "error",
				Hints:   []string{"hint1", "hint2"},
			},
			want: "Error:\n  error\n\nHints:\n  - hint1\n  - hint2\n",
		},
		{
			name: "add to error",
			err: Error{
				Message: "error",
			},
			outside: &Error{
				Message: "error",
				Hints:   []string{"hint1", "hint2"},
			},
			want: "Error:\n  error\n\nError Details:\n  error\n\nHints:\n  - hint1\n  - hint2\n",
		},
		{
			name: "add to error with details",
			err: Error{
				Message: "previous",
				Details: "details",
			},
			outside: &Error{
				Message: "error",
				Hints:   []string{"hint1", "hint2"},
			},
			want: "Error:\n  error\n\nError Details:\n  previous: details\n\nHints:\n  - hint1\n  - hint2\n",
		},
		{
			name: "add to error with details and hints",
			err: Error{
				Message: "previous",
				Details: "details",
				Hints:   []string{"hint1", "hint2"},
			},
			outside: &Error{
				Message: "error",
				Hints:   []string{"hint3", "hint4"},
			},
			want: "Error:\n  error\n\nError Details:\n  previous: details\n\nHints:\n  - hint3\n  - hint4\n  - hint1\n  - hint2\n",
		},
		{
			name: "add to error with more details and hints",
			err: Error{
				Message: "previous",
				Details: "details",
				Hints:   []string{"hint1", "hint2"},
			},
			outside: &Error{
				Message: "error",
				Details: "moreDetails",
				Hints:   []string{"hint3", "hint4"},
			},
			want: "Error:\n  error\n\nError Details:\n  moreDetails: previous: details\n\nHints:\n  - hint3\n  - hint4\n  - hint1\n  - hint2\n",
		},
		{
			name: "empty wrap",
			err: Error{
				Message: "error",
				Details: "details",
				Hints:   []string{"hint1", "hint2"},
			},
			outside: &Error{},
			want:    "Error:\n  error\n\nError Details:\n  details\n\nHints:\n  - hint1\n  - hint2\n",
		},
		{
			name: "wrap with details",
			err: Error{
				Message: "error",
				Details: "details",
				Hints:   []string{"hint1", "hint2"},
			},
			outside: &Error{Details: "newDetails"},
			want:    "Error:\n  error\n\nError Details:\n  newDetails: details\n\nHints:\n  - hint1\n  - hint2\n",
		},

		{
			name: "wrap with hints",
			err: Error{
				Message: "error",
				Details: "details",
				Hints:   []string{"hint1", "hint2"},
			},
			outside: &Error{Hints: []string{"hint3", "hint4"}},
			want:    "Error:\n  error\n\nError Details:\n  details\n\nHints:\n  - hint3\n  - hint4\n  - hint1\n  - hint2\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.err.wrap(tt.outside)
			assert.Equal(t, tt.want, err.Error())
		})
	}
}

func Test_Wrap(t *testing.T) {
	tests := []struct {
		name    string
		inside  error
		outside *Error
		want    string
	}{
		{
			name:    "Wrap string error",
			inside:  fmt.Errorf("error"),
			outside: &Error{Message: "outside"},
			want:    "Error:\n  outside\n\nError Details:\n  error\n\n",
		},
		{
			name:    "Wrap Error",
			inside:  &Error{Message: "error", Details: "details"},
			outside: &Error{Message: "outside"},
			want:    "Error:\n  outside\n\nError Details:\n  error: details\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Wrap(tt.inside, tt.outside)
			assert.Equal(t, tt.want, err.Error())
		})
	}
}
