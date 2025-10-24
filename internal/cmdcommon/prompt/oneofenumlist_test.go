package prompt

import (
	"bytes"
	"testing"

	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/stretchr/testify/require"
)

func TestOneOfEnumList_Prompt_Table(t *testing.T) {
	values := []EnumValueWithDescription{
		{value: "apple", description: "is a primary source of an apple juice"},
		{value: "banana", description: "is an important element of monkeys diet"},
		{value: "orange", description: "apparently doesn't rhyme that well with many words"},
	}

	tests := []struct {
		name           string
		input          string
		expectError    bool
		expectSelected string
		expectErrMsg   string
	}{
		{
			name:           "SelectsCorrectValue",
			input:          "2\n",
			expectError:    false,
			expectSelected: "banana",
		},
		{
			name:         "EmptyInput",
			input:        "\n",
			expectError:  true,
			expectErrMsg: "no value was selected",
		},
		{
			name:         "InvalidNumber",
			input:        "5\n",
			expectError:  true,
			expectErrMsg: "invalid option selected",
		},
		{
			name:         "NonNumberInput",
			input:        "banana\n",
			expectError:  true,
			expectErrMsg: "provided value is not a number",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := bytes.NewBufferString(tc.input)
			output := &bytes.Buffer{}
			prompt := OneOfEnumList{
				reader:                 input,
				printer:                out.NewToWriter(output),
				message:                "Choose:",
				promptText:             "Your choice:",
				valuesWithDescriptions: values,
			}

			selected, err := prompt.Prompt()
			if tc.expectError {
				require.Error(t, err)
				if tc.expectErrMsg != "" {
					require.Contains(t, err.Error(), tc.expectErrMsg)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectSelected, selected)
			}
		})
	}
}
