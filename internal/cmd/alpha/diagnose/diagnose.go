package diagnose

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/diagnostics"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type diagnoseConfig struct {
	*cmdcommon.KymaConfig
	outputFormat types.Format
	outputPath   string
	verbose      bool
}

func NewDiagnoseCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := diagnoseConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "diagnose [flags]",
		Short: "Diagnose cluster health and configuration",
		Long:  "Use this command to quickly assess the health, configuration, and potential issues in your cluster for troubleshooting and support purposes.",
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(diagnose(&cfg))
		},
	}

	cmd.Flags().VarP(&cfg.outputFormat, "format", "f", "Output format (possible values: json, yaml)")
	cmd.Flags().StringVarP(&cfg.outputPath, "output", "o", "", "Path to the diagnostic output file. If not provided the output is printed to stdout")
	cmd.Flags().BoolVar(&cfg.verbose, "verbose", false, "Display verbose output, including error details during diagnostics collection")

	return cmd
}

func diagnose(cfg *diagnoseConfig) clierror.Error {
	if cfg.verbose {
		out.EnableDebug()
	}

	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	if cfg.outputPath != "" {
		out.Msgln("Collecting diagnostics data...")
	}

	diagnosticData := diagnostics.GetData(cfg.Ctx, client)
	err := render(&diagnosticData, cfg.outputPath, cfg.outputFormat)

	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to get diagnostic data"))
	}

	if cfg.outputPath != "" {
		out.Msgln("Done.")
	}

	return nil
}

func render(diagData *diagnostics.DiagnosticData, outputFilepath string, outputFormat types.Format) error {
	printer := out.Default

	if outputFilepath != "" {
		file, err := os.Create(outputFilepath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()

		printer = out.NewToWriter(file)
	}

	switch outputFormat {
	case types.JSONFormat:
		return renderJSON(printer, diagData)
	case types.YAMLFormat:
		return renderYAML(printer, diagData)
	case types.DefaultFormat:
		return renderYAML(printer, diagData)
	default:
		return fmt.Errorf("unexpected output.Format: %#v", outputFormat)
	}
}

func renderJSON(printer *out.Printer, data *diagnostics.DiagnosticData) error {
	obj, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	printer.Msg(string(obj))
	return nil
}

func renderYAML(printer *out.Printer, data *diagnostics.DiagnosticData) error {
	obj, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	printer.Msg(string(obj))
	return nil
}
