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

	err = defaultView(cfg)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to list modules"))
	}

	return nil
}

func defaultView(cfg *modulesConfig) clierror.Error {
	catalog, err := model.GetAllModules(nil)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get all Kyma catalog"))
	}
	installedWith, err := model.GetInstalledModules(catalog, cfg.KubeClientConfig, *cfg.KymaConfig)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get installed Kyma modules"))
	}
	managedWith, err := model.GetManagedModules(installedWith, cfg.KubeClientConfig, *cfg.KymaConfig)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get managed Kyma modules"))
	}

	var table [][]string
	for _, row := range managedWith {
		table = append(table, row)
	}

	twTable := model.SetTable(table)
	twTable.SetHeader([]string{"NAME", "REPOSITORY", "VERSION", "CONTROL-PLANE"})
	twTable.Render()

	return nil
}

func listInstalledModules(cfg *modulesConfig) clierror.Error {
	installed, err := model.GetInstalledModules(nil, cfg.KubeClientConfig, *cfg.KymaConfig)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get installed Kyma modules"))
	}

	var table [][]string
	for _, row := range installed {
		table = append(table, row)
	}

	twTable := model.SetTable(table)
	twTable.SetHeader([]string{"NAME", "VERSION"})
	twTable.Render()

	return nil
}

func listManagedModules(cfg *modulesConfig) clierror.Error {
	managed, err := model.GetManagedModules(nil, cfg.KubeClientConfig, *cfg.KymaConfig)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get managed Kyma modules"))
	}
	fmt.Println("Managed modules:\n")
	fmt.Println(managed)

	var table [][]string
	for _, row := range managed {
		table = append(table, row)
	}

	twTable := model.SetTable(table)
	twTable.SetHeader([]string{"NAME"})
	twTable.Render()

	return nil
}

func listAllModules() clierror.Error {
	catalog, err := model.GetAllModules(nil)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to get all Kyma catalog"))
	}

	var table [][]string
	for _, row := range catalog {
		table = append(table, row)
	}

	twTable := model.SetTable(table)
	twTable.SetHeader([]string{"NAME", "REPOSITORY"})
	twTable.Render()
	return nil
}
