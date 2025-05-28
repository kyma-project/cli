package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/kyma-project/cli.v3/internal/kube/resources"
	"github.com/kyma-project/cli.v3/internal/modules"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type addConfig struct {
	*cmdcommon.KymaConfig

	module    string
	channel   string
	crPath    string
	defaultCR bool
}

func newAddCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := addConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "add <module> [flags]",
		Short: "Add a module",
		Long:  "Use this command to add a module.",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, _ []string) {
			clierror.Check(flags.Validate(cmd.Flags(),
				flags.MarkMutuallyExclusive("cr-path", "default-cr"),
			))
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg.module = args[0]
			clierror.Check(runAdd(&cfg))
		},
	}

	cmd.Flags().StringVarP(&cfg.channel, "channel", "c", "", "Name of the Kyma channel to use for the module")
	cmd.Flags().StringVar(&cfg.crPath, "cr-path", "", "Path to the custom resource file")
	cmd.Flags().BoolVar(&cfg.defaultCR, "default-cr", false, "Deploys the module with the default CR")

	return cmd
}

func runAdd(cfg *addConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	crs, clierr := loadCustomCRs(cfg.crPath)
	if clierr != nil {
		return clierr
	}

	return modules.Enable(cfg.Ctx, client, cfg.module, cfg.channel, cfg.defaultCR, crs...)
}

func loadCustomCRs(crPath string) ([]unstructured.Unstructured, clierror.Error) {
	if crPath == "" {
		// skip if not set
		return nil, nil
	}

	crs, err := resources.ReadFromFiles(crPath)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("failed to read object from file"))
	}

	return crs, nil
}
