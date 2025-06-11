package prompt

import (
	"fmt"
	"strings"
)

type Bool struct {
	message      string
	defaultValue bool
}

func NewBool(message string, defaultValue bool) *Bool {
	return &Bool{
		message:      message,
		defaultValue: defaultValue,
	}
}

func (b *Bool) Prompt() (bool, error) {
	var userInput string
	fmt.Printf("%s %s: ", b.message, b.defaultValueDisplay())
	_, err := fmt.Scanln(&userInput)

	// If the user just presses Enter, Scanln returns an error, but userInput remains empty.
	if err != nil && userInput == "" {
		// Treat as empty input, use default value
		return b.defaultValue, nil
	}
	if err != nil {
		return false, err
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
	case "y":
		return true, nil
	case "n":
		return false, nil
	case "":
		return b.defaultValue, nil
	default:
		return false, fmt.Errorf("invalid input, please enter 'y' or 'n'")
	}
}
