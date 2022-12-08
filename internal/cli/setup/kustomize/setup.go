package setup

import (
	"github.com/kyma-project/cli/internal/kustomize"
	"github.com/kyma-project/cli/pkg/step"
)

const (
	DefaultNewStepMsg     = "Setting up kustomize..."
	DefaultStepSuccessMsg = "Kustomize ready"
)

// ExtendCmd is used to extend the CLI command with the SetupKustomize() function
type ExtendCmd struct {
}

func (kc ExtendCmd) SetupKustomize(s step.Step) error {
	if err := kustomize.Setup(s, true); err != nil {
		return err
	}
	return nil
}
