package cli

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// RunCmd executes a command with given arguments
func RunCmd(c string, args ...string) (string, error) {
	cmd := exec.Command(c, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Executing command '%s %s' failed with output '%s' and error message '%s'", c, args, out, err)
	}
	return strings.Replace(string(out), "'", "", -1), nil
}

//go:generate mockery --name Executor
type Executor interface {
	RunCmd(ctx context.Context, name string, args ...string) (string, error)
}

type executor struct{}

func NewExecutor() Executor {
	return &executor{}
}

func (e *executor) RunCmd(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	outBytes, err := cmd.CombinedOutput()
	out := string(outBytes)
	return out, err
}
