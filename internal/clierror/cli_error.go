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
func (e *Error) Wrap(message string, hints []string) *Error {
	newError := &Error{
		Message: message,
		Details: e.Message,
		Hints:   append(hints, e.Hints...),
	}

	if e.Details != "" {
		newError.Details = fmt.Sprintf("%s: %s", e.Message, e.Details)
	}

	return newError
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
