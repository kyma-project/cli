package clierror

import (
	"fmt"
)

type Error interface {
	String() string
}

type clierror struct {
	message string
	details string
	hints   []string
}

// New creates a new error with the given modifiers
func New(message string, hints ...string) Error {
	return new(message, hints...)
}

func new(message string, hints ...string) *clierror {
	return &clierror{
		message: message,
		hints:   hints,
	}
}

// Error returns the error string, compatible with the error interface
func (e *clierror) String() string {
	output := fmt.Sprintf("Error:\n  %s\n\n", e.message)
	if e.details != "" {
		output += fmt.Sprintf("Error Details:\n  %s\n\n", e.details)
	}
	if len(e.hints) > 0 {
		output += "Hints:\n"
		for _, hint := range e.hints {
			output += fmt.Sprintf("  - %s\n", hint)
		}
	}
	return output
}

// Wrap adds a new message and hints to the error
func (inside *clierror) wrap(outside *clierror) *clierror {
	newError := &clierror{
		message: inside.message,
		details: inside.details,
		hints:   inside.hints,
	}

	if outside.message != "" {
		newError.message = outside.message
	}

	if outside.hints != nil {
		newError.hints = append(outside.hints, newError.hints...)
	}

	if outside.message != "" {
		newError.details = wrapDetails(inside.message, newError.details)
	}

	if outside.details != "" {
		newError.details = wrapDetails(outside.details, newError.details)
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
