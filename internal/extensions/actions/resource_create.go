package actions

import (
	"encoding/json"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	cmd_types "github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/extensions/actions/common"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type resourceCreateActionConfig struct {
	DryRun        bool                   `yaml:"dryRun"`
	Output        cmd_types.Format       `yaml:"output"`
	OutputWarning string                 `yaml:"outputWarning"`
	OutputMessage string                 `yaml:"outputMessage"`
	Resource      map[string]interface{} `yaml:"resource"`
}

type resourceCreateAction struct {
	common.TemplateConfigurator[resourceCreateActionConfig]

	kymaConfig *cmdcommon.KymaConfig
}

func NewResourceCreate(kymaConfig *cmdcommon.KymaConfig) types.Action {
	return &resourceCreateAction{
		kymaConfig: kymaConfig,
	}
}

func (a *resourceCreateAction) Run(cmd *cobra.Command, _ []string) clierror.Error {
	u := &unstructured.Unstructured{
		Object: a.Cfg.Resource,
	}

	client, clierr := a.kymaConfig.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	if a.Cfg.OutputWarning != "" {
		out.Errln(a.Cfg.OutputWarning)
	}

	err := client.RootlessDynamic().Apply(a.kymaConfig.Ctx, u, a.Cfg.DryRun)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create resource"))
	}

	output, err := a.formatOutput(u)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to format output"))
	}

	out.Msgln(output)

	if a.Cfg.OutputMessage != "" && a.Cfg.Output == cmd_types.DefaultFormat {
		out.Msgln(a.Cfg.OutputMessage)
	}

	return nil
}

func (a *resourceCreateAction) formatOutput(u *unstructured.Unstructured) (string, error) {
	if a.Cfg.Output == cmd_types.JSONFormat {
		obj, err := json.MarshalIndent(u.Object, "", "  ")
		return string(obj), err
	}

	if a.Cfg.Output == cmd_types.YAMLFormat {
		obj, err := yaml.Marshal(u.Object)
		return string(obj), err
	}

	messageSuffix := ""
	if a.Cfg.DryRun {
		messageSuffix = " (dry run)"
	}

	return fmt.Sprintf("resource %s applied%s", u.GetName(), messageSuffix), nil
}
