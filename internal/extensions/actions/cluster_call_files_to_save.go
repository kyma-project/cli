package actions

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/extensions/actions/common"
	"github.com/kyma-project/cli.v3/internal/extensions/call"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/spf13/cobra"
)

type targetPodConfig struct {
	Selector  map[string]string `yaml:"selector"`
	Namespace string            `yaml:"namespace"`
	Port      string            `yaml:"port"`
}

type requestConfig struct {
	// TODO: support other HTTP methods
	// Method    string            `yaml:"method"`
	Path       string            `yaml:"path"`
	Parameters map[string]string `yaml:"parameters"`
}

type clusterCallFilesToSaveConfig struct {
	Request   requestConfig   `yaml:"request"`
	TargetPod targetPodConfig `yaml:"target"`
	OutputDir string          `yaml:"outputDir"`
}

func (c *clusterCallFilesToSaveConfig) validate() clierror.Error {
	if c.TargetPod.Namespace == "" {
		return clierror.New("empty target pod namespace")
	}
	if c.TargetPod.Selector == nil {
		return clierror.New("empty target pod selector")
	}
	if c.TargetPod.Port == "" {
		return clierror.New("empty target pod port")
	}
	if c.Request.Path == "" {
		return clierror.New("empty request path")
	}
	if c.OutputDir == "" {
		return clierror.New("empty output directory path")
	}

	return nil
}

type clusterCallFilesToSaveAction struct {
	common.TemplateConfigurator[clusterCallFilesToSaveConfig]

	kymaConfig *cmdcommon.KymaConfig
}

func NewClusterCallFilesToSaveAction(kymaConfig *cmdcommon.KymaConfig) types.Action {
	return &clusterCallFilesToSaveAction{
		kymaConfig: kymaConfig,
	}
}

func (a *clusterCallFilesToSaveAction) Run(cmd *cobra.Command, _ []string) clierror.Error {
	clierr := a.Cfg.validate()
	if clierr != nil {
		return clierror.WrapE(clierr, clierror.New("invalid action configuration"))
	}

	client, clierr := a.kymaConfig.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	podCaller := call.NewPodCaller(
		client,
		a.Cfg.TargetPod.Namespace,
		a.Cfg.TargetPod.Selector,
		a.Cfg.TargetPod.Port,
	)

	bytesResp, clierr := podCaller.Get(a.kymaConfig.Ctx, a.Cfg.Request.Path, a.Cfg.Request.Parameters)
	if clierr != nil {
		return clierror.WrapE(clierr, clierror.New("failed to call target pod"))
	}

	// Process the response
	var filesResp call.FilesListResponse
	if err := json.Unmarshal(bytesResp, &filesResp); err != nil {
		return clierror.Wrap(err, clierror.New("failed to decode response"))
	}

	data, err := base64.StdEncoding.DecodeString(filesResp.Files[0].Data)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to decode file data"))
	}

	fmt.Println(string(data))

	return nil
}
