package module

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	compdescv2 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/v2"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/pkg/module"
)

const (
	moduleConfigFileArg = "module-config-file"
)

type command struct {
	cli.Command
	opts *Options
}

// NewCmd creates a new Kyma CLI command
func NewCmd(o *Options) *cobra.Command {
	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:   "module [--module-config-file MODULE_CONFIG_FILE | --name MODULE_NAME --version MODULE_VERSION] --registry MODULE_REGISTRY [flags]",
		Short: "Creates a module bundled as an OCI image with the given OCI image name from the contents of the given path",
		Long: `Use this command to create a Kyma module and bundle it as an OCI image.

### Detailed description
[TODO: Update]
Kyma modules are individual components that can be deployed into a Kyma runtime. Modules are built and distributed as OCI container images. 
With this command, you can create such images out of a folder's contents.

This command creates a component descriptor in the descriptor path (./mod as a default) and packages all the contents on the provided path as an OCI image.
Kubebuilder projects are supported. If the path contains a kubebuilder project, it will be built and pre-defined layers will be created based on its known contents.

Alternatively, a custom (non kubebuilder) module can be created by providing a path that does not contain a kubebuilder project. In that case all the contents of the path will be bundled as a single layer.

Optionally, you can manually add additional layers with contents in other paths (see [resource flag](#flags) for more information).

Finally, if you provided a registry to which to push the artifact, the created module is validated and pushed. During the validation the default CR defined in the optional "default.yaml" file is validated against CustomResourceDefinition.
Alternatively, you can trigger an on-demand default CR validation with "--validateCR=true", in case you don't push to the registry.

To push the artifact into some registries, for example, the central docker.io registry, you have to change the OCM Component Name Mapping with the following flag: "--name-mapping=sha256-digest". This is necessary because the registry does not accept artifact URLs with more than two path segments, and such URLs are generated with the default name mapping: "urlPath". In the case of the "sha256-digest" mapping, the artifact URL contains just a sha256 digest of the full Component Name and fits the path length restrictions.

`,

		Example: `Examples:
Build module my-domain/modA in version 1.2.3 and push it to a remote registry
		kyma alpha create module -n my-domain/modA --version 1.2.3 -p /path/to/module --registry https://dockerhub.com
Build module my-domain/modB in version 3.2.1 and push it to a local registry "unsigned" subfolder without tls
		kyma alpha create module -n my-domain/modB --version 3.2.1 -p /path/to/module --registry http://localhost:5001/unsigned --insecure
`,
		RunE:    func(cobraCmd *cobra.Command, args []string) error { return c.Run(cobraCmd.Context()) },
		Aliases: []string{"mod"},
	}

	cmd.Flags().StringVar(
		&o.ModuleConfigFile, moduleConfigFileArg, "",
		"Specifies the module configuration file",
	)

	if o.WithModuleConfigFile() {
		return configureFlagsWithModuleConfigFile(cmd, o)
	}

	return configureLegacyFlags(cmd, o)
}

// configureFlagsWithModuleConfigFile configures the command for creating the module using module config file
func configureFlagsWithModuleConfigFile(cmd *cobra.Command, o *Options) *cobra.Command {

	cmd.Flags().StringVarP(&o.Path, "path", "p", "", "Path to the module's contents. (default current directory)")
	cmd.Flags().StringVar(
		&o.ModuleArchivePath, "module-archive-path", "./mod",
		"Specifies the path where the module artifacts are locally cached to generate the image. If the path already has a module, use the \"--module-archive-version-overwrite\" flag to overwrite it.",
	)
	cmd.Flags().BoolVar(
		&o.PersistentArchive, "module-archive-persistence", false,
		"Uses the host filesystem instead of in-memory archiving to build the module.",
	)
	cmd.Flags().BoolVar(&o.ArchiveVersionOverwrite, "module-archive-version-overwrite", false, "Overwrites existing component's versions of the module. If set to false, the push is a No-Op.")

	cmd.Flags().StringVar(
		&o.RegistryURL, "registry", "",
		"Context URL of the repository. The repository URL will be automatically added to the repository contexts in the module descriptor.",
	)
	cmd.Flags().StringVar(
		&o.NameMappingMode, "name-mapping", "urlPath",
		"Overrides the OCM Component Name Mapping, Use: \"urlPath\" or \"sha256-digest\".",
	)
	cmd.Flags().StringVar(
		&o.RegistryCredSelector, "registry-cred-selector", "",
		"Label selector to identify an externally created Secret of type \"kubernetes.io/dockerconfigjson\". "+
			"It allows the image to be accessed in private image registries. "+
			"It can be used when you push your module to a registry with authenticated access. "+
			"For example, \"label1=value1,label2=value2\".",
	)
	cmd.Flags().StringVarP(
		&o.Credentials, "credentials", "c", "",
		"Basic authentication credentials for the given registry in the user:password format",
	)
	cmd.Flags().StringVar(
		&o.DefaultCRPath, "default-cr", "",
		"File containing the default custom resource of the module. If the module is a kubebuilder project, the default CR is automatically detected.",
	)
	cmd.Flags().StringVarP(
		&o.TemplateOutput, "output", "o", "template.yaml",
		"File to write the module template if the module is uploaded to a registry.",
	)
	cmd.Flags().StringVar(&o.Channel, "channel", "regular", "Channel to use for the module template.")
	cmd.Flags().StringVar(&o.Target, "target", "control-plane", "Target to use when determining where to install the module. Use 'control-plane' or 'remote'.")
	cmd.Flags().StringVar(
		&o.SchemaVersion, "descriptor-version", compdescv2.SchemaVersion, fmt.Sprintf(
			"Schema version to use for the generated OCM descriptor. One of %s",
			strings.Join(compdesc.DefaultSchemes.Names(), ","),
		),
	)
	cmd.Flags().StringVarP(
		&o.Token, "token", "t", "",
		"Authentication token for the given registry (alternative to basic authentication).",
	)
	cmd.Flags().BoolVar(&o.Insecure, "insecure", false, "Uses an insecure connection to access the registry.")
	cmd.Flags().StringVar(
		&o.SecurityScanConfig, "sec-scanners-config", "sec-scanners-config.yaml", "Path to the file holding "+
			"the security scan configuration.",
	)

	cmd.Flags().StringVar(
		&o.PrivateKeyPath, "key", "", "Specifies the path where a private key is used for signing.",
	)

	return cmd
}

// configureLegacyFlags configures the command for the legacy (deprecated) way of creating the module
func configureLegacyFlags(cmd *cobra.Command, o *Options) *cobra.Command {

	cmd.Flags().StringVar(&o.Version, "version", "", "Version of the module. This flag is mandatory.")
	cmd.Flags().StringVarP(
		&o.Name, "name", "n", "",
		"Override the module name of the kubebuilder project. If the module is not a kubebuilder project, this flag is mandatory.",
	)

	cmd.Flags().StringVar(
		&o.ModuleArchivePath, "module-archive-path", "./mod",
		"Specifies the path where the module artifacts are locally cached to generate the image. If the path already has a module, use the \"--module-archive-version-overwrite\" flag to overwrite it.",
	)
	cmd.Flags().BoolVar(
		&o.PersistentArchive, "module-archive-persistence", false,
		"Uses the host filesystem instead of in-memory archiving to build the module.",
	)
	cmd.Flags().BoolVar(&o.ArchiveVersionOverwrite, "module-archive-version-overwrite", false, "Overwrites existing component's versions of the module. If set to false, the push is a No-Op.")

	cmd.Flags().StringVarP(&o.Path, "path", "p", "", "Path to the module's contents. (default current directory)")
	cmd.Flags().StringArrayVarP(
		&o.ResourcePaths, "resource", "r", []string{},
		"Add an extra resource in a new layer in the <NAME:TYPE@PATH> format. If you provide only a path, the name defaults to the last path element, and the type is set to 'helm-chart'.",
	)
	cmd.Flags().StringVar(
		&o.RegistryURL, "registry", "",
		"Context URL of the repository. The repository URL will be automatically added to the repository contexts in the module descriptor.",
	)
	cmd.Flags().StringVar(
		&o.NameMappingMode, "name-mapping", "urlPath",
		"Overrides the OCM Component Name Mapping, Use: \"urlPath\" or \"sha256-digest\".",
	)
	cmd.Flags().StringVar(
		&o.RegistryCredSelector, "registry-cred-selector", "",
		"Label selector to identify an externally created Secret of type \"kubernetes.io/dockerconfigjson\". "+
			"It allows the image to be accessed in private image registries. "+
			"It can be used when you push your module to a registry with authenticated access. "+
			"For example, \"label1=value1,label2=value2\".",
	)
	cmd.Flags().StringVarP(
		&o.Credentials, "credentials", "c", "",
		"Basic authentication credentials for the given registry in the user:password format",
	)
	cmd.Flags().StringVar(
		&o.DefaultCRPath, "default-cr", "",
		"File containing the default custom resource of the module. If the module is a kubebuilder project, the default CR is automatically detected.",
	)
	cmd.Flags().StringVarP(
		&o.TemplateOutput, "output", "o", "template.yaml",
		"File to write the module template if the module is uploaded to a registry.",
	)
	cmd.Flags().StringVar(&o.Channel, "channel", "regular", "Channel to use for the module template.")
	cmd.Flags().StringVar(&o.Target, "target", "control-plane", "Target to use when determining where to install the module. Use 'control-plane' or 'remote'.")
	cmd.Flags().StringVar(
		&o.SchemaVersion, "descriptor-version", compdescv2.SchemaVersion, fmt.Sprintf(
			"Schema version to use for the generated OCM descriptor. One of %s",
			strings.Join(compdesc.DefaultSchemes.Names(), ","),
		),
	)
	cmd.Flags().StringVarP(
		&o.Token, "token", "t", "",
		"Authentication token for the given registry (alternative to basic authentication).",
	)
	cmd.Flags().BoolVar(&o.Insecure, "insecure", false, "Uses an insecure connection to access the registry.")
	cmd.Flags().StringVar(
		&o.SecurityScanConfig, "sec-scanners-config", "sec-scanners-config.yaml", "Path to the file holding "+
			"the security scan configuration.",
	)

	cmd.Flags().StringVar(
		&o.PrivateKeyPath, "key", "", "Specifies the path where a private key is used for signing.",
	)

	return cmd
}

func (cmd *command) Run(ctx context.Context) error {
	osFS := osfs.New()

	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}
	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	l := cli.NewLogger(cmd.opts.Verbose).Sugar()
	undo := zap.RedirectStdLog(l.Desugar())
	defer undo()

	if !cmd.opts.NonInteractive {
		cli.AlphaWarn()
	}

	if err := cmd.opts.Validate(); err != nil {
		return err
	}

	modDef, err := cmd.moduleDefinitionFromOptions()
	if err != nil {
		return err
	}

	cmd.NewStep("Parse and build module...")

	// Create base resource defs with module root and its sub-layers
	if cmd.opts.WithModuleConfigFile() {
		if err := module.Inspect(modDef, l); err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
	} else {
		if err := module.InspectLegacy(modDef, cmd.opts.ResourcePaths, cmd.CurrentStep, l); err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
	}
	cmd.CurrentStep.Successf("Module built")

	if err := cmd.validateDefaultCR(ctx, modDef, l); err != nil {
		return err
	}

	cmd.NewStep("Creating module archive")
	var archiveFS vfs.FileSystem
	if cmd.opts.PersistentArchive {
		archiveFS = osFS
		l.Info("using host filesystem for archive")
	} else {
		archiveFS = memoryfs.New()
		l.Info("using in-memory archive")
	}
	// this builds the archive in memory, Alternatively one can store it on disk or in temp folder
	archive, err := module.CreateArchive(archiveFS, cmd.opts.ModuleArchivePath, modDef)
	if err != nil {
		cmd.CurrentStep.Failure()
		return err
	}
	cmd.CurrentStep.Successf("Module archive created")

	//DEBUG
	dump, _ := module.DumpDescriptor(archive.GetDescriptor())
	fmt.Println("// 1 ///////////////////////////////////")
	fmt.Println(dump)
	fmt.Println("== 1 ===================================")

	cmd.NewStep("Adding layers to archive...")

	if err := module.AddResources(archive, modDef, l, osFS, cmd.opts.RegistryCredSelector); err != nil {
		cmd.CurrentStep.Failure()
		return err
	}

	cmd.CurrentStep.Success()

	if cmd.opts.SecurityScanConfig != "" {
		if _, err := osFS.Stat(cmd.opts.SecurityScanConfig); err == nil {
			cmd.NewStep("Configuring security scanning...")
			err = module.AddSecurityScanningMetadata(archive.GetDescriptor(), cmd.opts.SecurityScanConfig)
			if err != nil {
				cmd.CurrentStep.Failure()
				return err
			}
			if err := archive.Update(); err != nil {
				return fmt.Errorf("could not write security scanning configuration into archive: %w", err)
			}
			cmd.CurrentStep.Successf("Security scanning configured")
		} else {
			l.Warnf("Security scanning configuration was skipped: %s", err.Error())
		}
	}

	/* -- PUSH & TEMPLATE -- */

	if cmd.opts.RegistryURL != "" {

		cmd.NewStep(fmt.Sprintf("Pushing image to %q", cmd.opts.RegistryURL))
		remote, err := cmd.getRemote(modDef.NameMappingMode)
		if err != nil {
			cmd.CurrentStep.Failure()
			return err
		}

		componentVersionAccess, err := remote.Push(archive, cmd.opts.ArchiveVersionOverwrite)

		if err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
		cmd.CurrentStep.Successf("Module successfully pushed to %q", cmd.opts.RegistryURL)

		if cmd.opts.PrivateKeyPath != "" {
			signCfg := &module.ComponentSignConfig{
				Name:    modDef.Name,
				Version: modDef.Version,
				KeyPath: cmd.opts.PrivateKeyPath,
			}

			cmd.NewStep("Fetching and signing component descriptor...")
			if err = module.Sign(signCfg, remote); err != nil {
				cmd.CurrentStep.Failure()
				return err
			}
			cmd.CurrentStep.Success()
		}

		cmd.NewStep("Generating module template")
		t, err := module.Template(componentVersionAccess, cmd.opts.Channel, cmd.opts.Target, modDef.DefaultCR)
		if err != nil {
			cmd.CurrentStep.Failure()
			return err
		}

		if err := vfs.WriteFile(osFS, cmd.opts.TemplateOutput, t, os.ModePerm); err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
		cmd.CurrentStep.Successf("Template successfully generated at %s", cmd.opts.TemplateOutput)
	}

	return nil
}

func (cmd *command) validateDefaultCR(ctx context.Context, modDef *module.Definition, l *zap.SugaredLogger) error {
	cmd.NewStep("Validating Default CR")
	var crValidator *module.DefaultCRValidator
	var err error

	if cmd.opts.WithModuleConfigFile() {
		//TODO: Implement
		cmd.CurrentStep.Successf("Default CR validation skipped (not implemented yet)")
		return nil
	} else {
		crValidator, err = module.NewDefaultCRValidator(modDef.DefaultCR, modDef.Source)
	}
	if err != nil {
		cmd.CurrentStep.Failure()
		return err
	}

	if err := crValidator.Run(ctx, l); err != nil {
		if errors.Is(err, module.ErrEmptyCR) {
			cmd.CurrentStep.Successf("Default CR validation skipped - no default CR")
			return nil
		}
		return err
	}
	cmd.CurrentStep.Successf("Default CR validation succeeded")
	return nil
}

func (cmd *command) getRemote(nameMapping module.NameMapping) (*module.Remote, error) {

	res := &module.Remote{
		Registry:    cmd.opts.RegistryURL,
		NameMapping: nameMapping,
		Credentials: cmd.opts.Credentials,
		Token:       cmd.opts.Token,
		Insecure:    cmd.opts.Insecure,
	}

	if strings.HasPrefix(strings.ToLower(cmd.opts.RegistryURL), "https:") {
		res.Insecure = false
		return res, nil
	}

	if strings.HasPrefix(strings.ToLower(cmd.opts.RegistryURL), "http:") {
		res.Insecure = true

		if !cmd.opts.Insecure {
			cmd.CurrentStep.LogWarn("CAUTION: Pushing the module artifact to the insecure registry")
			if !cmd.opts.NonInteractive && !cmd.CurrentStep.PromptYesNo("Do you really want to proceed? ") {
				return nil, errors.New("command stopped by user")
			}
		}
	}

	return res, nil
}

func (cmd *command) moduleDefinitionFromOptions() (*module.Definition, error) {

	var res *module.Definition
	nameMappingMode, err := module.ParseNameMapping(cmd.opts.NameMappingMode)
	if err != nil {
		return nil, err
	}

	if cmd.opts.WithModuleConfigFile() {
		//new approach, config-file  based

		moduleConfig, err := ParseConfig(cmd.opts.ModuleConfigFile)
		if err != nil {
			return nil, err
		}

		err = moduleConfig.Validate()
		if err != nil {
			return nil, err
		}

		defaultCRPath, err := resolveFilePath(moduleConfig.DefaultCRPath, cmd.opts.Path)
		if err != nil {
			return nil, fmt.Errorf("%w,  %w", ErrDefaultCRPathValidation, err)
		}

		moduleManifestPath, err := resolveFilePath(moduleConfig.ManifestPath, cmd.opts.Path)
		if err != nil {
			return nil, fmt.Errorf("%w,  %w", ErrManifestPathValidation, err)
		}

		res = &module.Definition{
			Name:               moduleConfig.Name,
			Version:            moduleConfig.Version,
			Source:             cmd.opts.Path,
			RegistryURL:        cmd.opts.RegistryURL,
			NameMappingMode:    nameMappingMode,
			DefaultCRPath:      defaultCRPath,
			SingleManifestPath: moduleManifestPath,
			SchemaVersion:      cmd.opts.SchemaVersion,
		}
	} else {
		//legacy approach, flag-based
		res = &module.Definition{
			Name:            cmd.opts.Name,
			Version:         cmd.opts.Version,
			Source:          cmd.opts.Path,
			RegistryURL:     cmd.opts.RegistryURL,
			NameMappingMode: nameMappingMode,
			DefaultCRPath:   cmd.opts.DefaultCRPath,
			SchemaVersion:   cmd.opts.SchemaVersion,
		}
	}

	return res, nil
}

// resolvePath resolves given path if it's absolute or uses the provided prefix to make it absolute.
// Returns an error if the path does not exist or is a directory.
func resolveFilePath(given, absolutePrefix string) (string, error) {

	res := given

	if !filepath.IsAbs(res) {
		res = filepath.Join(absolutePrefix, given)
	}

	fi, err := os.Stat(res)
	if err != nil {
		return "", err
	}
	if fi.IsDir() {
		return "", fmt.Errorf("%q is directory", res)
	}

	return res, nil
}
