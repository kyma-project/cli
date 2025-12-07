package diagnose

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
	istioformatting "istio.io/istio/istioctl/pkg/util/formatting"
	istioanalysisdiag "istio.io/istio/pkg/config/analysis/diag"
	istioresource "istio.io/istio/pkg/config/resource"
	istiolog "istio.io/istio/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	istioanalysisanalyzers "istio.io/istio/pkg/config/analysis/analyzers"
	istioanalysislocal "istio.io/istio/pkg/config/analysis/local"
	istiokube "istio.io/istio/pkg/kube"
)

type diagnoseIstioConfig struct {
	*cmdcommon.KymaConfig
	namespace     string
	allNamespaces bool
	outputFormat  types.Format
	outputLevel   types.IstioLevel
	outputPath    string
	verbose       bool
	timeout       time.Duration
}

func NewDiagnoseIstioCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := diagnoseIstioConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "istio [flags]",
		Short: "Checks Istio configuration",
		Example: `  # Analyze Istio configuration across all namespaces
  kyma alpha diagnose istio
  # or
  kyma alpha diagnose istio --all-namespaces

  # Analyze Istio configuration in a specific namespace
  kyma alpha diagnose istio --namespace my-namespace

  # Print only warnings and errors
  kyma alpha diagnose istio --level warning

  # Output as JSON to a file
  kyma alpha diagnose istio --format json --output istio-diagnostics.json`,

		Long: "Use this command to quickly assess potential Istio configuration issues in your cluster for troubleshooting and support purposes.",
		PreRun: func(cmd *cobra.Command, args []string) {
			clierror.Check(flags.Validate(cmd.Flags(),
				flags.MarkExclusive("all-namespaces", "namespace"),
			))
		},
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(diagnoseIstio(&cfg))
		},
	}

	cmd.Flags().BoolVarP(&cfg.allNamespaces, "all-namespaces", "A", false, "Analyzes all namespaces")
	cmd.Flags().StringVarP(&cfg.namespace, "namespace", "n", "", "The namespace that the workload instances belongs to")
	cmd.Flags().VarP(&cfg.outputFormat, "format", "f", "Output format (possible values: json, yaml)")
	cfg.outputLevel = "warning"
	cmd.Flags().Var(&cfg.outputLevel, "level", "Output message level (possible values: info, warning, error)")
	cmd.Flags().StringVarP(&cfg.outputPath, "output", "o", "", "Path to the diagnostic output file. If not provided the output is printed to stdout")
	cmd.Flags().BoolVar(&cfg.verbose, "verbose", false, "Displays verbose output, including error details during diagnostics collection")
	cmd.Flags().DurationVar(&cfg.timeout, "timeout", 30*time.Second, "Timeout for diagnosis")

	return cmd
}

func diagnoseIstio(cfg *diagnoseIstioConfig) clierror.Error {
	if cfg.verbose {
		out.EnableVerbose()
	}

	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	muteIstioLogger()

	namespace := calculateNamespace(cfg.allNamespaces, cfg.namespace)

	ctx, cancel := context.WithTimeout(cfg.Ctx, cfg.timeout)
	defer cancel()

	diagnosticData, err := getIstioData(ctx, client, namespace)
	if err != nil {
		return err
	}

	diagnosticData.Messages = filterDataByLevel(diagnosticData.Messages, cfg.outputLevel.ToInternalIstioLevel())

	err = printIstioOutput(diagnosticData, cfg.outputFormat, cfg.outputPath)
	if err != nil {
		return err
	}
	return nil
}

func filterDataByLevel(messages istioanalysisdiag.Messages, minLevel istioanalysisdiag.Level) istioanalysisdiag.Messages {
	var filtered []istioanalysisdiag.Message
	for _, msg := range messages {
		if msg.Type.Level().IsWorseThanOrEqualTo(minLevel) {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

func calculateNamespace(allNamespaces bool, namespace string) string {
	if allNamespaces {
		return metav1.NamespaceAll
	} else if namespace == "" {
		return metav1.NamespaceAll
	}
	return namespace
}

func muteIstioLogger() {
	for _, s := range istiolog.Scopes() {
		s.SetOutputLevel(istiolog.NoneLevel)
		s.SetStackTraceLevel(istiolog.NoneLevel)
		s.SetLogCallers(false)
	}
}

func getIstioData(ctx context.Context, client kube.Client, namespace string) (*istioanalysislocal.AnalysisResult, clierror.Error) {
	combinedAnalyzers := istioanalysisanalyzers.AllCombined()
	sa := istioanalysislocal.NewSourceAnalyzer(
		combinedAnalyzers,
		istioresource.Namespace(namespace),
		"istio-system",
		nil)

	istioClient, err := istiokube.NewCLIClient(
		istiokube.NewClientConfigForRestConfig(client.RestConfig()),
		istiokube.WithRevision(""),
		istiokube.WithCluster(""),
		istiokube.WithTimeout(time.Second*100),
	)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("failed to create Istio kube client"))
	}

	k := istiokube.EnableCrdWatcher(istioClient)
	sa.AddRunningKubeSource(k)

	sa.AddRunningKubeSource(istioClient)

	cancel := make(chan struct{})
	go func() {
		<-ctx.Done()
		close(cancel)
		out.Errfln("Istio analysis cancelled because of timeout")
	}()
	result, err := sa.Analyze(cancel)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("failed to analyze Istio configuration"))
	}
	return &result, nil
}

func printIstioOutput(analysisResult *istioanalysislocal.AnalysisResult, format types.Format, path string) clierror.Error {
	var istioFormat string
	switch format {
	case types.JSONFormat:
		istioFormat = istioformatting.JSONFormat
	case types.YAMLFormat, types.DefaultFormat:
		istioFormat = istioformatting.YAMLFormat
	default:
		return clierror.New(fmt.Sprintf("unexpected output.Format: %v", format))
	}

	printer := out.Default

	if path != "" {
		file, err := os.Create(path)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to create output file: %w"))
		}
		defer file.Close()

		printer = out.NewToWriter(file)
	}

	// Special behavior for no messages
	if len(analysisResult.Messages) == 0 {
		if format == types.JSONFormat {
			printer.Msgfln("{}")
		}
		return nil
	}

	output, err := istioformatting.Print(analysisResult.Messages, istioFormat, false)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to format output"))
	}
	re := regexp.MustCompile(`(?m)^\t+`)
	output = re.ReplaceAllStringFunc(output, func(match string) string {
		return strings.Repeat("  ", len(match))
	})
	printer.Msgfln(output)
	return nil
}
