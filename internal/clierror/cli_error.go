package clierror

import (
	"fmt"
	"os"
)

type Error struct {
	Message string
	Details string
	Hints   []string
}

// Wrap adds a new message and hints to the error
func (inside *Error) wrap(outside *Error) *Error {
	newError := &Error{
		Message: outside.Message,
		Details: inside.Details,
		Hints:   append(outside.Hints, inside.Hints...),
	}

	if inside.Message != "" {
		newError.Details = wrapDetails(inside.Message, newError.Details)
	}

	if outside.Details != "" {
		newError.Details = wrapDetails(outside.Details, newError.Details)
	}

	return newError
}

func wrapDetails(left, right string) string {
	if left == "" {
		return right
	}
	if right == "" {
		return left
	}
	return fmt.Sprintf("%s: %s", left, right)
}

func Wrap(inside error, outside *Error) error {
	if err, ok := inside.(*Error); ok {
		return err.wrap(outside)
	} else {
		return &Error{
			Message: outside.Message,
			Details: wrapDetails(outside.Details, inside.Error()),
			Hints:   outside.Hints,
		}
	}
}

func (e *Error) Print() {
	fmt.Printf("%s\n", e.Error())
}

func (e *Error) PrintStderr() {
	fmt.Fprintf(os.Stderr, "%s\n", e.Error())
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
