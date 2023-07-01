package cli

import (
	"github.com/avast/retry-go"
	"github.com/kyma-project/cli/internal/cli/alpha/module"
	"github.com/pkg/errors"
)

var ErrorCodeMap = map[error]int{
	module.ErrKymaInWarningState: 2,
}

func GetExitCode(err error) int {
	switch err := err.(type) {
	default:
		return handleSingleError(err)
	case retry.Error:
		return handleListOfErrors(err)
	}
}

func handleSingleError(err error) int {
	if errorCode, found := mapNestedErrorToCode(err); found {
		return errorCode
	}
	return 1
}

func handleListOfErrors(errorList retry.Error) int {
	for _, err := range errorList {
		if errorCode, found := mapNestedErrorToCode(err); found {
			return errorCode
		}
	}
	return 1
}

func mapNestedErrorToCode(err error) (int, bool) {
	for {
		if errorCode, ok := ErrorCodeMap[err]; ok {
			return errorCode, true
		}
		err = errors.Unwrap(err)
		if err == nil {
			return -1, false
		}
	}
}
