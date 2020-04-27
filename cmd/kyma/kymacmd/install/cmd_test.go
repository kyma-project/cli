package install

import (
	"testing"

	"errors"

	trustMocks "github.com/kyma-project/cli/internal/trust/mocks"
	stepMocks "github.com/kyma-project/cli/pkg/step/mocks"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
)

func TestImportCertificate(t *testing.T) {
	cases := []struct {
		// params
		name        string
		description string
		cert        trustMocks.Certifier
		wait        bool
		// results
		success            bool
		stopped            bool
		expectedStepStatus []string
		expectedStepInfos  []string
		expectedStepErrors []string
		expectedErr        error
	}{
		{
			name:        "Certificate import",
			description: "Imports the correct certificate",
			cert: trustMocks.Certifier{
				Crt: "Hi, I am a fake certificate!",
			},
			wait:               true,
			success:            true,
			stopped:            false,
			expectedStepStatus: []string{"Kyma root certificate imported"},
			expectedErr:        nil,
		},
		{
			name:        "Certificate retrieval failed",
			description: "Not possible to retrieve the certificate",
			cert: trustMocks.Certifier{
				Crt: "",
			},
			wait:        true,
			success:     false,
			stopped:     false,
			expectedErr: errors.New("Could not retrieve the certificate"),
		},
		{
			name:        "No Wait",
			description: "Certificate not imported due to not waiting for Kyma installation",
			cert: trustMocks.Certifier{
				Crt: "", // empty because certificate retrieval should not be attempted
			},
			wait:               false,
			success:            false,
			stopped:            false,
			expectedStepErrors: []string{"Manual OS-specific instructions for certificate import"},
			expectedErr:        nil,
		},
	}

	cmd := command{
		opts: NewOptions(cli.NewOptions()),
	}

	mockStep := &stepMocks.Step{}
	cmd.CurrentStep = mockStep

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd.opts.NoWait = !tc.wait
			err := cmd.importCertificate(tc.cert)

			require.Equal(t, tc.expectedErr, err, "Error not as expected")
			require.Equal(t, tc.success, mockStep.IsSuccessful(), "Import certificate step must be successful")
			require.Equal(t, tc.stopped, mockStep.IsStopped(), "Import certificate step must not be stopped")
			require.Equal(t, tc.expectedStepStatus, mockStep.Statuses(), "Status messages not as expected")
			require.Equal(t, tc.expectedStepInfos, mockStep.Infos(), "Logged info messages not as expected")
			require.Equal(t, tc.expectedStepErrors, mockStep.Errors(), "Logged error messages not as expected")

			mockStep.Reset()
		})
	}

}
