package version

import (
	"bytes"
	"testing"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/metadata"
	"github.com/kyma-project/cli/cmd/kyma/version"
	"github.com/stretchr/testify/assert"
)

func TestPrintVersion(t *testing.T) {
	tests := []struct {
		name            string
		clientOnly      bool
		clientVersion   string
		clusterMetadata *metadata.KymaMetadata
		want            string
	}{
		{
			name:       "client version only (empty)",
			clientOnly: true,
			clusterMetadata: &metadata.KymaMetadata{
				Version: "1.19",
			},
			want: "Kyma CLI version: N/A\n",
		},
		{
			name:          "client version only (non empty)",
			clientVersion: "1.20",
			clientOnly:    true,
			clusterMetadata: &metadata.KymaMetadata{
				Version: "1.19",
			},
			want: "Kyma CLI version: 1.20\n",
		},
		{
			name:          "client and server version (cluster metadata contains version)",
			clientVersion: "1.20",
			clientOnly:    false,
			clusterMetadata: &metadata.KymaMetadata{
				Version: "",
			},
			want: "Kyma CLI version: 1.20\nKyma cluster version: N/A\n",
		},
		{
			name:          "client and server version (cluster metadata contains no version)",
			clientVersion: "1.20",
			clientOnly:    false,
			clusterMetadata: &metadata.KymaMetadata{
				Version: "1.19",
			},
			want: "Kyma CLI version: 1.20\nKyma cluster version: 1.19\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			version.Version = tc.clientVersion

			printVersion(buf, tc.clientOnly, tc.clusterMetadata)
			assert.Equal(t, tc.want, buf.String())
		})
	}

}
