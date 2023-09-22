package module

import (
	"slices"

	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"github.com/pkg/errors"
)

var validStates = []v1beta2.State{
	v1beta2.StateReady,
	v1beta2.StateProcessing,
	v1beta2.StateError,
	v1beta2.StateDeleting,
	v1beta2.StateWarning,
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
		return csc.MappedState == v1beta2.StateError
	})

	containsReady := slices.ContainsFunc(checks, func(csc v1beta2.CustomStateCheck) bool {
		return csc.MappedState == v1beta2.StateReady
	})

	return containsError && containsReady
}
