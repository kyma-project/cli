package function

import (
	"fmt"
	"os"
	"path"

	"github.com/kyma-project/cli/cmd/kyma/sync/function"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/vscode"
	"github.com/kyma-project/hydroform/function/pkg/workspace"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	defaultRuntime   = "nodejs18"
	defaultReference = "main"
	defaultBaseDir   = "/"
)

var (
	deprecatedRuntimes = map[string]struct{}{
		"python39": {},
	}
)

type command struct {
	opts *Options
	cli.Command
}

// NewCmd creates a new init command
func NewCmd(o *Options) *cobra.Command {
	c := command{
		opts:    o,
		Command: cli.Command{Options: o.Options},
	}
	cmd := &cobra.Command{
		Use:   "function",
		Short: "Creates local resources for your Function.",
		Long: `Use this command to create the local workspace with the default structure of your Function's code and dependencies. Update this configuration to your references and apply it to a Kyma cluster. 
Use the flags to specify the initial configuration for your Function or to choose the location for your project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Run()
		},
	}

	cmd.Flags().StringVar(&o.Name, "name", "", `Function name.`)
	cmd.Flags().StringVar(&o.Namespace, "namespace", "", `Namespace to which you want to apply your Function.`)
	cmd.Flags().StringVarP(&o.Dir, "dir", "d", "", `Full path to the directory where you want to save the project.`)
	cmd.Flags().StringVar(&o.RuntimeImageOverride, "runtime-image-override", "", `Set custom runtime image base.`)
	cmd.Flags().StringVarP(
		&o.Runtime, "runtime", "r", defaultRuntime,
		`Flag used to define the environment for running your Function. Use one of these options:
	- nodejs18 
	- nodejs20
	- python39 (deprecated)
	- python312`,
	)
	cmd.Flags().StringVar(&o.SchemaVersion, "schema-version", string(workspace.SchemaVersionDefault), `Version of the config API.`)

	// git function options
	cmd.Flags().StringVar(&o.URL, "url", "", `Git repository URL`)
	cmd.Flags().StringVar(&o.RepositoryName, "repository-name", "", `The name of the Git repository to be created`)
	cmd.Flags().StringVar(&o.Reference, "reference", defaultReference, `Commit hash or branch name`)
	cmd.Flags().StringVar(
		&o.BaseDir, "base-dir", defaultBaseDir, `A directory in the repository containing the Function's sources`,
	)
	cmd.Flags().BoolVar(
		&o.VsCode, "vscode", false,
		`Generate VS Code settings containing config.yaml JSON schema for autocompletion (see "kyma get schema -h" for more info)`,
	)

	return cmd
}

func (c *command) Run() error {
	s := c.NewStep("Generating project structure")

	var err error
	if c.K8s, err = kube.NewFromConfig("", c.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	if err := c.opts.setDefaults(c.K8s.DefaultNamespace()); err != nil {
		s.Failure()
		return err
	}

	if _, err := os.Stat(c.opts.Dir); os.IsNotExist(err) {
		err = os.MkdirAll(c.opts.Dir, 0700)
		if err != nil {
			s.Failure()
			return err
		}
	}

	if _, ok := deprecatedRuntimes[c.opts.Runtime]; ok {
		s.LogWarnf(
			"Runtime %s is deprecated and will be removed in the future. We recommend using a supported runtime version",
			c.opts.Runtime,
		)
	}

	schemaVersion, err := function.ParseSchemaVersion(c.opts.SchemaVersion)
	if err != nil {
		s.Failure()
		return err
	}

	configuration := workspace.Cfg{
		Runtime:              c.opts.Runtime,
		RuntimeImageOverride: c.opts.RuntimeImageOverride,
		Name:                 c.opts.Name,
		Namespace:            c.opts.Namespace,
		Source:               c.opts.source(),
		SchemaVersion:        schemaVersion,
	}

	err = workspace.Initialize(configuration, c.opts.Dir)
	if err != nil {
		s.Failure()
		return err
	}

	if !c.opts.VsCode {
		s.Successf("Project generated in %s", c.opts.Dir)
		return nil
	}

	if err := validateVsCodeDir(c.opts.Dir); err != nil {
		s.Failure()
		return err
	}

	vsCodeDirPath := path.Join(c.opts.Dir, ".vscode")
	err = os.MkdirAll(vsCodeDirPath, 0700)
	if err != nil {
		s.Failure()
		return err
	}

	if err := vscode.Workspace.Build(vsCodeDirPath); err != nil {
		s.Failure()
		return err
	}

	s.Successf("Project generated in %s", c.opts.Dir)
	return nil
}

func validateVsCodeDir(vsCodeDir string) error {
	for _, filename := range []string{"settings.json", "schema.json"} {
		_, err := os.Stat(path.Join(vsCodeDir, ".vscode", filename))

		if err == nil {
			return fmt.Errorf("%q already exists", filename)
		}

		if !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}
