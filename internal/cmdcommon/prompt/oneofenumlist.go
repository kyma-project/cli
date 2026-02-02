package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/kyma-project/cli.v3/internal/out"
)

type EnumValueWithDescription struct {
	value       string
	description string
}

func NewEnumValWithDesc(val, desc string) *EnumValueWithDescription {
	return &EnumValueWithDescription{
		value:       val,
		description: desc,
	}
}

type OneOfEnumList struct {
	reader                 io.Reader
	printer                *out.Printer
	message                string
	promptText             string
	valuesWithDescriptions []EnumValueWithDescription
}

func NewOneOfEnumList(message, promptText string, valuesWithDescriptions []EnumValueWithDescription) *OneOfEnumList {
	return &OneOfEnumList{
		reader:                 os.Stdin,
		printer:                out.Default,
		message:                message,
		promptText:             promptText,
		valuesWithDescriptions: valuesWithDescriptions,
	}
}

func (l *OneOfEnumList) Prompt() (string, error) {
	l.printer.Msgf("%s\n%s\n\n%s", l.message, l.valuesListString(), l.promptText)
	scanner := bufio.NewScanner(l.reader)
	scanner.Scan()
	err := scanner.Err()
	userInput := scanner.Text()
	l.printer.Msg("\n")

	if err != nil {
		return "", err
	}

	validatedUserInput, err := l.validateUserInput(userInput)
	if err != nil {
		return "", err
	}

	return l.valuesWithDescriptions[validatedUserInput-1].value, nil
}

func (l *OneOfEnumList) valuesListString() string {
	var rows []string
	for index, row := range l.valuesWithDescriptions {
		rows = append(rows, fmt.Sprintf("%d. %s (%s)", index+1, row.value, row.description))
	}

	return strings.Join(rows, "\n")
}

func (l *OneOfEnumList) validateUserInput(userInput string) (int, error) {
	if strings.TrimSpace(userInput) == "" {
		return -1, fmt.Errorf("no value was selected")
	}

	parsedInput, err := strconv.Atoi(strings.TrimSpace(userInput))
	if err != nil {
		return -1, fmt.Errorf("provided value is not a number")
	}

	if !l.isUserInputValidOptionsNumber(parsedInput) {
		return -1, fmt.Errorf("invalid option selected")
	}

	return parsedInput, nil
}

func (l *OneOfEnumList) isUserInputValidOptionsNumber(parsedInput int) bool {
	lowerRange := 1
	higherRange := len(l.valuesWithDescriptions)

	return lowerRange <= parsedInput && parsedInput <= higherRange
}
