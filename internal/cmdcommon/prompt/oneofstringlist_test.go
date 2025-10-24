package prompt

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"testing/iotest"

	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/stretchr/testify/require"
)

func TestOneOfStringListPrompt_Table(t *testing.T) {
	tests := []struct {
		name        string
		inputReader io.Reader
		values      []string
		parseFunc   func(string) (string, error)
		want        string
		wantErr     string
		wantOutput  string
	}{
		{
			name:        "Valid input",
			inputReader: bytes.NewBufferString("banana"),
			values:      []string{"apple", "banana", "orange"},
			want:        "banana",
			wantErr:     "",
			wantOutput:  "Select a fruit:\n - apple\n - banana\n - orange\n\nType the version number: ",
		},
		{
			name:        "Invalid input",
			inputReader: bytes.NewBufferString("Faraon"),
			values:      []string{"apple", "banana", "orange"},
			want:        "",
			wantErr:     "provided value is not present on the list: Faraon",
			wantOutput:  "Select a fruit:\n - apple\n - banana\n - orange\n\nType the version number: ",
		},
		{
			name:        "Empty input",
			inputReader: bytes.NewBufferString(""),
			values:      []string{"apple", "banana", "orange"},
			want:        "",
			wantErr:     "no value was selected",
			wantOutput:  "Select a fruit:\n - apple\n - banana\n - orange\n\nType the version number: ",
		},
		{
			name:        "Invalid reader",
			inputReader: iotest.ErrReader(errors.New("test error")),
			values:      []string{"apple", "banana", "orange"},
			want:        "",
			wantErr:     "test error",
			wantOutput:  "Select a fruit:\n - apple\n - banana\n - orange\n\nType the version number: ",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			output := bytes.NewBuffer([]byte{})
			listPrompt := OneOfStringList{
				reader:     tc.inputReader,
				printer:    out.NewToWriter(output),
				message:    "Select a fruit:",
				promptText: "Type the version number: ",
				values:     tc.values,
			}

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
