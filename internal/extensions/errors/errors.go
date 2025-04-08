package errors

import (
	"errors"
	"fmt"
	"strings"
)

var (
	New  = errors.New
	Newf = fmt.Errorf
)

type ErrorList struct {
	errors []error
}

func NewList(errs ...error) error {
	errList := []error{}
	for i := range errs {
		if errs[i] != nil {
			errList = append(errList, errs[i])
		}
	}

	if len(errList) > 0 {
		return &ErrorList{
			errors: errList,
		}
	}

	return nil
}

func (l *ErrorList) Error() string {
	return JoinWithSeparator("\n", l.errors...).Error()
}

// use to wrap error with another, more detailed
func Wrap(inner error, message string) error {
	if isErrorList(inner) {
		return JoinWithSeparator(":", New(message), addPrefix(inner))
	}

	return JoinWithSeparator(": ", New(message), inner)
}

// use to wrap error with another, more detailed with format
func Wrapf(inner error, format string, args ...any) error {
	return Wrap(inner, fmt.Sprintf(format, args...))
}

// use to join errors with given separator
func JoinWithSeparator(separator string, errs ...error) error {
	errMsgs := []string{}
	for i := range errs {
		if errs[i] != nil {
			errMsgs = append(errMsgs, errs[i].Error())
		}
	}

	if len(errMsgs) == 0 {
		// no valid error found
		return nil
	}

	return New(strings.Join(errMsgs, separator))
}

func addPrefix(err error) error {
	return Newf("\n  %s", strings.ReplaceAll(err.Error(), "\n", "\n  "))
}

func isErrorList(err error) bool {
	if err == nil {
		return false
	}

	switch err.(type) {
	case *ErrorList:
		return true
	}

	return false
}
