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

type resourceDeleteAction struct {
	configurator[resourceDeleteActionConfig]

	kymaConfig *cmdcommon.KymaConfig
}

func NewResourceDelete(kymaConfig *cmdcommon.KymaConfig) types.Action {
	return &resourceDeleteAction{
		kymaConfig: kymaConfig,
	}
}

func (a *resourceDeleteAction) Run(cmd *cobra.Command, _ []string) clierror.Error {
	u := &unstructured.Unstructured{
		Object: a.cfg.Resource,
	}

	client, clierr := a.kymaConfig.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	err := client.RootlessDynamic().Remove(a.kymaConfig.Ctx, u)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to delete resource"))
	}

	fmt.Fprintf(cmd.OutOrStdout(), "resource %s deleted\n", u.GetName())

	return nil
}
