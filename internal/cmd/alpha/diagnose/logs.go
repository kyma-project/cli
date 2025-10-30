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
		Short: "Aggregate error logs from pods belonging to enabled Kyma Modules",
		Long:  "Some better long description",
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
	cmd.Flags().StringSliceVar(&cfg.modules, "module", []string{}, "Restrict to specific module(s). Can be used multiple times")
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

	logsCollector := diagnostics.NewModuleLogsCollector(kubeClient)

	var moduleLogs diagnostics.ModuleLogs

	if cfg.since > 0 {
		moduleLogs = logsCollector.RunSince(ctx, modules, cfg.since)
	} else if cfg.lines > 0 {
		moduleLogs = logsCollector.RunLast(ctx, modules, cfg.lines)
	} else {
		return clierror.New("either --since or --lines flag must be specified")
	}

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
			return fmt.Errorf("failed to create output file")
		}
		defer file.Close()

		printer = out.NewToWriter(file)
	}

	switch outputFormat {
	case types.JSONFormat:
		return renderJSON(printer, moduleLogs)
	case types.YAMLFormat:
		return renderYAML(printer, moduleLogs)
	case types.DefaultFormat:
		return renderYAML(printer, moduleLogs)
	default:
		return fmt.Errorf("unexpected output.Format: %#v", outputFormat)
	}
}
