package module

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("<file-contents>"))
	if err != nil {
		return
	}
}

func TestDownloadRemoteFileToTempFile(t *testing.T) {
	t.Parallel()

	mockServer := httptest.NewServer(http.HandlerFunc(testHandler))
	defer mockServer.Close()
	tmpFiles := NewTmpFilesManager()
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
				url:      mockServer.URL,
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
