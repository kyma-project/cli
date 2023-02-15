package module

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kustomize"
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
		Use:   "module [flags]",
		Short: "Creates a module bundled as an OCI image with the given OCI image name from the contents of the given path",
		Long: `Use this command to create a Kyma module and bundle it as an OCI image.

### Detailed description

Kyma modules are individual components that can be deployed into a Kyma runtime. Modules are built and distributed as OCI container images. 
With this command, you can create such images out of a folder's contents.

This command creates a component descriptor in the descriptor path (./mod as a default) and packages all the contents on the provided path as an OCI image.
Kubebuilder projects are supported. If the path contains a kubebuilder project, it will be built and pre-defined layers will be created based on its known contents.

Alternatively, a custom (non kubebuilder) module can be created by providing a path that does not contain a kubebuilder project. In that case all the contents of the path will be bundled as a single layer.

Optionally, you can manually add additional layers with contents in other paths (see [resource flag](#flags) for more information).

Finally, if you provided a registry to which to push the artifact, the created module is validated and pushed. During the validation the default CR defined in the optional "default.yaml" file is validated against CustomResourceDefinition.
Alternatively, you can trigger an on-demand default CR validation with "--validateCR=true", in case you don't push to the registry.

To push the artifact into some registries, for example, the central docker.io registry, you have to change the OCM Component Name Mapping with the following flag: "--nameMapping=sha256-digest". This is necessary because the registry does not accept artifact URLs with more than two path segments, and such URLs are generated with the default name mapping: "urlPath". In the case of the "sha256-digest" mapping, the artifact URL contains just a sha256 digest of the full Component Name and fits the path length restrictions.

`,

		Example: `Examples:
Build module my-domain/modA in version 1.2.3 and push it to a remote registry
		kyma alpha create module -n my-domain/modA --version 1.2.3 -p /path/to/module --registry https://dockerhub.com
Build module my-domain/modB in version 3.2.1 and push it to a local registry "unsigned" subfolder without tls
		kyma alpha create module -n my-domain/modB --version 3.2.1 -p /path/to/module --registry http://localhost:5001/unsigned --insecure
`,
		RunE:    func(_ *cobra.Command, args []string) error { return c.Run(args) },
		Aliases: []string{"mod"},
	}

	cmd.Flags().StringVar(&o.Version, "version", "", "Version of the module. This flag is mandatory.")
	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "Override the module name of the kubebuilder project. If the module is not a kubebuilder project, this flag is mandatory.")
	cmd.Flags().StringVarP(&o.Path, "path", "p", "", "Path to the module contents. (default current directory)")
	cmd.Flags().StringVar(&o.ModCache, "mod-cache", "./mod", "Specifies the path where the module artifacts are locally cached to generate the image. If the path already has a module, use the overwrite flag to overwrite it.")
	cmd.Flags().StringArrayVarP(&o.ResourcePaths, "resource", "r", []string{}, "Add an extra resource in a new layer with format <NAME:TYPE@PATH>. It is also possible to provide only a path; name will default to the last path element and type to 'helm-chart'")
	cmd.Flags().StringVar(&o.RegistryURL, "registry", "", "Repository context url for module to upload. The repository url will be automatically added to the repository contexts in the module")
	cmd.Flags().StringVar(&o.NameMappingMode, "nameMapping", "urlPath", "Overrides the OCM Component Name Mapping, one of: \"urlPath\" or \"sha256-digest\"")
	cmd.Flags().StringVar(&o.RegistryCredSelector, "registry-cred-selector", "",
		"label selector to identify a secret of type kubernetes.io/dockerconfigjson (that needs to be created externally) which allows the image to be accessed in private image registries. This can be used if you push your module to a registry with authenticated access. Example: \"label1=value1,label2=value2\"")
	cmd.Flags().StringVarP(&o.Credentials, "credentials", "c", "", "Basic authentication credentials for the given registry in the format user:password")
	cmd.Flags().StringVar(&o.DefaultCRPath, "default-cr", "", "File containing the default custom resource of the module. If the module is a kubebuilder project, the default CR will be automatically detected.")
	cmd.Flags().StringVarP(&o.TemplateOutput, "output", "o", "template.yaml", "File to which to output the module template if the module is uploaded to a registry")
	cmd.Flags().StringVar(&o.Channel, "channel", "regular", "Channel to use for the module template.")
	cmd.Flags().StringVarP(&o.Token, "token", "t", "", "Authentication token for the given registry (alternative to basic authentication).")
	cmd.Flags().BoolVarP(&o.Overwrite, "overwrite", "w", false, "overwrites the existing mod-path directory if it exists")
	cmd.Flags().BoolVar(&o.Insecure, "insecure", false, "Use an insecure connection to access the registry.")
	cmd.Flags().BoolVar(&o.Clean, "clean", false, "Remove the mod-path folder and all its contents at the end.")
	cmd.Flags().StringVar(&o.SecurityScanConfig, "sec-scan-cfg", "", "Path to security scan configuration file.")

	return cmd
}

func (cmd *command) Run(args []string) error {

	if !cmd.opts.NonInteractive {
		cli.AlphaWarn()
	}

	l := cli.NewLogger(cmd.opts.Verbose).Sugar()

	if err := cmd.opts.ValidatePath(); err != nil {
		return err
	}

	if err := cmd.opts.ValidateChannel(); err != nil {
		return err
	}

	nameMappingMode, err := module.ParseNameMapping(cmd.opts.NameMappingMode)
	if err != nil {
		return err
	}

	modDef := &module.Definition{
		Name:            cmd.opts.Name,
		Version:         cmd.opts.Version,
		Source:          cmd.opts.Path,
		ArchivePath:     cmd.opts.ModCache,
		Overwrite:       cmd.opts.Overwrite,
		RegistryURL:     cmd.opts.RegistryURL,
		NameMappingMode: nameMappingMode,
		DefaultCRPath:   cmd.opts.DefaultCRPath,
	}

	cmd.NewStep("Setting up kustomize...")
	if err := kustomize.Setup(cmd.CurrentStep, true); err != nil {
		cmd.CurrentStep.Failure()
		return err
	}
	cmd.CurrentStep.Successf("Kustomize ready")

	/* -- Inspect and build Module -- */
	cmd.NewStep("Parse and build module...")

	// Create base resource defs with module root and its sub-layers
	if err := module.Inspect(modDef, cmd.opts.ResourcePaths, cmd.CurrentStep, l); err != nil {
		cmd.CurrentStep.Failure()
		return err
	}
	cmd.CurrentStep.Successf("Module built")

	/* -- VALIDATE DEFAULT CR -- */
	if err := cmd.validateDefaultCR(modDef, l); err != nil {
		return err
	}

	/* -- CREATE ARCHIVE -- */
	fs := osfs.New()

	cmd.NewStep(fmt.Sprintf("Creating module archive at %q", cmd.opts.ModCache))
	archive, err := module.Build(fs, modDef)
	if err != nil {
		cmd.CurrentStep.Failure()
		return err
	}
	cmd.CurrentStep.Success()

	/* -- Create Image -- */
	cmd.NewStep("Creating image...")

	if err := module.AddResources(archive, modDef, l, fs); err != nil {
		cmd.CurrentStep.Failure()
		return err
	}

	cmd.CurrentStep.Successf("Image created")

	/* -- ADD SECURITY SCANNING METADATA -- */
	if cmd.opts.SecurityScanConfig != "" {
		cmd.NewStep("Configuring security scanning...")
		err = module.AddSecurityScanningMetadata(archive.ComponentDescriptor, modDef, fs, cmd.opts.SecurityScanConfig)
		if err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
		cmd.CurrentStep.Successf("Security scanning configured")
	}
	/* -- PUSH & TEMPLATE -- */

	if cmd.opts.RegistryURL != "" {

		cmd.NewStep(fmt.Sprintf("Pushing image to %q", cmd.opts.RegistryURL))
		r, err := cmd.validateInsecureRegistry(modDef.NameMappingMode)
		if err != nil {
			cmd.CurrentStep.Failure()
			return err
		}

		if err := module.Push(archive, r, l); err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
		cmd.CurrentStep.Success()

		cmd.NewStep("Generating module template")
		t, err := module.Template(archive, cmd.opts.Channel, modDef.DefaultCR, cmd.opts.RegistryCredSelector)
		if err != nil {
			cmd.CurrentStep.Failure()
			return err
		}

		if err := vfs.WriteFile(fs, cmd.opts.TemplateOutput, t, os.ModePerm); err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
		cmd.CurrentStep.Success()
	}

	/* -- CLEANUP -- */

	if cmd.opts.Clean {
		// TODO clean generated chart
		cmd.NewStep(fmt.Sprintf("Cleaning up mod path %q", cmd.opts.ModCache))
		if err := os.RemoveAll(cmd.opts.ModCache); err != nil {
			cmd.CurrentStep.Failure()
			return err
		}
		cmd.CurrentStep.Success()
	}

	return nil
}

func (cmd *command) validateDefaultCR(modDef *module.Definition, l *zap.SugaredLogger) error {
	cmd.NewStep("Validating Default CR")
	crValidator, err := module.NewDefaultCRValidator(modDef.DefaultCR, modDef.Source)
	if err != nil {
		cmd.CurrentStep.Failure()
		return err
	}

	if err := crValidator.Run(cmd.CurrentStep, cmd.opts.Verbose, l); err != nil {
		if errors.Is(err, module.ErrEmptyCR) {
			cmd.CurrentStep.Successf("Default CR validation skipped - no default CR")
			return nil
		}
		return err
	}
	cmd.CurrentStep.Successf("Default CR validation succeeded")
	return nil
}

func (cmd *command) validateInsecureRegistry(nameMapping module.NameMapping) (*module.Remote, error) {

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

		if !cmd.opts.Insecure && !cmd.opts.NonInteractive {
			cmd.CurrentStep.LogWarn("CAUTION: Pushing the module artifact to the insecure registry")
			if !cmd.CurrentStep.PromptYesNo("Do you really want to proceed? ") {
				return nil, errors.New("Command stopped by user")
			}

		}
	}

	return res, nil
}
