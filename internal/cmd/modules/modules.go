package modules

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/model"
	"github.com/spf13/cobra"
)

type modulesConfig struct {
	*cmdcommon.KymaConfig
	cmdcommon.KubeClientConfig

	catalog   bool
	managed   bool
	installed bool
	raw       bool
}

func NewModulesCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := modulesConfig{
		KymaConfig:       kymaConfig,
		KubeClientConfig: cmdcommon.KubeClientConfig{},
	}

	cmd := &cobra.Command{
		Use:   "modules",
		Short: "List modules.",
		Long:  `List either installed, managed or available Kyma modules.`,
		PreRun: func(_ *cobra.Command, args []string) {
			clierror.Check(cfg.KubeClientConfig.Complete())
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(listModules(&cfg))
		},
	}
	cfg.KubeClientConfig.AddFlag(cmd)

	cmd.Flags().BoolVar(&cfg.catalog, "catalog", false, "List of al available Kyma modules.")
	cmd.Flags().BoolVar(&cfg.managed, "managed", false, "List of all Kyma modules managed by central control-plane.")
	cmd.Flags().BoolVar(&cfg.installed, "installed", false, "List of all currently installed Kyma modules.")
	cmd.Flags().BoolVar(&cfg.raw, "raw", false, "Simple output format without table rendering.")

	cmd.MarkFlagsMutuallyExclusive("catalog", "managed")
	cmd.MarkFlagsMutuallyExclusive("catalog", "installed")
	cmd.MarkFlagsMutuallyExclusive("managed", "installed")

	return cmd
}

// listModules collects all the methods responsible for the command and its flags
func listModules(cfg *modulesConfig) clierror.Error {
	var err clierror.Error

	if cfg.catalog {
		err = listModulesCatalog(cfg)
		if err != nil {
			return clierror.WrapE(err, clierror.New("failed to list all Kyma modules"))
		}
		return nil
	}

	if cfg.managed {
		err = listManagedModules(cfg)
		if err != nil {
			return clierror.WrapE(err, clierror.New("failed to list managed Kyma modules"))
		}
		return nil
	}

	if cfg.installed {
		err = listInstalledModules(cfg)
		if err != nil {
			return clierror.WrapE(err, clierror.New("failed to list installed Kyma modules"))
		}
		return nil
	}

	err = collectiveView(cfg)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to list modules"))
	}

	return nil
}

// collectiveView combines the list of all available, installed and managed modules
func collectiveView(cfg *modulesConfig) clierror.Error {
	catalog, err := model.ModulesCatalog(nil)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get all Kyma catalog"))
	}
	installedWith, err := model.InstalledModules(catalog, cfg.KubeClientConfig, *cfg.KymaConfig)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get installed Kyma modules"))
	}
	managedWith, err := model.ManagedModules(installedWith, cfg.KubeClientConfig, *cfg.KymaConfig)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get managed Kyma modules"))
	}

	model.RenderTable(cfg.raw, managedWith, []string{"NAME", "REPOSITORY", "VERSION INSTALLED", "CONTROL-PLANE"})

	return nil
}

// listInstalledModules lists all installed modules
func listInstalledModules(cfg *modulesConfig) clierror.Error {
	installed, err := model.InstalledModules(nil, cfg.KubeClientConfig, *cfg.KymaConfig)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get installed Kyma modules"))
	}

	model.RenderTable(cfg.raw, installed, []string{"NAME", "VERSION"})

	return nil
}

// listManagedModules lists all managed modules
func listManagedModules(cfg *modulesConfig) clierror.Error {
	managed, err := model.ManagedModules(nil, cfg.KubeClientConfig, *cfg.KymaConfig)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get managed Kyma modules"))
	}

	model.RenderTable(cfg.raw, managed, []string{"NAME"})

	return nil
}

// listModulesCatalog lists all available modules
func listModulesCatalog(cfg *modulesConfig) clierror.Error {
	catalog, err := model.ModulesCatalog(nil)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get all Kyma catalog"))
	}

	model.RenderTable(cfg.raw, catalog, []string{"NAME", "REPOSITORY"})
	return nil
}
