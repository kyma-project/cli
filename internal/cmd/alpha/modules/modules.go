package modules

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/communitymodules"
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
		PreRun: func(_ *cobra.Command, _ []string) {
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

	cmd.MarkFlagsMutuallyExclusive("catalog", "managed", "installed")

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
	catalog, err := communitymodules.ModulesCatalog()
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get all Kyma catalog"))
	}
	installedWith, err := communitymodules.InstalledModules(cfg.KubeClientConfig, *cfg.KymaConfig)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get installed Kyma modules"))
	}
	managedWith, err := communitymodules.ManagedModules(cfg.KubeClientConfig, *cfg.KymaConfig)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get managed Kyma modules"))
	}

	combinedData := communitymodules.MergeRowMaps(catalog, installedWith, managedWith)

	communitymodules.RenderModules(cfg.raw, combinedData, communitymodules.CollectiveTableInfo)
	return nil
}

// listInstalledModules lists all installed modules
func listInstalledModules(cfg *modulesConfig) clierror.Error {
	installed, err := communitymodules.InstalledModules(cfg.KubeClientConfig, *cfg.KymaConfig)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get installed Kyma modules"))
	}

	communitymodules.RenderModules(cfg.raw, installed, communitymodules.InstalledTableInfo)
	return nil
}

// listManagedModules lists all managed modules
func listManagedModules(cfg *modulesConfig) clierror.Error {
	managed, err := communitymodules.ManagedModules(cfg.KubeClientConfig, *cfg.KymaConfig)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get managed Kyma modules"))
	}

	communitymodules.RenderModules(cfg.raw, managed, communitymodules.ManagedTableInfo)
	return nil
}

// listModulesCatalog lists all available modules
func listModulesCatalog(cfg *modulesConfig) clierror.Error {
	catalog, err := communitymodules.ModulesCatalog()
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get all Kyma catalog"))
	}

	communitymodules.RenderModules(cfg.raw, catalog, communitymodules.CatalogTableInfo)
	return nil
}
