package junitxml

import (
	"encoding/xml"
	"fmt"
	"io"
	"time"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-project/cli/cmd/kyma/version"
	"github.com/pkg/errors"
)

//go:generate mockery -name=logsFetcher -output=automock -outpkg=automock -case=underscore

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
	report, err := c.generateReport(suite)
	if err != nil {
		return errors.Wrap(err, "while generating JUnit XML report")
	}

	if err := c.write(out, report); err != nil {
		return errors.Wrap(err, "while writing JUnit XML")
	}

	return nil
}

func (c *Creator) generateReport(suite *oct.ClusterTestSuite) (JUnitTestSuites, error) {
	var suiteTotalTime time.Duration
	// CompletionTime is not set when test suite is timed out
	if suite.Status.CompletionTime != nil {
		suiteTotalTime = suite.Status.CompletionTime.Sub(suite.Status.StartTime.Time)
	}

	tc, err := c.mapToTestCases(suite.Status.Results)
	if err != nil {
		return JUnitTestSuites{}, errors.Wrap(err, "while mapping test results into junit test cases")
	}

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
	}, nil
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

func (c *Creator) mapToTestCases(results []oct.TestResult) ([]JUnitTestCase, error) {
	var cases []JUnitTestCase

	for _, r := range results {
		switch r.Status {
		case oct.TestSucceeded:
			cases = append(cases, c.newSuccessJUnitTestCase(r))
		case oct.TestSkipped:
			cases = append(cases, c.newSkippedJUnitTestCase(r))
		case oct.TestFailed:
			jtc, err := c.newFailedJUnitTestCase(r)
			if err != nil {
				return nil, errors.Wrap(err, "while mapping octopus test result into JUnit test case")
			}
			cases = append(cases, jtc)
		case oct.TestUnknown:
			cases = append(cases, c.newUnknownJUnitTestCase(r))
		case oct.TestRunning:
			cases = append(cases, c.newRunningJUnitTestCase(r))
		}
	}

	return cases, nil
}

func (c *Creator) newSuccessJUnitTestCase(tc oct.TestResult) JUnitTestCase {
	var totalExecutionTime time.Duration
	for _, e := range tc.Executions {
		// CompletionTime is not set when test case is timed out or still running
		if e.CompletionTime != nil {
			totalExecutionTime += e.CompletionTime.Sub(e.StartTime.Time)
		}
	}

	return JUnitTestCase{
		Classname: "octopus",
		Name:      fmt.Sprintf("[testing] %s/%s", tc.Namespace, tc.Name),
		Time:      c.formatDurationAsSeconds(totalExecutionTime),
	}
}

func (c *Creator) newSkippedJUnitTestCase(r oct.TestResult) JUnitTestCase {
	return JUnitTestCase{
		Classname:   "octopus",
		Name:        fmt.Sprintf("[testing] %s/%s", r.Namespace, r.Name),
		SkipMessage: &JUnitSkipMessage{Message: "Test was skipped."},
	}
}

func (c *Creator) newFailedJUnitTestCase(r oct.TestResult) (JUnitTestCase, error) {
	logs, err := c.logsFetcher.Logs(r)
	if err != nil {
		return JUnitTestCase{}, errors.Wrapf(err, "while fetching logs for %s test", r.Name)
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
		Name:      fmt.Sprintf("[testing] %s/%s (executions: %d)", r.Namespace, r.Name, len(r.Executions)),
		Time:      c.formatDurationAsSeconds(totalExecutionTime),
		Failure: &JUnitFailure{
			Message:  "Failed",
			Contents: logs,
		},
	}, nil
}

func (c *Creator) newRunningJUnitTestCase(r oct.TestResult) JUnitTestCase {
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
		Name:      fmt.Sprintf("[testing] %s/%s (executions: %d)", r.Namespace, r.Name, len(r.Executions)),
		Time:      c.formatDurationAsSeconds(totalExecutionTime),
		Failure: &JUnitFailure{
			Message:  "Failed",
			Contents: fmt.Sprintf("Test was marked as failed due to too long Running status: %s\n", logs),
		},
	}
}

func (c *Creator) newUnknownJUnitTestCase(r oct.TestResult) JUnitTestCase {
	return JUnitTestCase{
		Classname: "octopus",
		Name:      fmt.Sprintf("[testing] %s/%s", r.Namespace, r.Name),
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
