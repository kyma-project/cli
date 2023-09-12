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
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/internal/nice"
	"github.com/kyma-project/cli/pkg/module"
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
		Use:   "module [--module-config-file MODULE_CONFIG_FILE | --name MODULE_NAME --version MODULE_VERSION] [--path MODULE_DIRECTORY] [--registry MODULE_REGISTRY] [flags]",
		Short: "Creates a module bundled as an OCI artifact",
		Long: `Use this command to create a Kyma module, bundle it as an OCI artifact and optionally push it to the OCI registry.

### Detailed description

This command allows you to create a Kyma module as an OCI artifact and optionally push it to the OCI registry of your choice.
For more information about a Kyma module see the [documentation](https://github.com/kyma-project/lifecycle-manager).

This command creates a module from an existing directory containing the module's source files.
The directory must be a valid git project that is publicly available.
The command supports two types of directory layouts for the module:
- Simple: Just a directory with a valid git configuration. All the module's sources are defined in this directory.
- Kubebuilder (DEPRECATED): A directory with a valid Kubebuilder project. Module operator(s) are created using the Kubebuilder toolset.
Both simple and Kubebuilder projects require providing an explicit path to the module's project directory using the "--path" flag or invoking the command from within that directory.

### Simple mode configuration

To configure the simple mode, provide the "--module-config-file" flag with a config file path.
The module config file is a YAML file used to configure the following attributes for the module:

- name:         a string, required, the name of the module
- version:      a string, required, the version of the module
- channel:      a string, required, channel that should be used in the ModuleTemplate CR
- manifest:     a string, required, reference to the manifest, must be a relative file name
- defaultCR:    a string, optional, reference to a YAML file containing the default CR for the module, must be a relative file name
- resourceName: a string, optional, default={NAME}-{CHANNEL}, the name for the ModuleTemplate CR that will be created
- security:     a string, optional, name of the security scanners config file
- internal:     a boolean, optional, default=false, determines whether the ModuleTemplate CR should have the internal flag or not
- beta:         a boolean, optional, default=false, determines whether the ModuleTemplate CR should have the beta flag or not
- labels:       a map with string keys and values, optional, additional labels for the generated ModuleTemplate CR
- annotations:  a map with string keys and values, optional, additional annotations for the generated ModuleTemplate CR

The **manifest** and **defaultCR** paths are resolved against the module's directory, as configured with the "--path" flag.
The **manifest** file contains all the module's resources in a single, multi-document YAML file. These resources will be created in the Kyma cluster when the module is activated.
The **defaultCR** file contains a default custom resource for the module that will be installed along with the module.
The Default CR is additionally schema-validated against the Custom Resource Definition. The CRD used for the validation must exist in the set of the module's resources.

### Kubebuilder mode configuration
The Kubebuilder mode is DEPRECATED.
The Kubebuilder mode is configured automatically if the "--module-config-file" flag is not provided.

In this mode, you have to explicitly provide the module name and version using the "--name" and "--version" flags, respectively.
Some defaults, like the module manifest file location and the default CR file location, are then resolved automatically, but you can override these with the available flags.

### Modules as OCI artifacts
Modules are built and distributed as OCI artifacts. 
This command creates a component descriptor in the configured descriptor path (./mod as a default) and packages all the contents on the provided path as an OCI artifact.
The internal structure of the artifact conforms to the [Open Component Model](https://ocm.software/) scheme version 3.

If you configured the "--registry" flag, the created module is validated and pushed to the configured registry.
During the validation the **defaultCR** resource, if defined, is validated against a corresponding CustomResourceDefinition.
You can also trigger an on-demand **defaultCR** validation with "--validateCR=true", in case you don't push the module to the registry.

#### Name Mapping
To push the artifact into some registries, for example, the central docker.io registry, you have to change the OCM Component Name Mapping with the following flag: "--name-mapping=sha256-digest". This is necessary because the registry does not accept artifact URLs with more than two path segments, and such URLs are generated with the default name mapping: **urlPath**. In the case of the "sha256-digest" mapping, the artifact URL contains just a sha256 digest of the full Component Name and fits the path length restrictions. The downside of the "sha256-mapping" is that the module name is no longer visible in the artifact URL, as it contains the sha256 digest of the defined name.

`,

		Example: `Examples:
Build a simple module and push it to a remote registry
		kyma alpha create module --module-config-file=/path/to/module-config-file -path /path/to/module --registry http://localhost:5001/unsigned --insecure
Build a Kubebuilder module my-domain/modB in version 1.2.3 and push it to a remote registry
		kyma alpha create module --name my-domain/modB --version 1.2.3 --path /path/to/module --registry https://dockerhub.com
Build a Kubebuilder module my-domain/modC in version 3.2.1 and push it to a local registry "unsigned" subfolder without tls
		kyma alpha create module --name my-domain/modC --version 3.2.1 --path /path/to/module --registry http://localhost:5001/unsigned --insecure

`,
		RunE:    func(cobraCmd *cobra.Command, args []string) error { return c.Run(cobraCmd.Context()) },
		Aliases: []string{"mod"},
	}

	cmd.Flags().StringVar(
		&o.ModuleConfigFile, "module-config-file", "",
		"Specifies the module configuration file",
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

	cmd.Flags().StringVar(
		&o.GitRemote, "git-remote", "origin",
		"Specifies the remote name of the wanted GitHub repository. For Example \"origin\" or \"upstream\"",
	)

	cmd.Flags().StringVarP(&o.Path, "path", "p", "", "Path to the module's contents. (default current directory)")

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

	cmd.Flags().StringVarP(
		&o.TemplateOutput, "output", "o", "template.yaml",
		"File to write the module template if the module is uploaded to a registry.",
	)

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

	cmd.Flags().BoolVar(&o.KubebuilderProject, "kubebuilder-project", false, "Specifies provided module is a Kubebuilder Project.")

	configureLegacyFlags(cmd, o)

	return cmd
}

// configureLegacyFlags configures the command for the legacy (deprecated) way of creating the module
func configureLegacyFlags(cmd *cobra.Command, o *Options) *cobra.Command {

	cmd.Flags().StringVar(&o.Version, "version", "", "Version of the module. This flag is mandatory.")

	cmd.Flags().StringVarP(
		&o.Name, "name", "n", "",
		"Override the module name of the kubebuilder project. If the module is not a kubebuilder project, this flag is mandatory.",
	)

	cmd.Flags().StringArrayVarP(
		&o.ResourcePaths, "resource", "r", []string{},
		"Add an extra resource in a new layer in the <NAME:TYPE@PATH> format. If you provide only a path, the name defaults to the last path element, and the type is set to 'helm-chart'.",
	)

	cmd.Flags().StringVar(
		&o.DefaultCRPath, "default-cr", "",
		"File containing the default custom resource of the module. If the module is a kubebuilder project, the default CR is automatically detected.",
	)

	cmd.Flags().StringVar(&o.Channel, "channel", "regular", "Channel to use for the module template.")

	cmd.Flags().StringVar(&o.Namespace, "namespace", kcpSystemNamespace, "Specifies the namespace where the ModuleTemplate is deployed.")

	return cmd
}

type validator interface {
	GetCrd() []byte
	Run(ctx context.Context, log *zap.SugaredLogger) error
}

const kcpSystemNamespace = "kcp-system"

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

	modDef, modCnf, err := cmd.moduleDefinitionFromOptions()

	if err != nil {
		return err
	}

	cmd.NewStep("Parse and build module...")

	// Create base resource defs with module root and its sub-layers
	if cmd.opts.KubebuilderProject {
		if err := module.InspectLegacy(modDef, cmd.opts.ResourcePaths, cmd.CurrentStep, l); err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
	} else {
		if err := module.Inspect(modDef, l); err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
	}
	cmd.CurrentStep.Successf("Module built")

	var crValidator validator
	if crValidator, err = cmd.validateDefaultCR(ctx, modDef, l); err != nil {
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

	var archive *comparch.ComponentArchive
	gitPath, err := files.SearchForTargetDirByName(modDef.Source, ".git")
	if gitPath == "" || err != nil {
		l.Warnf("could not find git repository root, using %s directory", modDef.Source)
		n := nice.NewNice(cmd.opts.NonInteractive)
		n.PrintImportant("\n! CAUTION: The target folder is not a git repository. The sources will be not added to the layer")
		if files.IsFileExists(cmd.opts.SecurityScanConfig) {
			n.PrintImportant("  The security scan configuration file has been provided, but it will be skipped due to the absence of repository information.")
		}
		if !cmd.avoidUserInteraction() {
			if !cmd.CurrentStep.PromptYesNo("Do you want to continue? ") {
				cmd.CurrentStep.Failure()
				return errors.New("command stopped by user")
			}
		}

		archive, err = module.CreateArchive(archiveFS, cmd.opts.ModuleArchivePath, cmd.opts.GitRemote, modDef, false)
		if err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
	} else {
		l.Infof("found git repository root at %s", gitPath)
		l.Infof("adding sources to the layer")
		modDef.Source = gitPath // set the source to the git root
		archive, err = module.CreateArchive(archiveFS, cmd.opts.ModuleArchivePath, cmd.opts.GitRemote, modDef, true)
		if err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
	}

	cmd.CurrentStep.Successf("Module archive created")

	cmd.NewStep("Adding layers to archive...")

	if err := module.AddResources(archive, modDef, l, osFS, cmd.opts.RegistryCredSelector); err != nil {
		cmd.CurrentStep.Failure()
		return err
	}

	cmd.CurrentStep.Success()

	// Security Scan
	if cmd.opts.SecurityScanConfig != "" && gitPath != "" { // security scan is only supported for target git repositories
		cmd.NewStep("Configuring security scanning...")
		if files.IsFileExists(cmd.opts.SecurityScanConfig) {
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
			l.Warnf("Security scanning configuration was skipped")
			cmd.CurrentStep.Failure()
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

		labels := map[string]string{}
		annotations := map[string]string{}

		var resourceName = ""

		if modCnf != nil {
			resourceName = modCnf.ResourceName

			maps.Copy(labels, modCnf.Labels)
			maps.Copy(annotations, modCnf.Annotations)

			if modCnf.Beta {
				labels["operator.kyma-project.io/beta"] = "true"
			}
			if modCnf.Internal {
				labels["operator.kyma-project.io/internal"] = "true"
			}
		}
		isClusterScoped := isCrdClusterScoped(crValidator.GetCrd())
		if isClusterScoped {
			annotations["operator.kyma-project.io/is-cluster-scoped"] = "true"
		} else {
			annotations["operator.kyma-project.io/is-cluster-scoped"] = "false"
		}

		var channel = cmd.opts.Channel
		if modCnf != nil {
			channel = modCnf.Channel
		}

		var namespace = cmd.opts.Namespace
		if modCnf != nil && modCnf.Namespace != "" {
			namespace = modCnf.Namespace
		}

		t, err := module.Template(componentVersionAccess, resourceName, namespace,
			channel, modDef.DefaultCR, labels, annotations, modDef.CustomStateChecks)

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

func (cmd *command) validateDefaultCR(ctx context.Context, modDef *module.Definition, l *zap.SugaredLogger) (validator, error) {
	cmd.NewStep("Validating Default CR")

	var crValidator validator
	if cmd.opts.KubebuilderProject {
		crValidator = module.NewDefaultCRValidator(modDef.DefaultCR, modDef.Source)
	} else {
		crValidator = module.NewSingleManifestFileCRValidator(modDef.DefaultCR, modDef.SingleManifestPath)
	}

	if err := crValidator.Run(ctx, l); err != nil {
		if errors.Is(err, module.ErrEmptyCR) {
			cmd.CurrentStep.Successf("Default CR validation skipped - no default CR")
			return crValidator, nil
		}
		return crValidator, err
	}
	cmd.CurrentStep.Successf("Default CR validation succeeded")
	return crValidator, nil
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

func (cmd *command) moduleDefinitionFromOptions() (*module.Definition, *Config, error) {

	var def *module.Definition
	var cnf *Config

	nameMappingMode, err := module.ParseNameMapping(cmd.opts.NameMappingMode)
	if err != nil {
		return nil, nil, err
	}

	if cmd.opts.KubebuilderProject {
		np := nice.Nice{}
		np.PrintImportant("WARNING: The Kubebuilder support is DEPRECATED. Use the simple mode by providing the \"--module-config-file\" flag instead.")

		//legacy approach, flag-based
		def = &module.Definition{
			Name:              cmd.opts.Name,
			Version:           cmd.opts.Version,
			Source:            cmd.opts.Path,
			RegistryURL:       cmd.opts.RegistryURL,
			NameMappingMode:   nameMappingMode,
			DefaultCRPath:     cmd.opts.DefaultCRPath,
			SchemaVersion:     cmd.opts.SchemaVersion,
			CustomStateChecks: nil,
		}
		return def, cnf, nil
	}

	//new approach, config-file  based
	moduleConfig, err := ParseConfig(cmd.opts.ModuleConfigFile)
	if err != nil {
		return nil, nil, err
	}

	err = moduleConfig.Validate()
	if err != nil {
		return nil, nil, err
	}

	var defaultCRPath string
	if moduleConfig.DefaultCRPath != "" {
		defaultCRPath, err = resolveFilePath(moduleConfig.DefaultCRPath, cmd.opts.Path)
		if err != nil {
			return nil, nil, fmt.Errorf("%w,  %w", ErrDefaultCRPathValidation, err)
		}
	}

	moduleManifestPath, err := resolveFilePath(moduleConfig.ManifestPath, cmd.opts.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("%w,  %w", ErrManifestPathValidation, err)
	}

	def = &module.Definition{
		Name:               moduleConfig.Name,
		Version:            moduleConfig.Version,
		Source:             cmd.opts.Path,
		RegistryURL:        cmd.opts.RegistryURL,
		NameMappingMode:    nameMappingMode,
		DefaultCRPath:      defaultCRPath,
		SingleManifestPath: moduleManifestPath,
		SchemaVersion:      cmd.opts.SchemaVersion,
		CustomStateChecks:  moduleConfig.CustomStateChecks,
	}
	cnf = moduleConfig

	return def, cnf, nil
}

// avoidUserInteraction returns true if user won't provide input
func (cmd *command) avoidUserInteraction() bool {
	return cmd.NonInteractive || cmd.CI
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

func isCrdClusterScoped(crdBytes []byte) bool {
	if crdBytes == nil {
		return false
	}

	crd := &apiextensions.CustomResourceDefinition{}
	if err := yaml.Unmarshal(crdBytes, crd); err != nil {
		return false
	}

	return crd.Spec.Scope == apiextensions.ClusterScoped
}
