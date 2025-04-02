package actions

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type resourceCreateActionConfig struct {
	Resource map[string]interface{} `yaml:"resource"`
}

func NewResourceCreate(kymaConfig *cmdcommon.KymaConfig, actionConfig map[string]interface{}) *cobra.Command {
	return &cobra.Command{
		Run: func(cmd *cobra.Command, _ []string) {
			cfg := resourceCreateActionConfig{}
			clierror.Check(parseActionConfig(actionConfig, &cfg))
			clierror.Check(createResource(kymaConfig, &cfg))
		},
	}
}

func createResource(kymaConfig *cmdcommon.KymaConfig, cfg *resourceCreateActionConfig) clierror.Error {
	u := &unstructured.Unstructured{
		Object: cfg.Resource,
	}

	client, clierr := kymaConfig.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	err := client.RootlessDynamic().Apply(kymaConfig.Ctx, u)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create resource"))
	}

	fmt.Printf("resource %s applied\n", u.GetName())
	return nil
}
