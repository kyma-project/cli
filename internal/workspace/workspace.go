package workspace

import (
	"github.com/kyma-project/cli/internal/resources/types"
	"github.com/pkg/errors"
	"os"
	"path"
)

type FileName string

type workspace []file

func (ws workspace) build(cfg Cfg, dirPath string) error {
	workspaceFiles := append(ws, cfg)
	for _, fileTemplate := range workspaceFiles {
		if err := write(dirPath, fileTemplate, cfg); err != nil {
			return err
		}
	}
	return nil
}

func write(destinationDirPath string, fileTemplate file, cfg Cfg) error {
	outFilePath := path.Join(destinationDirPath, fileTemplate.fileName())

	file, err := os.Create(outFilePath)
	if err != nil {
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			return
		}
	}()

	err = fileTemplate.write(file, cfg)
	if err != nil {
		return err
	}

	return nil
}

var errUnsupportedRuntime = errors.New("unsupported runtime")

func Initialize(cfg Cfg, dirPath string) error {
	ws, err := fromRuntime(cfg.Runtime)
	if err != nil {
		return err
	}
	return ws.build(cfg, dirPath)
}

func fromRuntime(runtime types.Runtime) (workspace, error) {
	switch runtime {
	case types.Nodejs12, types.Nodejs10:
		return workspaceNodeJs, nil
	case types.Python38:
		return workspacePython, nil
	default:
		return nil, errUnsupportedRuntime
	}
}
