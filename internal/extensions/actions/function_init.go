package actions

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/spf13/cobra"
)

type functionInitActionConfig struct {
	UseRuntime string                   `yaml:"useRuntime"`
	OutputDir  string                   `yaml:"outputDir"`
	Runtimes   map[string]runtimeConfig `yaml:"runtimes"`
}

type runtimeConfig struct {
	DepsFilename    string `yaml:"depsFilename"`
	DepsData        string `yaml:"depsData"`
	HandlerFilename string `yaml:"handlerFilename"`
	HandlerData     string `yaml:"handlerData"`
}

func NewFunctionInit(_ *cmdcommon.KymaConfig, actionConfig types.ActionConfig) types.CmdRun {
	return func(cmd *cobra.Command, _ []string) {
		cfg := functionInitActionConfig{}
		clierror.Check(parseActionConfig(actionConfig, &cfg))
		clierror.Check(validate(cfg.Runtimes, cfg.UseRuntime))
		clierror.Check(runInit(&cfg, cmd.OutOrStdout()))
	}
}

func validate(runtimes map[string]runtimeConfig, runtime string) clierror.Error {
	if _, ok := runtimes[runtime]; !ok {
		return clierror.New(
			fmt.Sprintf("unsupported runtime %s", runtime),
			fmt.Sprintf("use on the allowed runtimes on this cluster [ %s ]", sortedRuntimesString(runtimes)),
		)
	}

	return nil
}

func runInit(cfg *functionInitActionConfig, out io.Writer) clierror.Error {
	runtimeCfg := cfg.Runtimes[cfg.UseRuntime]

	handlerPath := path.Join(cfg.OutputDir, runtimeCfg.HandlerFilename)
	err := os.WriteFile(handlerPath, []byte(runtimeCfg.HandlerData), os.ModePerm)
	if err != nil {
		return clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to write sources file to %s", handlerPath)))
	}

	depsPath := path.Join(cfg.OutputDir, runtimeCfg.DepsFilename)
	err = os.WriteFile(depsPath, []byte(runtimeCfg.DepsData), os.ModePerm)
	if err != nil {
		return clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to write deps file to %s", depsPath)))
	}

	outDir, err := filepath.Abs(cfg.OutputDir)
	if err != nil {
		// undexpected error, use realtive path
		outDir = cfg.OutputDir
	}

	fmt.Fprintf(out, "Functions files of runtime %s initialized to dir %s\n", cfg.UseRuntime, outDir)
	fmt.Fprint(out, "\nNext steps:\n")
	fmt.Fprint(out, "* update output files in your favorite IDE\n")
	fmt.Fprintf(out, "* create Function, for example:\n")
	fmt.Fprintf(out, "  kyma alpha function create %s --runtime %s --source %s --dependencies %s\n", cfg.UseRuntime, cfg.UseRuntime, handlerPath, depsPath)
	return nil
}

func sortedRuntimesString(m map[string]runtimeConfig) string {
	keys := []string{}
	for key := range m {
		keys = append(keys, key)
	}

	sort.Strings(sort.StringSlice(keys))
	return strings.Join(keys, ", ")
}
