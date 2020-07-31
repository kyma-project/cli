package install

import (
	"testing"
	"time"

	"errors"

	trustMocks "github.com/kyma-project/cli/internal/trust/mocks"
	stepMocks "github.com/kyma-project/cli/pkg/step/mocks"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
)

// TestInstallFlags ensures that the provided command flags are stored in the options.
func TestInstallFlags(t *testing.T) {
	o := NewOptions(&cli.Options{})
	c := NewCmd(o)

	// test default flag values
	require.Equal(t, false, o.NoWait, "Default value for the noWait flag not as expected.")
	require.Equal(t, defaultDomain, o.Domain, "Default value for the domain flag not as expected.")
	require.Equal(t, "", o.TLSCert, "Default value for the tlsCert flag not as expected.")
	require.Equal(t, "", o.TLSKey, "Default value for the tlsKey flag not as expected.")
	require.Equal(t, DefaultKymaVersion, o.Source, "Default value for the source flag not as expected.")
	require.Equal(t, "", o.LocalSrcPath, "Default value for the src-path flag not as expected.")
	require.Equal(t, 1*time.Hour, o.Timeout, "Default value for the timeout flag not as expected.")
	require.Equal(t, "", o.Password, "Default value for the password flag not as expected.")
	require.Equal(t, []string([]string(nil)), o.OverrideConfigs, "Default value for the override flag not as expected.")
	require.Equal(t, "", o.ComponentsConfig, "Default value for the components flag not as expected.")
	require.Equal(t, 5, o.FallbackLevel, "Default value for the fallbackLevel flag not as expected.")
	require.Equal(t, "", o.CustomImage, "Default value for the custom-image flag not as expected.")

	// test passing flags
	err := c.ParseFlags([]string{
		"-n", "true",
		"-d", "fake-domain",
		"--tlsCert", "fake-cert",
		"--tlsKey", "fake-key",
		"-s", "test-registry/test-image:1",
		"--src-path", "fake/path/to/source",
		"--timeout", "100s",
		"-p", "fake-pwd",
		"-o", "fake/path/to/overrides",
		"-c", "fake/path/to/components",
		"--fallbackLevel", "7",
		"--custom-image", "test-registry/test-image:2",
	})
	require.NoError(t, err, "Parsing flags should not return an error")
	require.Equal(t, true, o.NoWait, "The parsed value for the noWait flag not as expected.")
	require.Equal(t, "fake-domain", o.Domain, "The parsed value for the domain flag not as expected.")
	require.Equal(t, "fake-cert", o.TLSCert, "The parsed value for the tlsCert flag not as expected.")
	require.Equal(t, "fake-key", o.TLSKey, "The parsed value for the tlsKey flag not as expected.")
	require.Equal(t, "test-registry/test-image:1", o.Source, "The parsed value for the source flag not as expected.")
	require.Equal(t, "fake/path/to/source", o.LocalSrcPath, "The parsed value for the src-path flag not as expected.")
	require.Equal(t, 100*time.Second, o.Timeout, "The parsed value for the timeout flag not as expected.")
	require.Equal(t, "fake-pwd", o.Password, "The parsed value for the password flag not as expected.")
	require.Equal(t, []string([]string{"fake/path/to/overrides"}), o.OverrideConfigs, "The parsed value for the override flag not as expected.")
	require.Equal(t, "fake/path/to/components", o.ComponentsConfig, "The parsed value for the components flag not as expected.")
	require.Equal(t, 7, o.FallbackLevel, "The parsed value for the fallbackLevel flag not as expected.")
	require.Equal(t, "test-registry/test-image:2", o.CustomImage, "The parsed value for the custom-image flag not as expected.")
}

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
