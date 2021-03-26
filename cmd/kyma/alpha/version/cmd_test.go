package version

import (
	"bytes"
	"testing"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
	"github.com/kyma-project/cli/cmd/kyma/version"
	"github.com/stretchr/testify/assert"
)

func TestPrintVersion(t *testing.T) {
	tests := []struct {
		name           string
		clientOnly     bool
		versionDetails bool
		clientVersion  string
		kymaVersionSet *helm.KymaVersionSet
		want           string
	}{
		{
			name:       "client version only (empty)",
			clientOnly: true,
			kymaVersionSet: &helm.KymaVersionSet{
				Versions: []*helm.KymaVersion{
					&helm.KymaVersion{
						Version: "1.19",
					},
				},
			},
			want: "Kyma CLI version: N/A\n",
		},
		{
			name:          "client version only (non empty)",
			clientVersion: "1.20",
			clientOnly:    true,
			kymaVersionSet: &helm.KymaVersionSet{
				Versions: []*helm.KymaVersion{
					&helm.KymaVersion{
						Version: "1.19",
						Profile: "evaluation",
					},
				},
			},
			want: "Kyma CLI version: 1.20\n",
		},
		{
			name:           "client and server version (no Kyma installed)",
			clientVersion:  "1.20",
			clientOnly:     false,
			kymaVersionSet: &helm.KymaVersionSet{},
			want:           "Kyma CLI version: 1.20\nKyma cluster versions: N/A\n",
		},
		{
			name:          "client and server version (1 Kyma version installed)",
			clientVersion: "1.20",
			clientOnly:    false,
			kymaVersionSet: &helm.KymaVersionSet{
				Versions: []*helm.KymaVersion{
					&helm.KymaVersion{
						Version: "1.19",
					},
				},
			},
			want: "Kyma CLI version: 1.20\nKyma cluster versions: 1.19\n",
		},
		{
			name:          "client and server version (2 Kyma versions installed)",
			clientVersion: "1.20",
			clientOnly:    false,
			kymaVersionSet: &helm.KymaVersionSet{
				Versions: []*helm.KymaVersion{
					&helm.KymaVersion{
						Version: "1.19",
						Profile: "evaluation",
					},
					&helm.KymaVersion{
						Version: "1.23",
					},
				},
			},
			want: "Kyma CLI version: 1.20\nKyma cluster versions: 1.19, 1.23\n",
		},
		{
			name:           "client and server version with details (2 Kyma versions installed)",
			clientVersion:  "1.20",
			clientOnly:     false,
			versionDetails: true,
			kymaVersionSet: &helm.KymaVersionSet{
				Versions: []*helm.KymaVersion{
					&helm.KymaVersion{
						Version:      "1.19",
						Profile:      "evaluation",
						CreationTime: 1616070314, //Thursday, 18-Mar-21 13:25:14 CET
						Components: []*helm.KymaComponentMetadata{
							&helm.KymaComponentMetadata{
								Name:      "comp1",
								Namespace: "ns1",
							},
							&helm.KymaComponentMetadata{
								Name:      "comp2",
								Namespace: "ns1",
							},
						},
					},
					&helm.KymaVersion{
						Version:      "1.20",
						Profile:      "",
						CreationTime: 1616070315, //Thursday, 18-Mar-21 13:25:15 CET
						Components: []*helm.KymaComponentMetadata{
							&helm.KymaComponentMetadata{
								Name:      "comp1",
								Namespace: "ns1",
							},
							&helm.KymaComponentMetadata{
								Name:      "comp3",
								Namespace: "ns1",
							},
						},
					},
				},
			},
			want: `Kyma CLI version: 1.20
Kyma cluster versions: 1.19, 1.20
-----------------
Kyma cluster version: 1.19
Deployed at: Thursday, 18-Mar-21 13:25:14 CET
Profile: evaluation
Components: comp1, comp2
-----------------
Kyma cluster version: 1.20
Deployed at: Thursday, 18-Mar-21 13:25:15 CET
Profile: default
Components: comp1, comp3
`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			command := command{
				opts: &Options{
					ClientOnly: tc.clientOnly,
				},
			}

			buf := new(bytes.Buffer)
			version.Version = tc.clientVersion
			command.printCliVersion(buf)
			if !tc.clientOnly {
				command.printKymaVersion(buf, tc.kymaVersionSet)
			}
			if tc.versionDetails {
				command.printKymaVersionDetails(buf, tc.kymaVersionSet)
			}
			assert.Equal(t, tc.want, buf.String())
		})
	}

}
