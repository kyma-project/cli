package diagnose

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
	//istioclient "istio.io/client-go/pkg/clientset/versioned"
	istioanalysisanalyzers "istio.io/istio/pkg/config/analysis/analyzers"
	//istioanalysislocal "istio.io/istio/pkg/config/analysis/local"
	//	istioresource "istio.io/istio/pkg/config/resource"
)

type diagnoseIstioConfig struct {
	*cmdcommon.KymaConfig
	outputFormat types.Format
	outputPath   string
	verbose      bool
}

func NewDiagnoseIstioCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := diagnoseIstioConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "istio [flags]",
		Short: "Diagnose Istio configuration",
		Long:  "Use this command to quickly assess potential Istio configuration issues in your cluster for troubleshooting and support purposes.",
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(diagnoseIstio(&cfg))
		},
	}

	cmd.Flags().VarP(&cfg.outputFormat, "format", "f", "Output format (possible values: json, yaml)")
	cmd.Flags().StringVarP(&cfg.outputPath, "output", "o", "", "Path to the diagnostic output file. If not provided the output is printed to stdout")
	cmd.Flags().BoolVar(&cfg.verbose, "verbose", false, "Display verbose output, including error details during diagnostics collection")

	//TODO:
	// --namespace/-n <ns>: Limit analysis to a namespace (default: all).
	// --all-namespaces/-A: Explicitly analyze all namespaces (mutually exclusive with --namespace).
	// --format: Output format (possible values: json, yaml)
	// --output: Path to the diagnostic output file. If not provided the output is printed to stdout
	// --verbose: Display verbose output, including error details during diagnostics collection
	// --timeout : Timeout for analysis
	return cmd
}

func diagnoseIstio(cfg *diagnoseIstioConfig) clierror.Error {
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

	/*diagnosticData :=*/
	getIstioData(cfg.Ctx, client, cfg)

	//if err != nil {
	//	return clierror.Wrap(err, clierror.New("failed to get diagnostic data"))
	//}

	if cfg.outputPath != "" {
		out.Msgln("Done.")
	}

	return nil
}

func getIstioData(ctx context.Context, client kube.Client, cfg *diagnoseIstioConfig) {
	//cs := istioclient.New(client.RestClient())
	combinedAnalyzers := istioanalysisanalyzers.AllCombined()
	sa := istioanalysislocal.NewSourceAnalyzer(
		combinedAnalyzers,
		"default",
		"kyma-system",
		nil)
	cancel := make(chan struct{})
	sa.AddRunningKubeSource(client)
	result, err := sa.Analyze(cancel)
	if err != nil {
		out.Errf("Istio analysis error: %v", err)
		return
	}
	out.Msgfln("Istio analysis result: %v", result)
}
