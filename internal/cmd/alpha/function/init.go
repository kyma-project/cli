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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type extensionConfig struct {
	DefaultRuntime string                   `yaml:"defaultRuntime"`
	Runtimes       map[string]runtimeConfig `yaml:"runtimes"`
}

type runtimeConfig struct {
	DepsFilename    string `yaml:"depsFilename"`
	DepsData        string `yaml:"depsData"`
	HandlerFilename string `yaml:"handlerFilename"`
	HandlerData     string `yaml:"handlerData"`
}

type initConfig struct {
	*cmdcommon.KymaConfig

	extensionConfig *extensionConfig

	runtime string
	dir     string
}

func NewInitCmd(kymaConfig *cmdcommon.KymaConfig, cmdConfig interface{}) (*cobra.Command, error) {
	extensionConfig, err := parseExtensionConfig(cmdConfig)
	if err != nil {
		return nil, err
	}

	cfg := &initConfig{
		KymaConfig:      kymaConfig,
		extensionConfig: extensionConfig,
	}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Init source and dependencies files locally",
		Long:  "Use this command to initialize source and dependencies files for a Function.",
		PreRun: func(cmd *cobra.Command, _ []string) {
			clierror.Check(cfg.validate())
		},
		Run: func(cmd *cobra.Command, _ []string) {
			clierror.Check(runInit(cfg, cmd.OutOrStdout()))
		},
	}

	cmd.Flags().StringVar(&cfg.runtime, "runtime", cfg.extensionConfig.DefaultRuntime, fmt.Sprintf("Runtime for which the files are generated [ %s ]", strings.Join(mapKeys(cfg.extensionConfig.Runtimes), ", ")))
	cmd.Flags().StringVar(&cfg.dir, "dir", ".", "Path to the directory where files must be created")

	return cmd, nil
}

func parseExtensionConfig(cmdConfig interface{}) (*extensionConfig, error) {
	if cmdConfig == nil {
		return nil, errors.New("unexpected extension error, empty config object")
	}

	configBytes, err := yaml.Marshal(cmdConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal config")
	}

	extCfg := extensionConfig{}
	err = yaml.Unmarshal(configBytes, &extCfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config")
	}

	if extCfg.DefaultRuntime == "" || len(extCfg.Runtimes) == 0 {
		// simple validation
		return nil, errors.New("unexpected extension error, empty config data")
	}

	return &extCfg, nil
}

func (c *initConfig) validate() clierror.Error {
	if _, ok := c.extensionConfig.Runtimes[c.runtime]; !ok {
		return clierror.New(
			fmt.Sprintf("unsupported runtime %s", c.runtime),
			fmt.Sprintf("use on the allowed runtimes on this cluster [ %s ]", strings.Join(mapKeys(c.extensionConfig.Runtimes), ", ")),
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
	runtimeCfg := cfg.extensionConfig.Runtimes[cfg.runtime]

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

	fmt.Fprintf(out, "Functions files of runtime %s initialized to dir %s\n", cfg.runtime, outDir)
	fmt.Fprint(out, "\nNext steps:\n")
	fmt.Fprint(out, "* update output files in your favorite IDE\n")
	fmt.Fprintf(out, "* create Function, for example:\n")
	fmt.Fprintf(out, "  kyma alpha function create %s --runtime %s --source %s --dependencies %s\n", cfg.runtime, cfg.runtime, handlerPath, depsPath)
	return nil
}
