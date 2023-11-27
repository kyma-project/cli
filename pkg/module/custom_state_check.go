package module

import (
	"slices"

	"github.com/kyma-project/lifecycle-manager/api/shared"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"github.com/pkg/errors"
)

var validStates = []shared.State{
	shared.StateReady,
	shared.StateProcessing,
	shared.StateError,
	shared.StateDeleting,
	shared.StateWarning,
}

var ErrCustomStateCheckValidation = errors.New("custom state check validation failed")

func IsValidMappedState(s string) bool {
	for _, state := range validStates {
		if string(state) == s {
			return true
		}
	}
	return false
}

func ContainsAllRequiredStates(checks []v1beta2.CustomStateCheck) bool {
	containsError := slices.ContainsFunc(checks, func(csc v1beta2.CustomStateCheck) bool {
		return csc.MappedState == shared.StateError
	})

	containsReady := slices.ContainsFunc(checks, func(csc v1beta2.CustomStateCheck) bool {
		return csc.MappedState == shared.StateReady
	})

	return containsError && containsReady
}
