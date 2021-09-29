package istio

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
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
		name    string
		fields  fields
		expectedVersion string
		wantErr bool
	}{
		{name: "Fetch Istio Version", fields: fields{IstioChartPath: "mock/Chart.yaml"}, expectedVersion: "1.11.2", wantErr: false},
		{name: "Istio Chart not existing", fields: fields{IstioChartPath: "mock/nonExisting.yaml"}, expectedVersion: "", wantErr: true},
		{name: "Corrupted Istio Chart", fields: fields{IstioChartPath: "mock/corruptedChart.yaml"}, expectedVersion: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Installation{
				WorkspacePath:  tt.fields.WorkspacePath,
				IstioChartPath: tt.fields.IstioChartPath,
				istioVersion:   tt.fields.istioVersion,
				osExt:          tt.fields.osExt,
				binName:        tt.fields.binName,
				winBinName:     tt.fields.winBinName,
			}
			err := i.getIstioVersion()
			if tt.wantErr  {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expectedVersion, i.istioVersion)
		})
	}
}

func TestInstallation_checkIfExists(t *testing.T) {
	type fields struct {
		binPath        string
	}
	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		{name: "File exists", fields: fields{binPath: "mock/Chart.yaml"}, want: true, wantErr: false},
		{name: "File does not exist", fields: fields{binPath: "mock/nonexistent"}, want: false, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Installation{
				binPath:        tt.fields.binPath,
			}
			got, err := i.checkIfExists()
			if tt.wantErr  {
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
		{name: "unGzip File", args: args{source: "mock/istio_mock.tar.gz", target: "mock/istio.tar", deleteSource: false}, wantErr: false},
		{name: "File does not exist", args: args{source: "mock/nonexistent.tar.gz", target: "mock/istio.tar", deleteSource: false}, wantErr: true},
	}
	for _, tt := range tests {
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
		name    string
		args    args
		expectedFile string
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "unTar File", args: args{source: "mock/istio_mock.tar", target: "mock", deleteSource: false}, expectedFile: "mock/istio.txt", wantErr: false},
		{name: "File does not exist", args: args{source: "mock/nonexistent.tar", target: "mock", deleteSource: false}, expectedFile: "mock/istio.txt", wantErr: true},
	}
	for _, tt := range tests {
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