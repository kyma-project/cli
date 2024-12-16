package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube/resources"
	"github.com/kyma-project/cli.v3/internal/modules"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type enableConfig struct {
	*cmdcommon.KymaConfig

	module    string
	channel   string
	crPath    string
	defaultCR bool
}

func newEnableCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := enableConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "enable <module>",
		Short: "Enable module.",
		Long:  "Use this command to enable module.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg.module = args[0]
			clierror.Check(runEnable(&cfg))
		},
	}

	cmd.Flags().StringVar(&cfg.channel, "channel", "", "Name of the Kyma channel to use for the module")
	cmd.Flags().StringVar(&cfg.crPath, "cr-path", "", "Path to the custom resource file")
	cmd.Flags().BoolVar(&cfg.defaultCR, "default-cr", false, "Use this flag to deploy module with default cr")

	cmd.MarkFlagsMutuallyExclusive("cr-path", "default-cr")

	return cmd
}

func runEnable(cfg *enableConfig) clierror.Error {
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

	for i := range crs {
		// make sure all resources have namespace set
		if crs[i].GetNamespace() == "" {
			crs[i].SetNamespace("default")
		}
	}

	return crs, nil
}
