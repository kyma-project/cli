package prompt_test

import (
	"bytes"
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/prompt"
	"github.com/stretchr/testify/require"
)

func TestOneOfStringListPrompt_Table(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		values     []string
		parseFunc  func(string) (string, error)
		want       string
		wantErr    string
		wantOutput string
	}{
		{
			name:       "Valid input",
			input:      "banana\n",
			values:     []string{"apple", "banana", "orange"},
			want:       "banana",
			wantErr:    "",
			wantOutput: "Select a fruit:\n - apple\n - banana\n - orange\n\nType the version number: ",
		},
		{
			name:       "Invalid input",
			input:      "Faraon Ramzes XIII\n",
			values:     []string{"apple", "banana", "orange"},
			want:       "",
			wantErr:    "provided value is not present on the list: Faraon",
			wantOutput: "Select a fruit:\n - apple\n - banana\n - orange\n\nType the version number: ",
		},
		{
			name:       "Empty input",
			input:      "\n",
			values:     []string{"apple", "banana", "orange"},
			want:       "",
			wantErr:    "no value was selected",
			wantOutput: "Select a fruit:\n - apple\n - banana\n - orange\n\nType the version number: ",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := bytes.NewBufferString(tc.input)
			output := bytes.NewBuffer([]byte{})
			listPrompt := prompt.NewCustomOneOfStringList(
				input,
				output,
				"Select a fruit:",
				"Type the version number: ",
				tc.values,
			)

			got, err := listPrompt.Prompt()

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Equal(t, tc.wantErr, err.Error())
				require.Equal(t, "", got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
			require.Equal(t, tc.wantOutput, output.String())
		})
	}
}
