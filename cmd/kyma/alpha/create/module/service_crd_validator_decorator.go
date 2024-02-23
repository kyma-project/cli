package module

import (
	"context"
	"errors"

	"github.com/kyma-project/cli/pkg/module"
)

type (
	cmdRunner interface {
		NewStep(string)
		CurrentStepSuccess(string)
	}

	crdValidatorDecorator struct {
		validator crdValidator
		cmd       cmdRunner
	}
)

func (v *crdValidatorDecorator) ValidateCRD(
	ctx context.Context,
	modDef *module.Definition,
	kubebuilderProject bool,
) (validator, error) {
	v.cmd.NewStep("Validating Default CR")

	res, err := v.validator.ValidateCRD(ctx, modDef, kubebuilderProject)
	if errors.Is(err, module.ErrEmptyCR) {
		v.cmd.CurrentStepSuccess("Default CR validation skipped - no default CR")
	}

	v.cmd.CurrentStepSuccess("Default CR validation succeeded")

	return res, err
}
