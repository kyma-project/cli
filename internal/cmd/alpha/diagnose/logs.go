package diagnose

import (
	"context"
	"fmt"
	"os"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/diagnostics"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
)

type diagnoseLogsConfig struct {
	*cmdcommon.KymaConfig
	outputFormat types.Format
	outputPath   string
	verbose      bool
	modules      []string
	since        time.Duration
	lines        int64
	timeout      time.Duration
}

func NewDiagnoseLogsCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := diagnoseLogsConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "logs [flags]",
		Short: "Aggregates error logs from Pods belonging to the added Kyma modules",
		Long: "Collects and aggregates recent error-level container logs for Kyma modules to help with rapid troubleshooting.\n\n" +
			"EXAMPLES:\n" +
			"  # Collect last 200 lines (default) from all enabled modules\n" +
			"  kyma alpha diagnose logs --lines 200\n\n" +
			"  # Collect error logs from the last 15 minutes for all enabled modules\n" +
			"  kyma alpha diagnose logs --since 15m\n\n" +
			"  # Restrict to specific modules (repeat --module) and increase line count\n" +
			"  kyma alpha diagnose logs --module serverless --module api-gateway --lines 500\n\n" +
			"  # Time-based collection for one module, output as JSON to a file\n" +
			"  kyma alpha diagnose logs --module serverless --since 30m --format json --output serverless-errors.json\n\n" +
			"  # Collect with verbose output and shorter timeout (useful in CI)\n" +
			"  kyma alpha diagnose logs --since 10m --timeout 10s --verbose\n\n" +
			"  # Use lines as a deterministic cap when time window would be too large\n" +
			"  kyma alpha diagnose logs --lines 1000\n\n" +
			"NOTE: --since takes precedence over --lines when both are provided; use only one for clarity.",
		PreRun: func(cmd *cobra.Command, args []string) {
			clierror.Check(flags.Validate(cmd.Flags(),
				flags.MarkExclusive("lines", "since"),
			))
		},
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(diagnoseLogs(&cfg))
		},
	}

	cmd.Flags().VarP(&cfg.outputFormat, "format", "f", "Output format (possible values: json, yaml)")
	cmd.Flags().StringVarP(&cfg.outputPath, "output", "o", "", "Path to the diagnostic output file. If not provided the output is printed to stdout")
	cmd.Flags().BoolVar(&cfg.verbose, "verbose", false, "Display verbose output, including error details during diagnostics collection")
	cmd.Flags().StringSliceVar(&cfg.modules, "module", []string{}, "Restrict to specific module(s). Can be used multiple times. When no value is present then logs from all Kyma CR modules are gathered")
	cmd.Flags().DurationVar(&cfg.since, "since", 0, "Log time range (e.g., 10m, 1h, 30s)")
	cmd.Flags().Int64Var(&cfg.lines, "lines", 200, "Max lines per container")
	cmd.Flags().DurationVar(&cfg.timeout, "timeout", 30*time.Second, "Timeout for log collection operations")

	return cmd
}

func diagnoseLogs(cfg *diagnoseLogsConfig) clierror.Error {
	if cfg.verbose {
		out.EnableVerbose()
	}

	kubeClient, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	ctx, cancel := context.WithTimeout(cfg.Ctx, cfg.timeout)
	defer cancel()

	modules, err := selectModules(ctx, kubeClient, cfg.modules)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to select modules for log collection"))
	}

	logOpts := diagnostics.LogOptions{Since: cfg.since, Lines: cfg.lines}
	collector := diagnostics.NewModuleLogsCollector(kubeClient, modules, logOpts)
	moduleLogs := collector.Run(ctx)

	if err := renderLogs(moduleLogs.Logs, cfg.outputPath, cfg.outputFormat); err != nil {
		return clierror.Wrap(err, clierror.New("failed to render logs to output"))
	}

	return nil
}

func selectModules(ctx context.Context, client kube.Client, modules []string) ([]string, error) {
	if len(modules) > 0 {
		return modules, nil
	}

	defaultKyma, err := client.Kyma().GetDefaultKyma(ctx)
	if err != nil && !apierrors.IsNotFound(err) {
		return []string{}, fmt.Errorf("failed to get default Kyma CR from the target Kyma environment: %w", err)
	}

	if apierrors.IsNotFound(err) {
		return []string{}, nil
	}

	kymaModules := []string{}
	for _, module := range defaultKyma.Spec.Modules {
		kymaModules = append(kymaModules, module.Name)
	}

	return kymaModules, nil
}

func renderLogs(moduleLogs any, outputPath string, outputFormat types.Format) error {
	printer := out.Default

	if outputPath != "" {
		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()

		printer = out.NewToWriter(file)
	}

	switch outputFormat {
	case types.JSONFormat:
		return renderJSON(printer, moduleLogs)
	case types.YAMLFormat, types.DefaultFormat:
		return renderYAML(printer, moduleLogs)
	default:
		return fmt.Errorf("unexpected output.Format: %v", outputFormat)
	}
}
