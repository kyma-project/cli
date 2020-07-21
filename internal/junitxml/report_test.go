package junitxml_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-project/cli/internal/junitxml"
	"github.com/kyma-project/cli/internal/junitxml/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gotest.tools/golden"
	"sigs.k8s.io/yaml"
)

// TestWrite tests that proper JUnit XML document is written to given output.
//
// This test is based on golden file.
// If the `-test.update-golden` flag is set then the actual content is written
// to the golden file.
//
// Example:
//   go test ./internal/junitxml/... -v -test.update-golden
func TestWriteJUnitXMLReport(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Test Suite Failed",
		},
		{
			name: "TestSuite is still running but timeout occur",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// given
			fixCTS := getCTSFromTestData(t)

			mockedLogsFetcher := &mocks.LogsFetcher{}
			defer mockedLogsFetcher.AssertExpectations(t)

			for _, result := range fixCTS.Status.Results {
				// expecting get logs call for each running and failed test results
				if result.Status == oct.TestRunning || result.Status == oct.TestFailed {
					logs := fmt.Sprintf("Faked logs for execution: %s", result.Name)
					mockedLogsFetcher.On("Logs", mock.Anything).
						Return(logs, nil).Once()
				}
			}

			creator := junitxml.NewCreator(mockedLogsFetcher)

			gotOutput := new(bytes.Buffer)

			// when
			err := creator.Write(gotOutput, &fixCTS)

			// then
			require.NoError(t, err)
			golden.Assert(t, gotOutput.String(), t.Name()+".golden.xml")
		})
	}
}

func getCTSFromTestData(t *testing.T) oct.ClusterTestSuite {
	raw, err := ioutil.ReadFile(path.Join("testdata", t.Name()+".input.yaml"))
	require.NoError(t, err)

	cts := oct.ClusterTestSuite{}
	err = yaml.Unmarshal(raw, &cts)
	require.NoError(t, err)

	return cts
}
