package diagnose

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/diagnostics"
	"github.com/kyma-project/cli.v3/internal/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type diagnoseConfig struct {
	*cmdcommon.KymaConfig
	outputFormat output.Format
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
		Long:  "Use this command to quickly assess the health, configuration, and potential issues in your cluster for troubleshooting and support purposes",
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(diagnose(&cfg))
		},
	}

	cmd.Flags().VarP(&cfg.outputFormat, "format", "f", "Output format (possible values: json, yaml)")
	cmd.Flags().StringVarP(&cfg.outputPath, "output", "o", "", "Path to the diagnostic output file. If not provided the output will be printed to stdout")
	cmd.Flags().BoolVar(&cfg.verbose, "verbose", false, "Display verbose output including error details during diagnostics collection")

	return cmd
}

func diagnose(cfg *diagnoseConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	fmt.Println("Collecting diagnostics data...")

	diagnosticData := diagnostics.GetData(cfg.Ctx, client, cfg.verbose)
	err := render(&diagnosticData, cfg.outputPath, cfg.outputFormat)

	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to get diagnostic data"))
	}

	fmt.Println("Done.")

	return nil
}

func render(diagData *diagnostics.DiagnosticData, outputFilepath string, outputFormat output.Format) error {
	var outputWriter io.Writer

	if outputFilepath == "" {
		outputWriter = os.Stdout
	} else {
		file, err := os.Create(outputFilepath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		outputWriter = file
	}

	switch outputFormat {
	case output.JSONFormat:
		return renderJSON(outputWriter, diagData)
	case output.YAMLFormat:
		return renderYAML(outputWriter, diagData)
	case output.DefaultFormat:
		return renderYAML(outputWriter, diagData)
	default:
		return fmt.Errorf("unexpected output.Format: %#v", outputFormat)
	}
}

func renderJSON(writer io.Writer, data *diagnostics.DiagnosticData) error {
	obj, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(writer, string(obj))
	return err
}

func renderYAML(writer io.Writer, data *diagnostics.DiagnosticData) error {
	obj, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(writer, string(obj))
	return err
}
