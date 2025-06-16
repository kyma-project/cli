package prompt

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

type OneOfStringList struct {
	reader     io.Reader
	writer     io.Writer
	message    string
	promptText string
	values     []string
}

func NewOneOfStringList(message, promptText string, values []string) *OneOfStringList {
	return &OneOfStringList{
		reader:     os.Stdin,
		writer:     os.Stdout,
		message:    message,
		promptText: promptText,
		values:     values,
	}
}

func NewCustomOneOfStringList(reader io.Reader, writer io.Writer, message, promptText string, values []string) *OneOfStringList {
	return &OneOfStringList{
		reader:     reader,
		writer:     writer,
		message:    message,
		promptText: promptText,
		values:     values,
	}
}

func (l *OneOfStringList) Prompt() (string, error) {
	var userInput string
	fmt.Fprintf(l.writer, "%s\n%s\n\n%s", l.message, l.valuesListString(), l.promptText)
	_, err := fmt.Fscan(l.reader, &userInput)
	if err != nil && err == io.EOF {
		return "", fmt.Errorf("no value was selected")
	}

	validatedUserInput, err := l.validateUserInput(userInput)
	if err != nil {
		return "", err
	}

	return validatedUserInput, nil
}

func (l *OneOfStringList) valuesListString() string {
	var rows []string
	for _, row := range l.values {
		rows = append(rows, fmt.Sprintf(" - %v", row))
	}
	return strings.Join(rows, "\n")
}

func (l *OneOfStringList) validateUserInput(userInput string) (string, error) {
	if l.isUserInputPresentInValuesList(userInput) {
		return userInput, nil
	}
	return "", fmt.Errorf("provided value is not present on the list: %v", userInput)
}

func (l *OneOfStringList) isUserInputPresentInValuesList(userInput string) bool {
	return slices.Contains(l.values, userInput)
}
