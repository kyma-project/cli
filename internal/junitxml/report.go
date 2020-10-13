package junitxml

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/pkg/errors"

	"github.com/kyma-project/cli/cmd/kyma/version"
)

//go:generate mockery --name logsFetcher --structname LogsFetcher

// logsFetcher allows you to fetch logs from the testing pods
type logsFetcher interface {
	Logs(result oct.TestResult) (string, error)
}

// Creator provides functionality for creating JUnit XML report for given test suite
type Creator struct {
	logsFetcher logsFetcher
}

// NewCreator returns new instance of the Creator
func NewCreator(logsFetcher logsFetcher) *Creator {
	return &Creator{logsFetcher: logsFetcher}
}

// Write creates an XML document for suite and writes it to out.
func (c *Creator) Write(out io.Writer, suite *oct.ClusterTestSuite) error {
	report := c.generateReport(suite)

	if err := c.write(out, report); err != nil {
		return errors.Wrap(err, "while writing JUnit XML")
	}

	return nil
}

func (c *Creator) generateReport(suite *oct.ClusterTestSuite) JUnitTestSuites {
	var suiteTotalTime time.Duration
	// CompletionTime is not set when test suite is timed out
	if suite.Status.CompletionTime != nil {
		suiteTotalTime = suite.Status.CompletionTime.Sub(suite.Status.StartTime.Time)
	}

	tc := c.mapToTestCases(suite.Status.Results, suite.Name)

	junitSuite := JUnitTestSuite{
		Name:       suite.Name,
		Tests:      len(suite.Status.Results),
		Time:       c.formatDurationAsSeconds(suiteTotalTime),
		Properties: c.packageProperties(),
		TestCases:  tc,
		Failures:   c.getNumberOfFailedTests(suite),
	}

	return JUnitTestSuites{
		Suites: []JUnitTestSuite{junitSuite},
	}
}
func (c *Creator) formatDurationAsSeconds(d time.Duration) string {
	return fmt.Sprintf("%f", d.Seconds())
}

func (c *Creator) getNumberOfFailedTests(testSuite *oct.ClusterTestSuite) int {
	result := 0
	for _, t := range testSuite.Status.Results {
		if t.Status == oct.TestFailed {
			result++
		}
	}
	return result
}

func (c *Creator) cliVersion() string {
	if version.Version == "" {
		return "N/A"
	}
	return version.Version
}

func (c *Creator) packageProperties() []JUnitProperty {
	return []JUnitProperty{
		{Name: "kyma.cli.version", Value: c.cliVersion()},
	}
}

func (c *Creator) mapToTestCases(results []oct.TestResult, suiteName string) []JUnitTestCase {
	var cases []JUnitTestCase

	for _, r := range results {
		switch r.Status {
		case oct.TestSucceeded:
			cases = append(cases, c.newSuccessJUnitTestCase(r, suiteName))
		case oct.TestSkipped:
			cases = append(cases, c.newSkippedJUnitTestCase(r, suiteName))
		case oct.TestFailed:
			cases = append(cases, c.newFailedJUnitTestCase(r, suiteName))
		case oct.TestUnknown:
			cases = append(cases, c.newUnknownJUnitTestCase(r, suiteName))
		case oct.TestRunning:
			cases = append(cases, c.newRunningJUnitTestCase(r, suiteName))
		}
	}

	return cases
}

func (c *Creator) newSuccessJUnitTestCase(tc oct.TestResult, suiteName string) JUnitTestCase {
	var totalExecutionTime time.Duration
	for _, e := range tc.Executions {
		// CompletionTime is not set when test case is timed out or still running
		if e.CompletionTime != nil {
			totalExecutionTime += e.CompletionTime.Sub(e.StartTime.Time)
		}
	}

	return JUnitTestCase{
		Classname: "octopus",
		Name:      fmt.Sprintf("[%s] %s/%s", suiteName, tc.Namespace, tc.Name),
		Time:      c.formatDurationAsSeconds(totalExecutionTime),
	}
}

func (c *Creator) newSkippedJUnitTestCase(r oct.TestResult, suiteName string) JUnitTestCase {
	return JUnitTestCase{
		Classname:   "octopus",
		Name:        fmt.Sprintf("[%s] %s/%s", suiteName, r.Namespace, r.Name),
		SkipMessage: &JUnitSkipMessage{Message: "Test was skipped."},
	}
}

func (c *Creator) newFailedJUnitTestCase(r oct.TestResult, suiteName string) JUnitTestCase {
	var totalExecutionTime time.Duration
	for _, e := range r.Executions {
		// CompletionTime is not set when test case is timed out or still running
		if e.CompletionTime != nil {
			totalExecutionTime += e.CompletionTime.Sub(e.StartTime.Time)
		}
	}

	logs, err := c.logsFetcher.Logs(r)
	if err != nil {
		return JUnitTestCase{Classname: "octopus",
			Name: fmt.Sprintf("[%s] %s/%s (executions: %d)", suiteName, r.Namespace, r.Name, len(r.Executions)),
			Time: c.formatDurationAsSeconds(totalExecutionTime),
			Failure: &JUnitFailure{
				Message:  "Failed",
				Contents: fmt.Sprintf("Cannot display output for %s test. Kyma CLI failed to fetch logs during report generation: %s", r.Name, err),
			},
		}
	}

	return JUnitTestCase{Classname: "octopus",
		Name: fmt.Sprintf("[%s] %s/%s (executions: %d)", suiteName, r.Namespace, r.Name, len(r.Executions)),
		Time: c.formatDurationAsSeconds(totalExecutionTime),
		Failure: &JUnitFailure{
			Message:  "Failed",
			Contents: strings.ToValidUTF8(logs, ""),
		},
	}
}

func (c *Creator) newRunningJUnitTestCase(r oct.TestResult, suiteName string) JUnitTestCase {
	logs, err := c.logsFetcher.Logs(r)
	if err != nil {
		logs = fmt.Sprintf("Cannot fetch logs, got error: %v", err)
	}

	var totalExecutionTime time.Duration
	for _, e := range r.Executions {
		// CompletionTime is not set when test case is timed out or still running
		if e.CompletionTime != nil {
			totalExecutionTime += e.CompletionTime.Sub(e.StartTime.Time)
		}
	}

	return JUnitTestCase{
		Classname: "octopus",
		Name:      fmt.Sprintf("[%s] %s/%s (executions: %d)", suiteName, r.Namespace, r.Name, len(r.Executions)),
		Time:      c.formatDurationAsSeconds(totalExecutionTime),
		Failure: &JUnitFailure{
			Message:  "Failed",
			Contents: fmt.Sprintf("Test was marked as failed due to too long Running status: %s\n", logs),
		},
	}
}

func (c *Creator) newUnknownJUnitTestCase(r oct.TestResult, suiteName string) JUnitTestCase {
	return JUnitTestCase{
		Classname: "octopus",
		Name:      fmt.Sprintf("[%s] %s/%s", suiteName, r.Namespace, r.Name),
		Failure: &JUnitFailure{
			Message:  "Failed",
			Contents: "Test status is marked as 'Unknown' by Octopus runner.",
		},
	}
}

func (c *Creator) write(out io.Writer, suites JUnitTestSuites) error {
	doc, err := xml.MarshalIndent(suites, "", "\t")
	if err != nil {
		return err
	}
	_, err = out.Write([]byte(xml.Header))
	if err != nil {
		return err
	}
	_, err = out.Write(doc)
	return err
}
