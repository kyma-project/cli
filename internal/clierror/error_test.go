package clierror

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CLIError_String(t *testing.T) {
	tests := []struct {
		name string
		err  clierror
		want string
	}{
		{
			name: "empty",
			err:  clierror{},
			want: "Error:\n  \n\n",
		},
		{
			name: "error",
			err: clierror{
				message: "error",
			},
			want: "Error:\n  error\n\n",
		},
		{
			name: "error and details",
			err: clierror{
				message: "error",
				details: "details",
			},
			want: "Error:\n  error\n\nError Details:\n  details\n\n",
		},
		{
			name: "error, details and hints",
			err: clierror{
				message: "error",
				details: "details",
				hints:   []string{"hint1", "hint2"},
			},
			want: "Error:\n  error\n\nError Details:\n  details\n\nHints:\n  - hint1\n  - hint2\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.String())
		})
	}
}

// test Wrap function
func Test_CLIError_Wrap(t *testing.T) {
	tests := []struct {
		name    string
		err     clierror
		outside *clierror
		want    string
	}{
		{
			name:    "Add to empty error",
			err:     clierror{},
			outside: &clierror{message: "error"},
			want:    "Error:\n  error\n\n",
		},
		{
			name: "Add with hints to empty error",
			err:  clierror{},
			outside: &clierror{
				message: "error",
				hints:   []string{"hint1", "hint2"},
			},
			want: "Error:\n  error\n\nHints:\n  - hint1\n  - hint2\n",
		},
		{
			name: "add to error",
			err: clierror{
				message: "error",
			},
			outside: &clierror{
				message: "error",
				hints:   []string{"hint1", "hint2"},
			},
			want: "Error:\n  error\n\nError Details:\n  error\n\nHints:\n  - hint1\n  - hint2\n",
		},
		{
			name: "add to error with details",
			err: clierror{
				message: "previous",
				details: "details",
			},
			outside: &clierror{
				message: "error",
				hints:   []string{"hint1", "hint2"},
			},
			want: "Error:\n  error\n\nError Details:\n  previous: details\n\nHints:\n  - hint1\n  - hint2\n",
		},
		{
			name: "add to error with details and hints",
			err: clierror{
				message: "previous",
				details: "details",
				hints:   []string{"hint1", "hint2"},
			},
			outside: &clierror{
				message: "error",
				hints:   []string{"hint3", "hint4"},
			},
			want: "Error:\n  error\n\nError Details:\n  previous: details\n\nHints:\n  - hint3\n  - hint4\n  - hint1\n  - hint2\n",
		},
		{
			name: "add to error with more details and hints",
			err: clierror{
				message: "previous",
				details: "details",
				hints:   []string{"hint1", "hint2"},
			},
			outside: &clierror{
				message: "error",
				details: "moreDetails",
				hints:   []string{"hint3", "hint4"},
			},
			want: "Error:\n  error\n\nError Details:\n  moreDetails: previous: details\n\nHints:\n  - hint3\n  - hint4\n  - hint1\n  - hint2\n",
		},
		{
			name: "empty wrap",
			err: clierror{
				message: "error",
				details: "details",
				hints:   []string{"hint1", "hint2"},
			},
			outside: &clierror{},
			want:    "Error:\n  error\n\nError Details:\n  details\n\nHints:\n  - hint1\n  - hint2\n",
		},
		{
			name: "wrap with details",
			err: clierror{
				message: "error",
				details: "details",
				hints:   []string{"hint1", "hint2"},
			},
			outside: &clierror{details: "newDetails"},
			want:    "Error:\n  error\n\nError Details:\n  newDetails: details\n\nHints:\n  - hint1\n  - hint2\n",
		},

		{
			name: "wrap with hints",
			err: clierror{
				message: "error",
				details: "details",
				hints:   []string{"hint1", "hint2"},
			},
			outside: &clierror{hints: []string{"hint3", "hint4"}},
			want:    "Error:\n  error\n\nError Details:\n  details\n\nHints:\n  - hint3\n  - hint4\n  - hint1\n  - hint2\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.err.wrap(tt.outside)
			assert.Equal(t, tt.want, err.String())
		})
	}
}
