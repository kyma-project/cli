package vscode

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/kyma-project/hydroform/function/pkg/workspace"
)

type errWriter struct{}

func (w *errWriter) Write(_ []byte) (n int, err error) {
	return -1, fmt.Errorf("write error")
}

func TestConfiguration_build(t *testing.T) {
	type args struct {
		dirPath        string
		writerProvider workspace.WriterProvider
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "write error",
			args: args{
				writerProvider: func(path string) (io.Writer, func() error, error) {
					return &errWriter{}, nil, nil
				},
			},
			wantErr: true,
		},
		{
			name: "happy path",
			args: args{
				writerProvider: func(path string) (io.Writer, func() error, error) {
					return &bytes.Buffer{}, nil, nil
				},
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			if err := Workspace.build(tt.args.dirPath, tt.args.writerProvider); (err != nil) != tt.wantErr {
				t.Errorf("Configuration.build() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
