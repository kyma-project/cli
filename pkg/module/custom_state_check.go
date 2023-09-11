package module

import (
	"fmt"
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

func ValidateCustomStateCheck(paths, values, states []string) error {
	if (len(paths)*len(values)*len(states) == 0) && (len(paths)+len(values)+len(states) != 0) {
		return fmt.Errorf(
			"%w, all 3 arguments must be provided",
			ErrCustomStateCheckValidation,
		)
	}

	if len(paths) != len(values) || len(values) != len(states) {
		return fmt.Errorf(
			"%w, same number of paths, values, and states must be provided",
			ErrCustomStateCheckValidation,
		)
	}

	for _, state := range states {
		if !IsValidMappedState(state) {
			return fmt.Errorf(
				"%w, the provided state [%q] is a not a valid state",
				ErrCustomStateCheckValidation, state,
			)
		}
	}

	return nil
}

func GenerateChecks(paths, values, states []string) []v1beta2.CustomStateCheck {
	var checks []v1beta2.CustomStateCheck

	for i := range paths {
		checks = append(checks, v1beta2.CustomStateCheck{
			JSONPath:    paths[i],
			Value:       values[i],
			MappedState: v1beta2.State(states[i]),
		})
	}

	return checks
}
