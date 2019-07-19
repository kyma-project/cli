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

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			// set desired helm home command mocked output
			helmCmd = exec.Command("echo", test.cmdOutput)
			home, err := Home()

			require.Equal(t, test.expected, home, test.description)
			require.Nil(t, err, test.description)
		})
	}
}
