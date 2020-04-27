package logs

import (
	"bufio"
	"fmt"
	"github.com/kyma-project/cli/internal/logs"
	"strings"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-project/cli/cmd/kyma/kymacmd/test"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	defaultIgnoredContainers = []string{"istio-init", "istio-proxy", "manager"}
	defaultLogsInStatus      = string(oct.TestFailed)
)

type command struct {
	opts *Options
	cli.Command
}

func NewCmd(o *Options) *cobra.Command {
	cmd := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:   "logs <test-suite-1> <test-suite-2> ... <test-suite-N>",
		Short: "Shows the logs of tests Pods for a given test suite.",
		Long: `Use this command to display logs of a test executed for a given test suite. By default, the command displays logs for failed tests, but you can change this behavior using the "test-status" flag. 

To print the status of specific test cases, run ` + "`kyma test logs testSuiteOne testSuiteTwo`" + `.
Provide at least one test suite name.
`,

		RunE: func(_ *cobra.Command, args []string) error { return cmd.Run(args) },
	}

	cobraCmd.Flags().StringVar(&o.InStatus, "test-status", defaultLogsInStatus, "Displays logs coming only from testing Pods with a given status.")
	cobraCmd.Flags().StringSliceVar(&o.IngoredContainers, "ignored-containers", defaultIgnoredContainers, "Container names which are ignored when fetching logs from testing Pods. Takes comma-separated list.")

	return cobraCmd
}

func (cmd *command) Run(args []string) error {
	if err := cmd.validateFlags(); err != nil {
		return err
	}

	if len(args) < 1 {
		return fmt.Errorf("Test suite name required")
	}

	var err error
	cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath)
	if err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure that your kubeconfig is valid.")
	}

	logsStep := cmd.NewStep("Fetching logs")
	logsStep.Start()

	testsList, err := test.ListTestSuitesByName(cmd.K8s.Octopus(), args)
	if err != nil {
		return errors.Wrap(err, "unable to list test suites")
	}

	if len(testsList) == 0 {
		logsStep.LogInfof("No test suites found for names: %q", strings.Join(args, ", "))
		return nil
	}

	results := filterResultsByStatus(testsList, cmd.opts.InStatus)
	if len(results) == 0 {
		logsStep.LogInfof("No logs to fetch for testing pods in status %q", oct.TestStatus(cmd.opts.InStatus))
		return nil
	}

	logsFetcher := logs.NewFetcherForTestingPods(cmd.K8s.Static().CoreV1(), cmd.opts.IngoredContainers)
	for _, result := range results {
		content, err := logsFetcher.Logs(result)
		if err != nil {
			logsStep.Failure()
			return errors.Wrap(err, "while fetching logs")
		}

		scanner := bufio.NewScanner(strings.NewReader(content))
		for scanner.Scan() {
			logsStep.LogInfo(scanner.Text())
		}
	}

	return nil
}

func filterResultsByStatus(testsList []oct.ClusterTestSuite, status string) []oct.TestResult {
	var results []oct.TestResult
	for _, t := range testsList {
		for _, r := range t.Status.Results {
			if r.Status == oct.TestStatus(status) {
				results = append(results, r)
			}
		}
	}

	return results
}

func (cmd *command) validateFlags() error {
	if err := validateStatusOpt(cmd.opts.InStatus); err != nil {
		return err
	}

	return nil
}

// validate validates if given input parameters exist as v1alpha1.TestStatus enum
// TODO(mszostok): move to the octopus because this project is proper owner of that logic.
func validateStatusOpt(in string) error {
	allowedValues := []string{
		string(oct.TestScheduled),
		string(oct.TestRunning),
		string(oct.TestUnknown),
		string(oct.TestFailed),
		string(oct.TestSucceeded),
		string(oct.TestSkipped),
	}

	for _, v := range allowedValues {
		if in == v {
			return nil
		}
	}

	return fmt.Errorf(`invalid argument %q for "--status" flag: allowed values are: %s`, in, strings.Join(allowedValues, ", "))
}
