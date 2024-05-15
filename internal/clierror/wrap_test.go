package clierror

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Wrap(t *testing.T) {
	tests := []struct {
		name    string
		inside  any
		outside []modifier
		want    string
	}{
		{
			name:    "Wrap basic error",
			inside:  fmt.Errorf("error"),
			outside: []modifier{Message("outside"), Hints("hint1", "hint2")},
			want:    "Error:\n  outside\n\nError Details:\n  error\n\nHints:\n  - hint1\n  - hint2\n",
		},
		{
			name: "Wrap string error",
			inside: "error",
			outside: []modifier{Message("outside"), Hints("hint1", "hint2")},
			want: "Error:\n  outside\n\nError Details:\n  error\n\nHints:\n  - hint1\n  - hint2\n",
		},
		{
			name:    "Wrap Error",
			inside:  &clierror{message: "error", details: "details", hints: []string{"hint1", "hint2"}},
			outside: []modifier{Message("outside"), Hints("hint3", "hint4")},
			want:    "Error:\n  outside\n\nError Details:\n  error: details\n\nHints:\n  - hint3\n  - hint4\n  - hint1\n  - hint2\n",
		},
		{
			name:    "Wrap Error with no hints",
			inside:  &clierror{message: "error", details: "details"},
			outside: []modifier{Message("outside"), Hints("hint3", "hint4")},
			want:    "Error:\n  outside\n\nError Details:\n  error: details\n\nHints:\n  - hint3\n  - hint4\n",
		},
		{
			name:    "Wrap Error with no details",
			inside:  &clierror{message: "error", hints: []string{"hint1", "hint2"}},
			outside: []modifier{Message("outside"), Hints("hint3", "hint4")},
			want:    "Error:\n  outside\n\nError Details:\n  error\n\nHints:\n  - hint3\n  - hint4\n  - hint1\n  - hint2\n",
		},
		{
			name:    "Wrap Error with no message",
			inside:  &clierror{details: "details", hints: []string{"hint1", "hint2"}},
			outside: []modifier{Message("outside"), Hints("hint3", "hint4")},
			want:    "Error:\n  outside\n\nError Details:\n  details\n\nHints:\n  - hint3\n  - hint4\n  - hint1\n  - hint2\n",
		},
		{
			name:    "Wrap Error with no message and no hints",
			inside:  &clierror{details: "details"},
			outside: []modifier{Message("outside"), Hints("hint3", "hint4")},
			want:    "Error:\n  outside\n\nError Details:\n  details\n\nHints:\n  - hint3\n  - hint4\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Wrap(tt.inside, tt.outside...)
			assert.Equal(t, tt.want, err.String())
		})
	}
}
