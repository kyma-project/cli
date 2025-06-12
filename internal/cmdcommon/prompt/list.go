package prompt

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

type OneOfList[T comparable] struct {
	reader         io.Reader
	writer         io.Writer
	message        string
	values         []T
	parseToStrFunc func(string) (T, error)
}

func NewOneOfList[T comparable](reader io.Reader, writer io.Writer, message string, values []T, parseToStrFunc func(string) (T, error)) *OneOfList[T] {
	return &OneOfList[T]{
		reader:         reader,
		writer:         writer,
		message:        message,
		values:         values,
		parseToStrFunc: parseToStrFunc,
	}
}

func NewOneOfStringList(message string, values []string) *OneOfList[string] {
	return &OneOfList[string]{
		reader:  os.Stdin,
		writer:  os.Stdout,
		message: message,
		values:  values,
		parseToStrFunc: func(s string) (string, error) {
			return s, nil
		},
	}
}

func (l *OneOfList[T]) Prompt() (T, error) {
	var userInput string
	fmt.Fprintf(l.writer, "%s\n%s\n\nType your choice: ", l.message, l.valuesListString())
	_, err := fmt.Fscan(l.reader, &userInput)
	// If the user just presses Enter, Fscan returns the EOF error
	if err != nil && err == io.EOF {
		return l.zeroValue(), fmt.Errorf("no value was selected")
	}

	parsedInput, err := l.parseToStrFunc(userInput)
	if err != nil {
		return l.zeroValue(), err
	}

	validatedUserInput, err := l.validateUserInput(parsedInput)
	if err != nil {
		return l.zeroValue(), err
	}

	return validatedUserInput, nil
}

func (l *OneOfList[T]) valuesListString() string {
	var rows []string

	for _, row := range l.values {
		rows = append(rows, fmt.Sprintf(" - %v", row))
	}

	return strings.Join(rows, "\n")
}

func (l *OneOfList[T]) validateUserInput(userInput T) (T, error) {
	if l.isUserInputPresentInValuesList(userInput) {
		return userInput, nil
	}

	return l.zeroValue(), fmt.Errorf("provided value is not present on the list: %v", userInput)
}

func (l *OneOfList[T]) isUserInputPresentInValuesList(userInput T) bool {
	return slices.Contains(l.values, userInput)
}

func (l *OneOfList[T]) zeroValue() T {
	var zero T
	return zero
}
