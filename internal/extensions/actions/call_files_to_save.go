package actions

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
	Path      string            `yaml:"path"`
}

type requestConfig struct {
	// TODO: support other HTTP methods
	// Method    string            `yaml:"method"`
	Parameters map[string]string `yaml:"parameters"`
}

type callFilesToSaveConfig struct {
	Request   requestConfig   `yaml:"request"`
	TargetPod targetPodConfig `yaml:"targetPod"`
	OutputDir string          `yaml:"outputDir"`
}

func (c *callFilesToSaveConfig) validate() clierror.Error {
	if c.TargetPod.Namespace == "" {
		return clierror.New("empty target pod namespace")
	}
	if c.TargetPod.Selector == nil {
		return clierror.New("empty target pod selector")
	}
	if c.TargetPod.Port == "" {
		return clierror.New("empty target pod port")
	}
	if c.TargetPod.Path == "" {
		return clierror.New("empty target pod path")
	}
	if c.OutputDir == "" {
		return clierror.New("empty output directory path")
	}

	return nil
}

type callFilesToSaveAction struct {
	common.TemplateConfigurator[callFilesToSaveConfig]

	kymaConfig *cmdcommon.KymaConfig
}

func NewCallFilesToSaveAction(kymaConfig *cmdcommon.KymaConfig) types.Action {
	return &callFilesToSaveAction{
		kymaConfig: kymaConfig,
	}
}

func (a *callFilesToSaveAction) Run(cmd *cobra.Command, _ []string) clierror.Error {
	clierr := a.Cfg.validate()
	if clierr != nil {
		return clierror.WrapE(clierr, clierror.New("invalid action configuration"))
	}

	if !filepath.IsLocal(a.Cfg.OutputDir) {
		// output dir is not a local path, ask user for confirmation
		clierr = getUserAcceptance(a.Cfg.OutputDir)
		if clierr != nil {
			return clierr
		}
	}

	client, clierr := a.kymaConfig.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	podCaller := call.NewPodCaller(
		a.kymaConfig.Ctx,
		client,
		a.Cfg.TargetPod.Namespace,
		a.Cfg.TargetPod.Selector,
		a.Cfg.TargetPod.Port,
	)

	bytesResp, clierr := podCaller.Call("GET", a.Cfg.TargetPod.Path, a.Cfg.Request.Parameters)
	if clierr != nil {
		return clierror.WrapE(clierr, clierror.New("failed to call server"))
	}

	var filesResp call.FilesListResponse
	if err := json.Unmarshal(bytesResp, &filesResp); err != nil {
		return clierror.Wrap(err, clierror.New("failed to decode server response"))
	}

	outDir, err := filepath.Abs(a.Cfg.OutputDir)
	if err != nil {
		// undexpected error, use realtive path
		outDir = a.Cfg.OutputDir
	}

	clierr = writeFilesResponse(outDir, filesResp)
	if clierr != nil {
		return clierror.WrapE(clierr, clierror.New("failed to write files to output directory"))
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Files saved to %s\n", outDir)
	return nil
}

func writeFilesResponse(outputDir string, filesResp call.FilesListResponse) clierror.Error {
	for _, file := range filesResp.Files {
		filePath := filepath.Join(outputDir, file.Name)
		err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		if err != nil {
			return clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to create directory for file %s", filePath)))
		}

		data, err := base64.StdEncoding.DecodeString(file.Data)
		if err != nil {
			return clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to decode data for file %s", filePath)))
		}

		err = os.WriteFile(filePath, data, os.ModePerm)
		if err != nil {
			return clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to write file %s", filePath)))
		}
	}

	return nil
}
