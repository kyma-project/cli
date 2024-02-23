package module

import (
	"context"
	"errors"

	"github.com/kyma-project/cli/pkg/module"
)

type (
	errorLogger interface {
		Error(args ...interface{})
	}

	crdValidator struct {
		logger errorLogger
	}
)

func (v crdValidator) ValidateCRD(
	ctx context.Context,
	modDef *module.Definition,
	kubebuilderProject bool,
) (validator, error) {
	var crValidator validator
	if kubebuilderProject {
		crValidator = module.NewDefaultCRValidator(modDef.DefaultCR, modDef.Source)
	} else {
		crValidator = module.NewSingleManifestFileCRValidator(modDef.DefaultCR, modDef.SingleManifestPath)
	}

	if err := crValidator.Run(ctx, v.logger); err != nil {
		if errors.Is(err, module.ErrEmptyCR) {
			return crValidator, nil
		}
		return crValidator, err
	}

	return crValidator, nil
}
