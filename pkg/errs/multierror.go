package errs

import (
	"errors"
	"fmt"
	"strings"
)

type Multierror []error

// As attempts to find the first error in the error list that matches the type
// of the value that target points to.
//
// This function allows errors.As to traverse the values stored on the
// multierr error.
func (merr Multierror) As(target interface{}) bool {
	for _, err := range merr {
		if errors.As(err, target) {
			return true
		}
	}
	return false
}

// Is attempts to match the provided error against errors in the error list.
//
// This function allows errors.Is to traverse the values stored on the
// multierr error.
func (merr Multierror) Is(target error) bool {
	for _, err := range merr {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

func (merr Multierror) Error() string {
	buf := strings.Builder{}
	for _, e := range merr {
		buf.WriteString(fmt.Sprintf("%s\n", e))
	}
	return buf.String()
}

// MergeErrs checks all errors passed and merges all non-nil errors into a MultiError
// if no errors are passed or all are nil, this function returns nil
func MergeErrors(errSlice ...error) error {
	res := Multierror{}

	for _, e := range errSlice {
		if e != nil {
			res = append(res, e)
		}
	}

	if len(res) > 0 {
		return res
	}

	return nil
}
