package modules

import (
	"fmt"
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

	cmd.MarkFlagsOneRequired("catalog", "managed", "installed")
	cmd.MarkFlagsMutuallyExclusive("catalog", "managed")
	cmd.MarkFlagsMutuallyExclusive("catalog", "installed")
	cmd.MarkFlagsMutuallyExclusive("managed", "installed")

	return cmd
}

func listModules(cfg *modulesConfig) clierror.Error {
	var err clierror.Error

	if cfg.catalog {
		err = listAllModules()
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

	return clierror.WrapE(err, clierror.New("failed to get modules", "please use one of: catalog, managed or installed flags"))
}

func listInstalledModules(cfg *modulesConfig) clierror.Error {
	installed, err := model.GetInstalledModules(cfg.KubeClientConfig, *cfg.KymaConfig)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get installed Kyma modules"))
	}
	fmt.Println("Installed modules:\n")

	installed.SetHeader([]string{"NAME", "VERSION"})
	installed.Render()
	return nil
}

func listManagedModules(cfg *modulesConfig) clierror.Error {
	managed, err := model.GetManagedModules(cfg.KubeClientConfig, *cfg.KymaConfig)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get managed Kyma modules"))
	}
	fmt.Println("Managed modules:\n")
	managed.SetHeader([]string{"NAME"})
	managed.Render()
	return nil
}

func listAllModules() clierror.Error {
	catalog, err := model.GetAllModules()
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get all Kyma catalog"))
	}
	fmt.Println("Available catalog:\n")

	catalog.SetHeader([]string{"NAME", "REPOSITORY"})
	catalog.Render()
	return nil
}
