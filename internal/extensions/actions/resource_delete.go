package actions

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type resourceDeleteActionConfig struct {
	Resource map[string]interface{} `yaml:"resource"`
}

func NewResourceDelete(kymaConfig *cmdcommon.KymaConfig, actionConfig types.ActionConfig) types.CmdRun {
	return func(cmd *cobra.Command, args []string) {
		cfg := resourceDeleteActionConfig{}
		clierror.Check(parseActionConfig(actionConfig, &cfg))
		clierror.Check(deleteResource(kymaConfig, &cfg))
	}
}

func deleteResource(kymaConfig *cmdcommon.KymaConfig, cfg *resourceDeleteActionConfig) clierror.Error {
	u := &unstructured.Unstructured{
		Object: cfg.Resource,
	}

	client, clierr := kymaConfig.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	err := client.RootlessDynamic().Remove(kymaConfig.Ctx, u)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to delete resource"))
	}

	fmt.Printf("resource %s deleted\n", u.GetName())

	return nil
}
