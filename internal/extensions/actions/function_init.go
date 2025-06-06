package actions

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/extensions/actions/common"
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

func (c *functionInitActionConfig) validate() clierror.Error {
	if _, ok := c.Runtimes[c.UseRuntime]; !ok {
		return clierror.New(
			fmt.Sprintf("unsupported runtime %s", c.UseRuntime),
			fmt.Sprintf("use on the allowed runtimes on this cluster [ %s ]", sortedRuntimesString(c.Runtimes)),
		)
	}

	if c.OutputDir == "" {
		return clierror.New("empty output directory path")
	}

	return nil
}

type functionInitAction struct {
	common.TemplateConfigurator[functionInitActionConfig]
}

func NewFunctionInit(*cmdcommon.KymaConfig) types.Action {
	return &functionInitAction{}
}

func (fi *functionInitAction) Run(cmd *cobra.Command, _ []string) clierror.Error {
	clierr := fi.Cfg.validate()
	if clierr != nil {
		return clierr
	}

	runtimeCfg := fi.Cfg.Runtimes[fi.Cfg.UseRuntime]
	handlerPath := path.Join(fi.Cfg.OutputDir, runtimeCfg.HandlerFilename)
	err := os.WriteFile(handlerPath, []byte(runtimeCfg.HandlerData), os.ModePerm)
	if err != nil {
		return clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to write sources file to %s", handlerPath)))
	}

	depsPath := path.Join(fi.Cfg.OutputDir, runtimeCfg.DepsFilename)
	err = os.WriteFile(depsPath, []byte(runtimeCfg.DepsData), os.ModePerm)
	if err != nil {
		return clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to write deps file to %s", depsPath)))
	}

	outDir, err := filepath.Abs(fi.Cfg.OutputDir)
	if err != nil {
		// undexpected error, use realtive path
		outDir = fi.Cfg.OutputDir
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Functions files of runtime %s initialized to dir %s\n", fi.Cfg.UseRuntime, outDir)
	fmt.Fprint(out, "\nNext steps:\n")
	fmt.Fprint(out, "* update output files in your favorite IDE\n")
	fmt.Fprintf(out, "* create Function, for example:\n")
	fmt.Fprintf(out, "  kyma function create %s --runtime %s --source %s --dependencies %s\n", fi.Cfg.UseRuntime, fi.Cfg.UseRuntime, handlerPath, depsPath)
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
