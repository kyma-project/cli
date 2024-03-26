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
func (f *Error) Wrap(message string, hints []string) {
	if f.Details != "" {
		f.Details = fmt.Sprintf("%s: %s", f.Message, f.Details)
	} else {
		f.Details = f.Message
	}
	f.Message = message
	f.Hints = append(hints, f.Hints...)
}

func (f *Error) Print() {
	fmt.Printf("%s\n", f.Error())
}

func (f *Error) PrintStderr() {
	fmt.Fprintf(os.Stderr, "%s\n", f.Error())
}

// Error returns the error string, compatible with the error interface
func (f Error) Error() string {
	output := fmt.Sprintf("Error:\n  %s\n\n", f.Message)
	if f.Details != "" {
		output += fmt.Sprintf("Error Details:\n  %s\n\n", f.Details)
	}
	if len(f.Hints) > 0 {
		output += "Hints:\n"
		for _, hint := range f.Hints {
			output += fmt.Sprintf("  - %s\n", hint)
		}
	}
	return output
}
