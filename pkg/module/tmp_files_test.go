package module

import (
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDownloadRemoteFileToTempFile(t *testing.T) {
	t.Parallel()

	tmpFiles := NewTmpFilesManager()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://example.com/manifest.yaml",
		httpmock.NewBytesResponder(200, []byte("<file-contents>")))
	defer tmpFiles.DeleteTmpFiles()

	type args struct {
		url      string
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "file download successful",
			args: args{
				url:      "https://example.com/manifest.yaml",
				filename: "manifest-*.yaml",
			},
			want:    []byte("<file-contents>"),
			wantErr: false,
		},
		{
			name: "invalid url results in error",
			args: args{
				url:      "invalid-url",
				filename: "manifest-*.yaml",
			},
			want:    []byte{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := tmpFiles.DownloadRemoteFileToTmpFile(tt.args.url, os.TempDir(), tt.args.filename)

			if err != nil && !tt.wantErr {
				t.Errorf("unexpected error occurred: %s", err.Error())
				return
			}
			if err != nil && tt.wantErr {
				return
			}
			if err == nil && tt.wantErr {
				t.Errorf("expected error did not occur: %s", err.Error())
				return
			}

			fileContent, err := os.ReadFile(got)
			if err != nil {
				t.Errorf("created file could not be read: %s", err.Error())
				return
			}
			assert.Equalf(t, tt.want, fileContent, "DownloadRemoteFileToTmpFile(%v, %v, %v)",
				tt.args.url, "", tt.args.filename)
		})
	}
}
