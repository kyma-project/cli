package function

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/flags"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type runtimeConfig struct {
	DepsFilename    string `yaml:"depsFilename"`
	DepsData        string `yaml:"depsData"`
	HandlerFilename string `yaml:"handlerFilename"`
	HandlerData     string `yaml:"handlerData"`
}

type initConfig struct {
	*cmdcommon.KymaConfig

	runtimesConfig map[string]runtimeConfig

	runtime string
	dir     string
}

func NewInitCmd(kymaConfig *cmdcommon.KymaConfig, cmdConfig interface{}) *cobra.Command {
	cfg := &initConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Init source and dependencies files locally",
		Long:  "Use this command to initialise source and dependencies files for a Function.",
		PreRun: func(cmd *cobra.Command, _ []string) {
			clierror.Check(flags.Validate(cmd.Flags(), flags.MarkRequired("runtime")))
			clierror.Check(cfg.complete(cmdConfig))
			clierror.Check(cfg.validate())
		},
		Run: func(cmd *cobra.Command, _ []string) {
			clierror.Check(runInit(cfg, cmd.OutOrStdout()))
		},
	}

	cmd.Flags().StringVar(&cfg.runtime, "runtime", "", "Runtime for which to generate files ")
	cmd.Flags().StringVar(&cfg.dir, "dir", ".", "Path to the directory where files should be created")

	return cmd
}

func (c *initConfig) complete(cmdConfig interface{}) clierror.Error {
	if cmdConfig == nil {
		return clierror.New("unexpected extension error, empty config")
	}

	configBytes, err := yaml.Marshal(cmdConfig)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to marshal config"))
	}

	err = yaml.Unmarshal(configBytes, &c.runtimesConfig)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to unmarshal config"))
	}

	return nil
}

func (c *initConfig) validate() clierror.Error {
	if _, ok := c.runtimesConfig[c.runtime]; !ok {
		return clierror.New(
			fmt.Sprintf("unsupported runtime %s", c.runtime),
			fmt.Sprintf("use on the allowed runtimes on this cluster [ %s ]", strings.Join(mapKeys(c.runtimesConfig), " / ")),
		)
	}

	return nil
}

func mapKeys(m map[string]runtimeConfig) []string {
	keys := []string{}
	for key := range m {
		keys = append(keys, key)
	}

	return keys
}

func runInit(cfg *initConfig, out io.Writer) clierror.Error {
	runtimeCfg := cfg.runtimesConfig[cfg.runtime]

	handlerPath := path.Join(cfg.dir, runtimeCfg.HandlerFilename)
	err := os.WriteFile(handlerPath, []byte(runtimeCfg.HandlerData), os.ModePerm)
	if err != nil {
		return clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to write sources file to %s", handlerPath)))
	}

	depsPath := path.Join(cfg.dir, runtimeCfg.DepsFilename)
	err = os.WriteFile(depsPath, []byte(runtimeCfg.DepsData), os.ModePerm)
	if err != nil {
		return clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to write deps file to %s", depsPath)))
	}

	outDir, err := filepath.Abs(cfg.dir)
	if err != nil {
		// undexpected error, use realtive path
		outDir = cfg.dir
	}

	fmt.Fprintf(out, "Functions files initialised to dir %s\n", outDir)
	fmt.Fprint(out, "\nExample usage:\n")
	fmt.Fprintf(out, "kyma alpha function create %s --runtime %s --source %s --dependencies %s\n", cfg.runtime, cfg.runtime, handlerPath, depsPath)
	return nil
}
