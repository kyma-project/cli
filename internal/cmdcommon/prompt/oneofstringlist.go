package prompt

import (
	"bufio"
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
	fmt.Fprintf(l.writer, "%s\n%s\n\n%s", l.message, l.valuesListString(), l.promptText)
	scanner := bufio.NewScanner(l.reader)
	scanner.Scan()
	err := scanner.Err()
	userInput := scanner.Text()

	if err != nil {
		return "", err
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
	if strings.TrimSpace(userInput) == "" {
		return "", fmt.Errorf("no value was selected")
	}
	if !l.isUserInputPresentInValuesList(userInput) {
		return "", fmt.Errorf("provided value is not present on the list: %v", userInput)
	}

	return userInput, nil
}

func (l *OneOfStringList) isUserInputPresentInValuesList(userInput string) bool {
	return slices.Contains(l.values, userInput)
}
