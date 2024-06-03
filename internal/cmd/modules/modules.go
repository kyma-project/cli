package modules

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"io"
	"net/http"

	"github.com/kyma-project/cli.v3/internal/clierror"
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

	config := modulesConfig{
		KymaConfig:       kymaConfig,
		KubeClientConfig: cmdcommon.KubeClientConfig{},
	}

	cmd := &cobra.Command{
		Use:   "modules",
		Short: "List modules.",
		Long:  `List either installed, managed or available Kyma modules.`,
		PreRun: func(_ *cobra.Command, args []string) {
			clierror.Check(config.KubeClientConfig.Complete())
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runModules(&config))
		},
	}

	cmd.Flags().BoolVar(&config.catalog, "catalog", false, "List of al available Kyma modules.")
	cmd.Flags().BoolVar(&config.managed, "managed", false, "List of all Kyma modules managed by central control-plane.")
	cmd.Flags().BoolVar(&config.installed, "installed", false, "List of all currently installed Kyma modules.")

	cmd.MarkFlagsOneRequired("catalog", "managed", "installed")
	cmd.MarkFlagsMutuallyExclusive("catalog", "managed")
	cmd.MarkFlagsMutuallyExclusive("catalog", "installed")
	cmd.MarkFlagsMutuallyExclusive("managed", "installed")

	return cmd
}

func runModules(config *modulesConfig) clierror.Error {
	var err error
	if config.catalog {
		modules, err := listAllModules()
		if err != nil {
			return clierror.WrapE(err, clierror.New("failed to list all Kyma modules"))
		}
		fmt.Println("Available modules:\n")
		for _, rec := range modules {
			fmt.Println(rec)
		}
		return nil
	}

	if config.managed {
		_, err := listManagedModules(config)
		clierror.WrapE(err, clierror.New("not implemented yet, please use the catalog flag"))
		return nil
	}

	if config.installed {
		clierror.Wrap(err, clierror.New("not implemented yet, please use the catalog flag"))
		return nil
	}
	//TODO: installed to implement

	return clierror.Wrap(err, clierror.New("failed to get modules", "please use one of: catalog, managed or installed flags"))
}

func listAllModules() ([]string, clierror.Error) {
	resp, err := http.Get("https://raw.githubusercontent.com/kyma-project/community-modules/main/model.json")
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while getting modules list"))
	}
	defer resp.Body.Close()

	var template []struct {
		Name string `json:"name"`
	}

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while reading http response"))
	}
	err = json.Unmarshal(bodyText, &template)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("while unmarshalling"))
	}

	var out []string
	for _, rec := range template {
		out = append(out, rec.Name)
	}
	return out, nil
}

func listManagedModules(config *modulesConfig) ([]string, clierror.Error) {
	trololo := config.KubeClient.Static().CoreV1().RESTClient().Get().AbsPath("kyma-project.io")
	fmt.Println(trololo)
	return nil, clierror.New("chleb")
}
