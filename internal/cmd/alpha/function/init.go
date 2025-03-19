package function

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type runtimesConfig struct {
	Python runtimeConfig `yaml:"python"`
	Nodejs runtimeConfig `yaml:"nodejs"`
}

type runtimeConfig struct {
	LatestVersion   string `yaml:"latestVersion"`
	DepsFilename    string `yaml:"depsFilename"`
	DepsData        string `yaml:"depsData"`
	HandlerFilename string `yaml:"handlerFilename"`
	HandlerData     string `yaml:"handlerData"`
}

var runtimes = []string{"python", "nodejs"}

type initConfig struct {
	*cmdcommon.KymaConfig

	runtimesConfig runtimesConfig

	runtime string
	dir     string
}

func NewInitCmd(kymaConfig *cmdcommon.KymaConfig, cmdConfig interface{}) *cobra.Command {
	cfg := &initConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "init source and dependencies files locally",
		Long:  "Use this command to initialise source and dependencies files for a Function.",
		PreRun: func(_ *cobra.Command, _ []string) {
			clierror.Check(cfg.complete(cmdConfig))
			clierror.Check(cfg.validate())
		},
		Run: func(cmd *cobra.Command, _ []string) {
			clierror.Check(runInit(cfg, cmd.OutOrStdout()))
		},
	}

	cmd.Flags().StringVar(&cfg.runtime, "runtime", "nodejs", fmt.Sprintf("Runtime for which to generate files [ %s ]", strings.Join(runtimes, " / ")))
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
	if !slices.Contains(runtimes, c.runtime) {
		return clierror.New(
			fmt.Sprintf("unsupported runtime %s", c.runtime),
			fmt.Sprintf("use on the allowed runtimes [ %s ]", strings.Join(runtimes, " / ")),
		)
	}

	return nil
}

func runInit(cfg *initConfig, out io.Writer) clierror.Error {
	if cfg.runtime == "python" {
		return initWorkspace(out, cfg.dir, &cfg.runtimesConfig.Python)
	}

	return initWorkspace(out, cfg.dir, &cfg.runtimesConfig.Nodejs)
}

func initWorkspace(out io.Writer, dir string, cfg *runtimeConfig) clierror.Error {
	handlerPath := path.Join(dir, cfg.HandlerFilename)
	err := os.WriteFile(handlerPath, []byte(cfg.HandlerData), os.ModePerm)
	if err != nil {
		return clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to write sources file to %s", handlerPath)))
	}

	depsPath := path.Join(dir, cfg.DepsFilename)
	err = os.WriteFile(depsPath, []byte(cfg.DepsData), os.ModePerm)
	if err != nil {
		return clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to write deps file to %s", depsPath)))
	}

	outDir, err := filepath.Abs(dir)
	if err != nil {
		// undexpected error, use realtive path
		outDir = dir
	}

	fmt.Fprintf(out, "Functions files initialised to dir %s\n", outDir)
	fmt.Fprint(out, "\nExample usage:\n")
	fmt.Fprintf(out, "kyma alpha function create %s --runtime %s --source %s --dependencies %s\n", cfg.LatestVersion, cfg.LatestVersion, handlerPath, depsPath)
	return nil
}
