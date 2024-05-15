package clierror

import (
	"fmt"
)

type Error struct {
	Message string
	Details string
	Hints   []string
}

// Error returns the error string, compatible with the error interface
func (e Error) Error() string {
	output := fmt.Sprintf("Error:\n  %s\n\n", e.Message)
	if e.Details != "" {
		output += fmt.Sprintf("Error Details:\n  %s\n\n", e.Details)
	}
	if len(e.Hints) > 0 {
		output += "Hints:\n"
		for _, hint := range e.Hints {
			output += fmt.Sprintf("  - %s\n", hint)
		}
	}
	return output
}

// Wrap adds a new message and hints to the error
func (inside *Error) wrap(outside *Error) *Error {
	newError := &Error{
		Message: inside.Message,
		Details: inside.Details,
		Hints:   inside.Hints,
	}

	if outside.Message != "" {
		newError.Message = outside.Message
	}

	if outside.Hints != nil {
		newError.Hints = append(outside.Hints, newError.Hints...)
	}

	if outside.Message != "" {
		newError.Details = wrapDetails(inside.Message, newError.Details)
	}

	if outside.Details != "" {
		newError.Details = wrapDetails(outside.Details, newError.Details)
	}

	return newError
}

func wrapDetails(outside, inside string) string {
	if outside == "" {
		return inside
	}
	if inside == "" {
		return outside
	}
	return fmt.Sprintf("%s: %s", outside, inside)
}
