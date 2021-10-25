package k3d

import (
	"context"
	"os/exec"
)

//go:generate mockery --name CmdRunner --filename cmd_runner.go
type CmdRunner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

//go:generate mockery --name PathLooker --filename path_looker.go
type PathLooker interface {
	Look(file string) (string, error)
}

type cmdRunner struct{}
type pathLooker struct{}

func NewCmdRunner() CmdRunner {
	return &cmdRunner{}
}

func NewPathLooker() PathLooker {
	return &pathLooker{}
}

func (e *cmdRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	outBytes, err := cmd.CombinedOutput()
	out := string(outBytes)
	return out, err
}

func (e *pathLooker) Look(file string) (string, error) {
	return exec.LookPath(file)
}
