package vscode

import (
	"path"

	serverlessw "github.com/kyma-project/cli/pkg/serverless"
	"github.com/kyma-project/hydroform/function/pkg/workspace"
)

type Configuration map[workspace.File]interface{}

func (s Configuration) Build(dirPath string) error {
	return s.build(dirPath, workspace.DefaultWriterProvider)
}

func (s Configuration) build(dirPath string, writerProvider workspace.WriterProvider) error {
	for file, cfg := range s {
		if err := writerProvider.Write(dirPath, file, cfg); err != nil {
			return err
		}
	}

	return nil
}

var (
	settingsTpl = `{
  "yaml.schemas": {
    "{{ .SchemaPath }}": "{{ .ConfigPath }}"
  }
}`
	settings = workspace.NewTemplatedFile(settingsTpl, "settings.json")

	schema serverlessw.Schema

	settingsCfg = struct {
		SchemaPath string
		ConfigPath string
	}{
		SchemaPath: path.Join(".", ".vscode", schema.FileName()),
		ConfigPath: "/config.yaml",
	}

	Workspace = Configuration(map[workspace.File]interface{}{
		settings: settingsCfg,
		schema:   nil,
	})
)
