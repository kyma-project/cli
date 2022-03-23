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
	t.Parallel()
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
}

func TestCheckRequirements(t *testing.T) {
	t.Parallel()
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
			expectedErr: "Specified VMDriver 'hyperv' requires the --hyperv-virtual-switch option",
			op: Options{
				VMDriver: "hyperv",
			},
		},
		{
			name:        "--docker-ports require VM Driver docker",
			shouldFail:  true,
			expectedErr: "docker-ports flag is applicable only for VMDriver 'docker'",
			op: Options{
				VMDriver:    "hyperkit",
				DockerPorts: []string{"8080:8081"},
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
