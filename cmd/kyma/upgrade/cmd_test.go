package upgrade

import (
	"testing"
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
)

// TestUpgradeFlags ensures that the provided command flags are stored in the options.
func TestUpgradeFlags(t *testing.T) {
	t.Parallel()
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
