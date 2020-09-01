package minikube

import (
	"fmt"
	"testing"

	"github.com/kyma-project/cli/pkg/step"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
)

// TestProvisionGKEFlags ensures that the provided command flags are stored in the options.
func TestProvisionMinikubeFlags(t *testing.T) {
	o := NewOptions(&cli.Options{})
	c := NewCmd(o)

	// test passing flags
	err := c.ParseFlags([]string{
		"--cpus", "6",
		"--memory", "4096",
		"--vm-driver", "kvm",
		"--profile", "fooProfile",
	})
	require.NoError(t, err, "Parsing flags should not return an error")
	require.Equal(t, "6", o.CPUS)
	require.Equal(t, "4096", o.Memory)
	require.Equal(t, "kvm", o.VMDriver)
	require.Equal(t, "fooProfile", o.Profile)
	if allowVPNSock {
		require.Equal(t, false, o.UseVPNKitSock)
	}

}

func TestCheckRequirements(t *testing.T) {

	cases := []struct {
		name        string
		shouldFail  bool
		expectedErr string
		op          Options
	}{
		{
			name:        "VM Driver not supported",
			shouldFail:  true,
			expectedErr: "Specified VMDriver 'fooDriver' is not supported by Minikube",
			op: Options{
				VMDriver: "fooDriver",
			},
		},
		{
			name:        "VM Driver requires hypervVirtualSwitch",
			shouldFail:  true,
			expectedErr: "Specified VMDriver 'hyperv' requires the --hypervVirtualSwitch option",
			op: Options{
				VMDriver: "hyperv",
			},
		},
	}
	var step step.Factory
	s := step.NewStep("checking requirements")
	for _, c := range cases {
		opts := &c.op
		cmd := command{
			opts: opts,
		}

		err := cmd.checkRequirements(s)
		if c.expectedErr != "" {
			require.EqualError(t, err, c.expectedErr, fmt.Sprintf("Test Case: %s", c.name))
		}

	}
}
