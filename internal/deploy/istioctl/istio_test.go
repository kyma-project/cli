package istioctl

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path"
	"testing"

	"go.uber.org/zap"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestInstallation_getIstioVersion(t *testing.T) {
	type fields struct {
		WorkspacePath  string
		IstioChartPath string
		istioVersion   string
		osExt          string
		binName        string
		winBinName     string
	}
	tests := []struct {
		name            string
		fields          fields
		expectedVersion string
		wantErr         bool
	}{
		{
			name:            "Fetch Istio Version",
			fields:          fields{IstioChartPath: "testdata/Chart.yaml"},
			expectedVersion: "1.11.2",
			wantErr:         false,
		},
		{
			name:            "Istio Chart not existing",
			fields:          fields{IstioChartPath: "testdata/nonExisting.yaml"},
			expectedVersion: "",
			wantErr:         true,
		},
		{
			name:            "Corrupted Istio Chart",
			fields:          fields{IstioChartPath: "testdata/corruptedChart.yaml"},
			expectedVersion: "",
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			i := &Installation{
				WorkspacePath:  tt.fields.WorkspacePath,
				IstioChartPath: tt.fields.IstioChartPath,
				istioVersion:   tt.fields.istioVersion,
				osExt:          tt.fields.osExt,
				binName:        tt.fields.binName,
				winBinName:     tt.fields.winBinName,
				logger:         zap.NewNop().Sugar(),
			}
			err := i.getIstioVersion()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expectedVersion, i.istioVersion)
		})
	}
}

func TestInstallation_checkIfBinaryExists(t *testing.T) {
	type fields struct {
		binPath string
	}
	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		{name: "File exists", fields: fields{binPath: "testdata/Chart.yaml"}, want: true, wantErr: false},
		{name: "File does not exist", fields: fields{binPath: "testdata/nonexistent"}, want: false, wantErr: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			i := &Installation{
				binPath: tt.fields.binPath,
				logger:  zap.NewNop().Sugar(),
			}
			got, err := i.checkIfBinaryExists()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_unGzip(t *testing.T) {
	type args struct {
		source       string
		target       string
		deleteSource bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "unGzip File",
			args:    args{source: "testdata/istio_mock.tar.gz", target: "testdata/istio.tar", deleteSource: false},
			wantErr: false,
		},
		{
			name:    "File does not exist",
			args:    args{source: "testdata/nonexistent.tar.gz", target: "testdata/istio.tar", deleteSource: false},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := unGzip(tt.args.source, tt.args.target, tt.args.deleteSource)
			if !tt.wantErr {
				require.NoError(t, err)
				_, err := os.Stat(tt.args.target)
				require.NoError(t, err)
				err = os.Remove(tt.args.target)
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				_, err := os.Stat(tt.args.target)
				require.Error(t, err)
			}

		})
	}
}

func Test_unTar(t *testing.T) {
	type args struct {
		source       string
		target       string
		deleteSource bool
	}
	tests := []struct {
		name         string
		args         args
		expectedFile string
		wantErr      bool
	}{
		{
			name:         "unTar File",
			args:         args{source: "testdata/istio_mock.tar", target: "testdata", deleteSource: false},
			expectedFile: "testdata/istio.txt",
			wantErr:      false,
		},
		{
			name:         "File does not exist",
			args:         args{source: "testdata/nonexistent.tar", target: "testdata", deleteSource: false},
			expectedFile: "testdata/istio.txt",
			wantErr:      true,
		},
		{
			name:         "File path contains `..` - Zip Slip Check",
			args:         args{source: "../istiomock.tar", target: "testdata", deleteSource: false},
			expectedFile: "testdata/istio.txt",
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := unTar(tt.args.source, tt.args.target, tt.args.deleteSource)
			if !tt.wantErr {
				require.NoError(t, err)
				_, err := os.Stat(tt.expectedFile)
				require.NoError(t, err)
				err = os.Remove(tt.expectedFile)
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				_, err := os.Stat(tt.expectedFile)
				require.Error(t, err)
			}
		})
	}
}

func Test_unZip(t *testing.T) {
	type args struct {
		source       string
		target       string
		deleteSource bool
	}
	tests := []struct {
		name         string
		args         args
		expectedFile string
		wantErr      bool
	}{
		{
			name:         "unZip File",
			args:         args{source: "testdata/istio_mock.zip", target: "testdata", deleteSource: false},
			expectedFile: "testdata/istio.txt",
			wantErr:      false,
		},
		{
			name:         "File does not exist",
			args:         args{source: "testdata/nonexistent.zip", target: "testdata", deleteSource: false},
			expectedFile: "testdata/istio.txt",
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := unZip(tt.args.source, tt.args.target, tt.args.deleteSource)
			if !tt.wantErr {
				require.NoError(t, err)
				_, err := os.Stat(tt.expectedFile)
				require.NoError(t, err)
				err = os.Remove(tt.expectedFile)
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				_, err := os.Stat(tt.expectedFile)
				require.Error(t, err)
			}
		})
	}
}

// MockClient is the mock client
type mockClient struct{}

// getGetFunc fetches the mock client's `Do` func
var getGetFunc func(url string) (*http.Response, error)

// Get is the mock client's `Get` func
func (m *mockClient) Get(url string) (*http.Response, error) {
	return getGetFunc(url)
}

func TestInstallation_downloadFile(t *testing.T) {
	type fields struct {
		Client HTTPClient
	}
	type args struct {
		filepath string
		filename string
		url      string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		getFunc func(url string) (*http.Response, error)
		wantErr bool
	}{
		{
			name:   "Download Istioctl",
			fields: fields{Client: &mockClient{}},
			args:   args{filepath: "tmp", filename: "mock_download.txt", url: "someUrl"},
			getFunc: func(url string) (*http.Response, error) {
				jsonBody := `{"name":"Istioctl","full_name":"Istioctl binary mock download","bin":{"data": "some binary"}}`
				r := io.NopCloser(bytes.NewReader([]byte(jsonBody)))
				return &http.Response{
					StatusCode: 200,
					Body:       r,
				}, nil
			},
			wantErr: false,
		},
		{
			name:   "404 - Not found",
			fields: fields{Client: &mockClient{}},
			args:   args{filepath: "tmp", filename: "mock_download.txt", url: "someUrl"},
			getFunc: func(url string) (*http.Response, error) {
				return &http.Response{
					StatusCode: 404,
				}, errors.New("404 - Not Found")
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			i := &Installation{
				Client: tt.fields.Client,
				logger: zap.NewNop().Sugar(),
			}
			getGetFunc = tt.getFunc
			err := i.downloadFile(tt.args.filepath, tt.args.filename, tt.args.url)
			if !tt.wantErr {
				require.NoError(t, err)
				tmpFile := path.Join(tt.args.filepath, tt.args.filename)
				_, err := os.Stat(tmpFile)
				require.NoError(t, err)
				err = os.Remove(tmpFile)
				require.NoError(t, err)
				err = os.Remove(tt.args.filepath)
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				tmpFile := path.Join(tt.args.filepath, tt.args.filename)
				_, err := os.Stat(tmpFile)
				require.Error(t, err)
			}
		})
	}
}
