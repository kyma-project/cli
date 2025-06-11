package prompt

import (
	"fmt"
	"slices"
	"strings"
)

type List struct {
	message string
	values  []string
}

func NewList(message string, values []string) *List {
	return &List{
		message: message,
		values:  values,
	}
}

func (l *List) Prompt() (string, error) {
	var userInput string

	fmt.Printf("%s\n%s\n\nType your choice: ", l.message, l.valuesListString())
	_, err := fmt.Scanln(&userInput)
	if err != nil {
		return "", err
	}

	validatedUserInput, err := l.validateUserInput(userInput)
	if err != nil {
		return "", err
	}

	return validatedUserInput, nil
}

func (l *List) valuesListString() string {
	var rows []string

	for _, row := range l.values {
		rows = append(rows, fmt.Sprintf(" - %s", row))
	}

	return strings.Join(rows, "\n")
}

func (l *List) validateUserInput(userInput string) (string, error) {
	if l.isUserInputPresentInValuesList(userInput) {
		return userInput, nil
	}

	return "", fmt.Errorf("provided value is not present on the list: %s", userInput)
}

func (l *List) isUserInputPresentInValuesList(userInput string) bool {
	return slices.Contains(l.values, userInput)
}
