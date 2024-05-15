package clierror

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Wrap(t *testing.T) {
	tests := []struct {
		name    string
		inside  error
		outside Error
		want    string
	}{
		{
			name:    "Wrap basic error",
			inside:  fmt.Errorf("error"),
			outside: &clierror{message: "outside", hints: []string{"hint1", "hint2"}},
			want:    "Error:\n  outside\n\nError Details:\n  error\n\nHints:\n  - hint1\n  - hint2\n",
		},
		{
			name:    "Wrap basic error with no hints",
			inside:  fmt.Errorf("error"),
			outside: &clierror{message: "outside", details: "details"},
			want:    "Error:\n  outside\n\nError Details:\n  details: error\n\n",
		},
		{
			name:    "Wrap basic error with no message",
			inside:  fmt.Errorf("error"),
			outside: &clierror{details: "outside", hints: []string{"hint1", "hint2"}},
			want:    "Error:\n  error\n\nError Details:\n  outside\n\nHints:\n  - hint1\n  - hint2\n",
		},
		{
			name:    "Wrap basic error with no message and no hints",
			inside:  fmt.Errorf("error"),
			outside: &clierror{details: "details"},
			want:    "Error:\n  error\n\nError Details:\n  details\n\n",
		},
		{
			name:    "Wrap basic error add hints only",
			inside:  fmt.Errorf("error"),
			outside: &clierror{hints: []string{"hint1", "hint2"}},
			want:    "Error:\n  error\n\nHints:\n  - hint1\n  - hint2\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Wrap(tt.inside, tt.outside)
			assert.Equal(t, tt.want, err.String())
		})
	}
}

func Test_WrapE(t *testing.T) {
	tests := []struct {
		name    string
		inside  Error
		outside Error
		want    string
	}{
		{
			name:    "Wrap Error",
			inside:  &clierror{message: "error", details: "details", hints: []string{"hint1", "hint2"}},
			outside: &clierror{message: "outside", hints: []string{"hint3", "hint4"}},
			want:    "Error:\n  outside\n\nError Details:\n  error: details\n\nHints:\n  - hint3\n  - hint4\n  - hint1\n  - hint2\n",
		},
		{
			name:    "Wrap Error with no hints",
			inside:  &clierror{message: "error", details: "details"},
			outside: &clierror{message: "outside", hints: []string{"hint3", "hint4"}},
			want:    "Error:\n  outside\n\nError Details:\n  error: details\n\nHints:\n  - hint3\n  - hint4\n",
		},
		{
			name:    "Wrap Error with no details",
			inside:  &clierror{message: "error", hints: []string{"hint1", "hint2"}},
			outside: &clierror{message: "outside", hints: []string{"hint3", "hint4"}},
			want:    "Error:\n  outside\n\nError Details:\n  error\n\nHints:\n  - hint3\n  - hint4\n  - hint1\n  - hint2\n",
		},
		{
			name:    "Wrap Error with no message",
			inside:  &clierror{details: "details", hints: []string{"hint1", "hint2"}},
			outside: &clierror{message: "outside", hints: []string{"hint3", "hint4"}},
			want:    "Error:\n  outside\n\nError Details:\n  details\n\nHints:\n  - hint3\n  - hint4\n  - hint1\n  - hint2\n",
		},
		{
			name:    "Wrap Error with no message and no hints",
			inside:  &clierror{details: "details"},
			outside: &clierror{message: "outside", hints: []string{"hint3", "hint4"}},
			want:    "Error:\n  outside\n\nError Details:\n  details\n\nHints:\n  - hint3\n  - hint4\n",
		},
		{
			name:    "Wrap Error add hints only",
			inside:  &clierror{message: "error", details: "details", hints: []string{"hint1", "hint2"}},
			outside: &clierror{hints: []string{"hint3", "hint4"}},
			want:    "Error:\n  error\n\nError Details:\n  details\n\nHints:\n  - hint3\n  - hint4\n  - hint1\n  - hint2\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WrapE(tt.inside, tt.outside)
			assert.Equal(t, tt.want, err.String())
		})
	}
}
