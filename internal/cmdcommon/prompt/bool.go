package prompt

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Bool struct {
	reader       io.Reader
	writer       io.Writer
	message      string
	defaultValue bool
}

func NewBool(message string, defaultValue bool) *Bool {
	return &Bool{
		reader:       os.Stdin,
		writer:       os.Stdout,
		message:      message,
		defaultValue: defaultValue,
	}
}

func (b *Bool) Prompt() (bool, error) {
	var userInput string
	fmt.Fprintf(b.writer, "%s %s: ", b.message, b.defaultValueDisplay())
	_, err := fmt.Fscan(b.reader, &userInput)

	// If the user just presses Enter, Fscan returns the EOF error
	if err != nil && err == io.EOF {
		// Treat as empty input, use default value
		return b.defaultValue, nil
	}

	parsedUserInput, err := b.validateUserInput(userInput)
	if err != nil {
		return false, err
	}

	return parsedUserInput, nil
}

func (b *Bool) defaultValueDisplay() string {
	if b.defaultValue {
		return "[Y/n]"
	}
	return "[y/N]"
}

func (b *Bool) validateUserInput(userInput string) (bool, error) {
	switch strings.TrimSpace(strings.ToLower(userInput)) {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	case "":
		return b.defaultValue, nil
	default:
		return false, fmt.Errorf("invalid input, please enter 'y' or 'n'")
	}
}
