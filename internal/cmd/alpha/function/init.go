package function

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
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
			clierror.Check(runInit(cfg, cmd.InOrStdin(), cmd.OutOrStdout()))
		},
	}

	cmd.Flags().StringVar(&cfg.runtime, "runtime", cfg.extensionConfig.DefaultRuntime, fmt.Sprintf("Runtime for which the files are generated [ %s ]", sortedRuntimesString(cfg.extensionConfig.Runtimes)))
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

	for runtimeName, runtimeCfg := range extCfg.Runtimes {
		if !filepath.IsLocal(runtimeCfg.DepsFilename) {
			return nil, errors.New(fmt.Sprintf("deps filename %s for runtime %s is not a single file name", runtimeCfg.DepsFilename, runtimeName))
		}

		if !filepath.IsLocal(runtimeCfg.HandlerFilename) {
			return nil, errors.New(fmt.Sprintf("handler filename %s for runtime %s is not a single file name", runtimeCfg.HandlerFilename, runtimeName))
		}
	}

	return &extCfg, nil
}

func (c *initConfig) validate() clierror.Error {
	if _, ok := c.extensionConfig.Runtimes[c.runtime]; !ok {
		return clierror.New(
			fmt.Sprintf("unsupported runtime %s", c.runtime),
			fmt.Sprintf("use on the allowed runtimes on this cluster [ %s ]", sortedRuntimesString(c.extensionConfig.Runtimes)),
		)
	}

	return nil
}

func runInit(cfg *initConfig, in io.Reader, out io.Writer) clierror.Error {
	runtimeCfg := cfg.extensionConfig.Runtimes[cfg.runtime]

	if !filepath.IsLocal(cfg.dir) {
		// output dir is not a local path, ask user for confirmation
		clierr := getUserAcceptance(in, out, cfg.dir)
		if clierr != nil {
			return clierr
		}
	}

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

func sortedRuntimesString(m map[string]runtimeConfig) string {
	keys := []string{}
	for key := range m {
		keys = append(keys, key)
	}

	sort.Strings(sort.StringSlice(keys))
	return strings.Join(keys, ", ")
}

func getUserAcceptance(in io.Reader, out io.Writer, path string) clierror.Error {
	fmt.Fprintf(out, "The output path ( %s ) seems to be outside of the current working directory.\n", path)
	fmt.Fprint(out, "Do you want to proceed? (y/n): ")

	input, err := bufio.NewReader(in).ReadString('\n') // wait for user to press enter
	fmt.Fprintln(out)

	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to read user input"))
	}

	lowerInput := strings.ToLower(input)
	if lowerInput == "y\n" || lowerInput == "yes\n" {
		// user accepted, continue
		return nil
	}

	return clierror.New(
		"function init aborted",
		"you must provide a local path for the output directory or accept the default one by typing 'y' and pressing enter",
	)
}
