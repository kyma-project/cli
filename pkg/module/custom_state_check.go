package module

import (
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
