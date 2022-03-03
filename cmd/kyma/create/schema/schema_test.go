package schema

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/kyma-project/cli/internal/cli"
)

var (
	refMap = map[string]func() ([]byte, error){
		"test_bad": func() ([]byte, error) {
			return nil, fmt.Errorf("test error")
		},
		"test_good": func() ([]byte, error) {
			return []byte("OK!"), nil
		},
	}
)

func Test_command_Run(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantedOutput string
		wantErr      bool
	}{
		{
			name:    "unknown schema",
			args:    []string{"unknown_schema"},
			wantErr: true,
		},
		{
			name:    "invalid argument",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "schema reflect error",
			args:    []string{"test_bad"},
			wantErr: true,
		},
		{
			name:         "OK",
			args:         []string{"test_good"},
			wantErr:      false,
			wantedOutput: "OK!",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buff bytes.Buffer
			c := NewCmd(NewOptions(cli.NewOptions(), &buff, refMap))

			if err := c.RunE(nil, tt.args); (err != nil) != tt.wantErr {
				t.Errorf("command.Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			output := buff.String()
			if tt.wantedOutput != "" && tt.wantedOutput != output {
				t.Errorf("command.Run() output = %v, wantOutput %v", output, tt.wantedOutput)
			}
		})
	}
}
