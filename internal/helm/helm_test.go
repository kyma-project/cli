package helm

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

//fake helm folder
var helmDir string

func TestMain(m *testing.M) {
	// setup
	var err error
	if helmDir, err = os.Getwd(); err != nil {
		log.Fatal(err)
	}
	helmDir = fmt.Sprintf("%s/.helm", helmDir)

	// run all tests
	os.Exit(m.Run())
}

func TestHome(t *testing.T) {
	defer os.Remove(helmDir)

	cases := []struct {
		name        string
		description string
		cmdOutput   string // mocked output of the helm home command
		expected    string
	}{
		{
			name:        "Helm home does not exist",
			description: "Helm Home without helm folder.",
			cmdOutput:   helmDir,
			expected:    helmDir,
		},
		{
			name:        "Helm home already exists",
			description: "Helm Home with existing helm folder.",
			cmdOutput:   helmDir,
			expected:    helmDir,
		},
		{
			name:        "Newline scraping",
			description: "Helm Home with new line characters.",
			cmdOutput:   fmt.Sprintf("%s\n\n\n\n\n\n\n\n\n\n\n\n", helmDir),
			expected:    helmDir,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// set desired helm home command mocked output
			helmHomeCmd = exec.Command("echo", tc.cmdOutput) //nolint:gosec
			home, err := Home()

			require.Equal(t, tc.expected, home, tc.description)
			require.Nil(t, err, tc.description)
		})
	}
}

func TestSupportedVersion(t *testing.T) {
	cases := []struct {
		name        string
		description string
		cmdOutput   string // mocked output of the helm version command
		supported   bool
	}{
		{
			name:        "Supported Helm version",
			description: "Helm version is supported. (i.e v2.x.x)",
			cmdOutput:   "Client: v2.1.16",
			supported:   true,
		},
		{
			name:        "Unsupported Helm version",
			description: "Helm version is not supported. (e.g. v3.x.x)",
			cmdOutput:   "v3.0.2",
			supported:   false,
		},
		{
			name:        "Empty string",
			description: "Helm version returns empty string.",
			cmdOutput:   "",
			supported:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// set desired helm home command mocked output
			helmVersionCmd = exec.Command("echo", tc.cmdOutput) //nolint:gosec
			supported, err := SupportedVersion()

			require.Equal(t, tc.supported, supported, tc.description)
			require.Nil(t, err, tc.description)
		})
	}
}
