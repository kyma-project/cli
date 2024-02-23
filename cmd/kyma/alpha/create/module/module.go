package module

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/kyma-project/lifecycle-manager/api/shared"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	compdescv2 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/v2"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/internal/nice"
	"github.com/kyma-project/cli/pkg/module"
)

//go:embed cmd_description.txt
var description string

//go:embed cmd_example.txt
var example string

type (
	moduleBuilderService interface {
		BuildModule() (*Config, *module.Definition, error)
	}

	componentDescriptorService interface {
		Provision() error
	}

	imagePusherService interface {
		Push() error
	}

	signerService interface {
		SignModuleDefinition() error
	}

	templateGeneratorService interface {
		GenerateModuleTemplate() error
	}

	command struct {
		cli.Command
		opts   *Options
		logger *zap.SugaredLogger

		moduleBuilder       moduleBuilderService
		componentDescriptor componentDescriptorService
		imagePusher         imagePusherService
		signer              signerService
		templateGenerator   templateGeneratorService
	}
)

// NewCmd creates a new Kyma CLI command
func NewCmd(o *Options) *cobra.Command {
	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:     "module [--module-config-file MODULE_CONFIG_FILE | --name MODULE_NAME --version MODULE_VERSION] [--path MODULE_DIRECTORY] [--registry MODULE_REGISTRY] [flags]",
		Short:   "Creates a module bundled as an OCI artifact",
		Long:    description,
		Example: example,
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

	cmd.Flags().BoolVar(
		&o.ArchiveVersionOverwrite, "module-archive-version-overwrite", false,
		"Overwrites existing component's versions of the module. If set to false, the push is a No-Op.",
	)

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

	cmd.Flags().BoolVar(
		&o.KubebuilderProject, "kubebuilder-project", false,
		"Specifies provided module is a Kubebuilder Project.",
	)

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

	cmd.Flags().StringVar(
		&o.Namespace, "namespace", kcpSystemNamespace,
		"Specifies the namespace where the ModuleTemplate is deployed.",
	)

	return cmd
}

type validator interface {
	GetCrd() []byte
	Run(ctx context.Context, log *zap.SugaredLogger) error
}

const kcpSystemNamespace = "kcp-system"

func (cmd *command) Run(ctx context.Context) error {
	var (
		modCnf                 *Config
		modDef                 *module.Definition
		crValidator            validator
		componentVersionAccess ocm.ComponentVersionAccess
		archive                *comparch.ComponentArchive
		remote                 *module.Remote
		err                    error
		osFS                   = osfs.New()
	)

	cmd.opts.Synchronise()

	modCnf, modDef, err = cmd.moduleBuilder.BuildModule()
	{
		modDef, modCnf, err = cmd.moduleDefinitionFromOptions()

		if err != nil {
			return err
		}

		cmd.NewStep("Parse and build module...")

		// Create base resource defs with module root and its sub-layers
		if cmd.opts.KubebuilderProject {
			if err := module.InspectLegacy(modDef, cmd.opts.ResourcePaths, cmd.CurrentStep, cmd.logger); err != nil {
				cmd.CurrentStep.Failure()
				return err
			}
		} else {
			if err := module.Inspect(modDef, cmd.logger); err != nil {
				cmd.CurrentStep.Failure()
				return err
			}
		}
		cmd.CurrentStep.Successf("Module built")

		if crValidator, err = cmd.validateDefaultCR(ctx, modDef, cmd.logger); err != nil {
			return err
		}
	}

	err = cmd.componentDescriptor.Provision()
	{
		var archiveFS vfs.FileSystem
		if cmd.opts.PersistentArchive {
			archiveFS = osFS
			cmd.logger.Info("using host filesystem for archive")
		} else {
			archiveFS = memoryfs.New()
			cmd.logger.Info("using in-memory archive")
		}

		cmd.NewStep("Creating component descriptor")
		componentDescriptor, err := module.InitComponentDescriptor(modDef)
		if err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
		cmd.CurrentStep.Successf("component descriptor created")

		gitPath, err := files.SearchForTargetDirByName(modDef.Source, ".git")
		if gitPath == "" || err != nil {
			cmd.logger.Warnf("could not find git repository root, using %s directory", modDef.Source)
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

		} else {
			cmd.logger.Infof("found git repository root at %s: adding git sources to the layer", gitPath)
			modDef.Source = gitPath // set the source to the git root
			if err := module.AddGitSources(componentDescriptor, modDef, cmd.opts.GitRemote); err != nil {
				cmd.CurrentStep.Failure()
				return err
			}
		}

		// Security Scan
		if cmd.opts.SecurityScanConfig != "" && gitPath != "" { // security scan is only supported for target git repositories
			cmd.NewStep("Configuring security scanning...")
			if files.IsFileExists(cmd.opts.SecurityScanConfig) {
				err = module.AddSecurityScanningMetadata(componentDescriptor, cmd.opts.SecurityScanConfig)
				if err != nil {
					cmd.CurrentStep.Failure()
					return err
				}
				cmd.CurrentStep.Successf("Security scanning configured")
			} else {
				cmd.logger.Warnf("Security scanning configuration was skipped")
				cmd.CurrentStep.Failure()
			}
		}

		cmd.NewStep("Creating module archive...")
		archive, err = module.CreateArchive(archiveFS, cmd.opts.ModuleArchivePath, componentDescriptor)
		if err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
		cmd.CurrentStep.Successf("Module archive created")

		cmd.NewStep("Adding layers to archive...")
		if err = module.AddResources(archive, modDef, cmd.logger, osFS, cmd.opts.RegistryCredSelector); err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
		cmd.CurrentStep.Successf("Layers successfully added to archive")
	}

	err = cmd.imagePusher.Push()
	{
		if cmd.opts.RegistryURL == "" {
			return nil
		}

		cmd.NewStep(fmt.Sprintf("Pushing image to %q", cmd.opts.RegistryURL))
		remote, err = cmd.getRemote(modDef.NameMappingMode)
		if err != nil {
			cmd.CurrentStep.Failure()
			return err
		}

		repo, err := remote.GetRepository(cpi.DefaultContext())
		if err != nil {
			cmd.CurrentStep.Failure()
			return err
		}

		shouldPushArchive, err := remote.ShouldPushArchive(repo, archive, cmd.opts.ArchiveVersionOverwrite)
		if err != nil {
			cmd.CurrentStep.Failure()
			return err
		}

		if shouldPushArchive {
			componentVersionAccess, err = remote.Push(repo, archive, cmd.opts.ArchiveVersionOverwrite)
			if err != nil {
				cmd.CurrentStep.Failure()
				return err
			}
			cmd.CurrentStep.Successf("Module successfully pushed")
		} else {
			componentVersionAccess, err = remote.GetComponentVersion(archive, repo)
			if err != nil {
				cmd.CurrentStep.Failure()
				return err
			}
			cmd.CurrentStep.Successf(
				fmt.Sprintf(
					"Module already exists. Retrieved image from %q",
					cmd.opts.RegistryURL,
				),
			)
		}
	}

	err = cmd.signer.SignModuleDefinition()
	{
		if cmd.opts.PrivateKeyPath != "" {
			cmd.NewStep("Fetching and signing component descriptor...")
			signConfig := &module.ComponentSignConfig{
				Name:    modDef.Name,
				Version: modDef.Version,
				KeyPath: cmd.opts.PrivateKeyPath,
			}
			if err = module.Sign(signConfig, remote); err != nil {
				cmd.CurrentStep.Failure()
				return err
			}
			cmd.CurrentStep.Success()
		}
	}

	err = cmd.templateGenerator.GenerateModuleTemplate()
	{
		cmd.NewStep("Generating module template...")
		var resourceName = ""
		mandatoryModule := false
		var channel = cmd.opts.Channel
		if modCnf != nil {
			resourceName = modCnf.ResourceName
			channel = modCnf.Channel
			mandatoryModule = modCnf.Mandatory
		}

		var namespace = cmd.opts.Namespace
		if modCnf != nil && modCnf.Namespace != "" {
			namespace = modCnf.Namespace

		}

		labels := cmd.getModuleTemplateLabels(modCnf)
		annotations := cmd.getModuleTemplateAnnotations(modCnf, crValidator)

		template, err := module.Template(
			componentVersionAccess, resourceName, namespace,
			channel, modDef.DefaultCR, labels, annotations, modDef.CustomStateChecks, mandatoryModule,
		)
		if err != nil {
			cmd.CurrentStep.Failure()
			return err
		}

		if err := vfs.WriteFile(osFS, cmd.opts.TemplateOutput, template, os.ModePerm); err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
		cmd.CurrentStep.Successf("Template successfully generated at %s", cmd.opts.TemplateOutput)
	}

	return nil
}

func (cmd *command) getModuleTemplateLabels(modCnf *Config) map[string]string {
	labels := map[string]string{}
	if modCnf != nil {
		maps.Copy(labels, modCnf.Labels)

		if modCnf.Beta {
			labels[shared.BetaLabel] = shared.EnableLabelValue
		}
		if modCnf.Internal {
			labels[shared.InternalLabel] = shared.EnableLabelValue
		}
	}

	return labels
}

func (cmd *command) getModuleTemplateAnnotations(modCnf *Config, crValidator validator) map[string]string {
	annotations := map[string]string{}
	moduleVersion := cmd.opts.Version
	if modCnf != nil {
		maps.Copy(annotations, modCnf.Annotations)

		moduleVersion = modCnf.Version
	}

	isClusterScoped := isCrdClusterScoped(crValidator.GetCrd())
	if isClusterScoped {
		annotations[shared.IsClusterScopedAnnotation] = shared.EnableLabelValue
	} else {
		annotations[shared.IsClusterScopedAnnotation] = shared.DisableLabelValue
	}
	annotations[shared.ModuleVersionAnnotation] = moduleVersion
	return annotations
}

func (cmd *command) validateDefaultCR(ctx context.Context, modDef *module.Definition, l *zap.SugaredLogger) (validator,
	error) {
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
		Registry:      cmd.opts.RegistryURL,
		NameMapping:   nameMapping,
		Credentials:   cmd.opts.Credentials,
		Token:         cmd.opts.Token,
		Insecure:      cmd.opts.Insecure,
		OciRepoAccess: &module.OciRepo{},
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

// Does not belong to the command object.
// Module definition creation can be used elsewhere.
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

		// legacy approach, flag-based
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

	// new approach, config-file  based
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
